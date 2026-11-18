package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rjcuff/woolly/internal/gitinfo"
	"github.com/rjcuff/woolly/internal/ground"
	"github.com/rjcuff/woolly/internal/handoff"
	"github.com/rjcuff/woolly/internal/memory"
	"github.com/rjcuff/woolly/internal/share"
)

func init() {
	groundRefs = func(dir string, refs []memory.Ref) []memory.Ref { return ground.Open(dir).Resolve(refs) }
}

// exitError ends woolly with a status code, after the command has printed
// everything it means to.
type exitError int

func (e exitError) Error() string { return fmt.Sprintf("exit status %d", int(e)) }

type checkOptions struct {
	backfill bool
	// base is the commit to compare against for facts that don't say which
	// commit they were true at, like facts read from woolly.md on a fresh
	// machine. In a pull request it's the branch being merged into.
	base     string
	deadline time.Time // stop checking after this; zero means no limit
}

type checkReport struct {
	Project    string          `json:"project"`
	Head       string          `json:"head"`
	Checked    int             `json:"checked"`
	NewStale   int             `json:"new_stale"`
	Waiting    int             `json:"waiting"` // facts already stale before this check
	Backfilled int             `json:"backfilled"`
	Caught     int             `json:"caught"` // stale facts caught in this project so far
	Results    []ground.Result `json:"results"`
}

// checkRepo checks the project's facts against the code in dir, marks the
// ones that went stale and saves sources that moved.
func (a *app) checkRepo(dir, project string, o checkOptions) (checkReport, error) {
	rep := checkReport{Project: project}
	g := ground.Open(dir)
	rep.Head = g.Head()
	head := gitinfo.Head(dir)
	mems, err := a.store.List(true)
	if err != nil {
		return rep, err
	}
	for _, m := range mems {
		if m.Project != project || m.Key == handoff.Key {
			continue
		}
		if m.Status == memory.StatusStale {
			rep.Waiting++
			continue
		}
		if m.Status != memory.StatusActive {
			continue
		}
		if !o.deadline.IsZero() && time.Now().After(o.deadline) {
			break
		}
		if len(m.Sources) == 0 {
			if !o.backfill {
				continue
			}
			refs := g.Suggest(m.Content)
			if len(refs) == 0 {
				continue
			}
			saved, err := a.store.SetSources(m.ID, refs, head)
			if err != nil {
				continue
			}
			rep.Backfilled++
			m = saved
		}
		if m.Commit == "" && o.base != "" {
			m.Commit = o.base
		}
		res := g.Check(m)
		rep.Checked++
		if res.Update {
			if _, err := a.store.SetSources(m.ID, res.Refs, head); err != nil {
				return rep, err
			}
		} else if m.Commit == "" && head != "" {
			_, _ = a.store.SetSources(m.ID, m.Sources, head)
		}
		if res.Verdict == ground.Stale {
			if _, err := a.store.MarkStale(m.ID, memory.StaleInput{Reason: res.Reason, Commit: res.Commit}); err != nil {
				return rep, err
			}
			a.noteCaught(project, res)
			rep.NewStale++
		}
		rep.Results = append(rep.Results, res)
	}
	ground.Sort(rep.Results)
	rep.Caught = len(a.caught(project))
	return rep, nil
}

// cmdCheck checks this repo's facts against its code.
func cmdCheck(args []string) error {
	fs := newFlags("check")
	asJSON := fs.Bool("json", false, "print the results as JSON")
	asMarkdown := fs.Bool("markdown", false, "print a report for a pull request comment")
	quiet := fs.Bool("quiet", false, "print only when something went stale")
	noBackfill := fs.Bool("no-backfill", false, "don't add sources to facts that have none")
	hook := fs.Bool("hook", false, "run from the git post-commit hook: never fail, print nothing")
	dir := fs.String("dir", "", "the repo (default: the current folder)")
	base := fs.String("base", "", "compare against this commit for facts that don't record one, like a pull request's base branch")
	if _, err := parse(fs, args); err != nil {
		return err
	}
	cwd := *dir
	if cwd == "" {
		cwd, _ = os.Getwd()
	}
	r := share.Find(cwd)
	if r == nil || gitinfo.Head(r.Dir) == "" {
		if *hook {
			return nil
		}
		return errors.New("run this inside a git repo with at least one commit")
	}
	a, err := open()
	if err != nil {
		if *hook {
			return nil
		}
		return err
	}
	if *hook {
		unlock, ok := a.lock("check")
		if !ok {
			return nil // another check is already running
		}
		defer unlock()
	}
	a.syncRepo(r.Dir)
	rep, err := a.checkRepo(r.Dir, r.Project, checkOptions{backfill: !*noBackfill, base: strings.TrimSpace(*base)})
	if err != nil {
		if *hook {
			return nil
		}
		return err
	}
	if rep.NewStale > 0 || rep.Backfilled > 0 {
		a.syncRepo(r.Dir)
	}
	var ids []string
	for _, res := range rep.Results {
		ids = append(ids, res.ID)
	}
	a.record("check", ids, rep.Checked)
	if *hook {
		return nil
	}
	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(rep); err != nil {
			return err
		}
	} else if *asMarkdown {
		fmt.Print(markdownCheck(rep))
	} else if !*quiet || rep.NewStale > 0 {
		printCheck(rep)
	}
	if rep.NewStale > 0 {
		return exitError(1)
	}
	return nil
}

func printCheck(rep checkReport) {
	if rep.Checked == 0 {
		fmt.Printf("No facts in %s point at code yet.\n", rep.Project)
		fmt.Println("Agents add sources when they save project facts. You can too, with woolly remember --source-ref main/pins.h:12.")
		if rep.Waiting > 0 {
			fmt.Printf("%s waiting for review. Run woolly review.\n", plural(rep.Waiting, "stale fact is", "stale facts are"))
		}
		return
	}
	fmt.Printf("Checked %s in %s against %s.\n", plural(rep.Checked, "fact", "facts"), rep.Project, rep.Head)
	for _, res := range rep.Results {
		switch res.Verdict {
		case ground.Stale, ground.Moved, ground.Changed:
		default:
			continue
		}
		fmt.Printf("  %-7s%s\n", string(res.Verdict), clip(res.Content, 72))
		why := res.Reason
		if res.Verdict == ground.Stale {
			if res.Commit != "" {
				why += " in " + res.Commit
			} else {
				why += " (not committed yet)"
			}
		}
		fmt.Printf("         %s\n", why)
	}
	if rep.Backfilled > 0 {
		fmt.Printf("Added sources to %s from the files and names they mention.\n", plural(rep.Backfilled, "fact", "facts"))
	}
	switch {
	case rep.NewStale > 0:
		fmt.Printf("%s stale. Run woolly review to confirm, fix or retire it.\n", plural(rep.NewStale, "fact went", "facts went"))
		if rep.Caught > 0 {
			fmt.Printf("Caught it. That's %s Woolly has kept away from your agents in %s.\n", plural(rep.Caught, "stale fact", "stale facts"), rep.Project)
		}
	case rep.Waiting > 0:
		fmt.Printf("Nothing new went stale. %s waiting for review. Run woolly review.\n", plural(rep.Waiting, "stale fact is", "stale facts are"))
	default:
		fmt.Println("Nothing went stale.")
	}
}

// CommentMarker starts every pull request comment Woolly writes, so the
// GitHub Action can find its own comment and update it.
const CommentMarker = "<!-- woolly-check -->"

// markdownCheck is the check as a pull request comment: which facts this
// change made stale and why, and what moved without breaking anything.
func markdownCheck(rep checkReport) string {
	var b strings.Builder
	b.WriteString(CommentMarker + "\n")
	cell := func(s string) string { return strings.ReplaceAll(strings.ReplaceAll(s, "|", "\\|"), "\n", " ") }
	var stale, moved []ground.Result
	for _, res := range rep.Results {
		switch res.Verdict {
		case ground.Stale:
			stale = append(stale, res)
		case ground.Moved, ground.Changed:
			moved = append(moved, res)
		}
	}
	switch {
	case rep.Checked == 0:
		b.WriteString("**Woolly** found no project facts that point at code yet, so there was nothing to check.\n")
		return b.String()
	case len(stale) == 0:
		fmt.Fprintf(&b, "**Woolly** checked %s against this change. Nothing went stale.\n", plural(rep.Checked, "project fact", "project facts"))
	default:
		fmt.Fprintf(&b, "**Woolly** checked %s against this change. **%s stale.**\n\n", plural(rep.Checked, "project fact", "project facts"), plural(len(stale), "fact went", "facts went"))
		b.WriteString("Agents won't be told these anymore until someone reviews them.\n\n")
		b.WriteString("| Fact | What changed |\n| --- | --- |\n")
		for _, res := range stale {
			why := res.Reason
			if res.Commit != "" {
				why += " in " + res.Commit
			}
			fmt.Fprintf(&b, "| %s | %s |\n", cell(res.Content), cell(why))
		}
		b.WriteString("\nAfter this merges, run `woolly review` to confirm, fix or retire each one. Nothing is deleted.\n")
	}
	if len(moved) > 0 {
		fmt.Fprintf(&b, "\n<details><summary>%s moved but still hold</summary>\n\n", plural(len(moved), "fact's code", "facts' code"))
		for _, res := range moved {
			fmt.Fprintf(&b, "- %s (%s)\n", cell(res.Content), cell(res.Reason))
		}
		b.WriteString("\n</details>\n")
	}
	if rep.Waiting > 0 {
		fmt.Fprintf(&b, "\n%s waiting for review from before.\n", plural(rep.Waiting, "stale fact is", "stale facts are"))
	}
	return b.String()
}

// lock takes a lock file in Woolly's folder, so two checks from back to
// back commits don't run at once. A lock older than two minutes is stale.
func (a *app) lock(name string) (func(), bool) {
	p := filepath.Join(a.home, name+".lock")
	if st, err := os.Stat(p); err == nil && time.Since(st.ModTime()) > 2*time.Minute {
		os.Remove(p)
	}
	f, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, false
	}
	f.Close()
	return func() { os.Remove(p) }, true
}

// briefs remembers when each project's last session brief was given, so the
// next one can say what went stale since.
func (a *app) lastBrief(project string, now time.Time) time.Time {
	p := filepath.Join(a.home, "briefs.json")
	seen := map[string]time.Time{}
	if b, err := os.ReadFile(p); err == nil {
		_ = json.Unmarshal(b, &seen)
	}
	last := seen[project]
	seen[project] = now
	if b, err := json.Marshal(seen); err == nil {
		_ = os.WriteFile(p, b, 0o600)
	}
	return last
}

// staleLine is what a session brief says about stale facts. When some went
// stale since the last session it names them, so the agent knows what not to
// rely on and why.
func (a *app) staleLine(project string) string {
	stale, err := a.store.Stale(project)
	if err != nil || project == "" {
		return ""
	}
	since := a.lastBrief(project, time.Now().UTC())
	var fresh []memory.Memory
	for _, m := range stale {
		if m.StaleAt != nil && m.StaleAt.After(since) {
			fresh = append(fresh, m)
		}
	}
	switch {
	case len(fresh) > 0:
		var b strings.Builder
		fmt.Fprintf(&b, "Woolly kept %s out of this session. The code no longer backs %s up:\n", plural(len(fresh), "stale fact", "stale facts"), pick(len(fresh), "it", "them"))
		for i, m := range fresh {
			if i == 3 {
				fmt.Fprintf(&b, "- and %d more\n", len(fresh)-3)
				break
			}
			why := m.StaleReason
			if m.StaleCommit != "" {
				why += " in " + m.StaleCommit
			}
			fmt.Fprintf(&b, "- %q (%s)\n", clip(strings.TrimSpace(m.Content), 100), why)
		}
		fmt.Fprintf(&b, "Don't rely on %s until someone runs woolly review.", pick(len(fresh), "it", "them"))
		return b.String()
	case len(stale) > 0:
		return fmt.Sprintf("%s waiting for review. Run woolly review.", plural(len(stale), "stale fact is", "stale facts are"))
	}
	return ""
}

func pick(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
