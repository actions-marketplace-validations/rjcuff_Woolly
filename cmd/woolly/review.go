package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rjcuff/woolly/internal/gitinfo"
	"github.com/rjcuff/woolly/internal/ground"
	"github.com/rjcuff/woolly/internal/memory"
	"github.com/rjcuff/woolly/internal/share"
)

// cmdReview walks through stale facts so you can confirm, fix or retire
// each one.
func cmdReview(args []string) error {
	fs := newFlags("review")
	project := fs.String("project", "", "review this project (default: the current repo)")
	list := fs.Bool("list", false, "only list stale facts")
	confirm := fs.String("confirm", "", "mark this stale fact as still true")
	retire := fs.String("retire", "", "mark this stale fact as no longer true")
	old := fs.Bool("old", false, "list AI guesses nothing has used in 90 days")
	keep := fs.String("keep", "", "keep this old guess; it won't be asked about for another 90 days")
	if _, err := parse(fs, args); err != nil {
		return err
	}
	if *old || *keep != "" {
		return reviewOld(*keep)
	}
	cwd, _ := os.Getwd()
	repo := share.Find(cwd)
	p := memory.NormalizeProject(*project)
	if p == "" && repo != nil {
		p = repo.Project
	}
	if repo != nil && repo.Project != p {
		repo = nil // reviewing another project: no code to look at here
	}
	a, err := open()
	if err != nil {
		return err
	}
	if repo != nil {
		a.syncRepo(repo.Dir)
	}
	rv := reviewer{a: a, repo: repo}

	switch {
	case *confirm != "":
		m, err := a.store.Get(*confirm)
		if err != nil {
			return err
		}
		if err := rv.confirm(m); err != nil {
			return err
		}
		rv.done()
		return nil
	case *retire != "":
		m, err := a.store.Expire(*retire)
		if err != nil {
			return err
		}
		fmt.Println("Retired:", clip(m.Content, 80))
		rv.done()
		return nil
	}

	stale, err := a.store.Stale(p)
	if err != nil {
		return err
	}
	if len(stale) == 0 {
		where := "anywhere"
		if p != "" {
			where = "in " + p
		}
		fmt.Printf("No stale facts %s.\n", where)
		return nil
	}
	if *list || !interactive() {
		for _, m := range stale {
			printStale(m)
			fmt.Printf("         id %s\n", m.ID)
		}
		fmt.Println("\nRun woolly review in a terminal to go through them, or use --confirm <id> or --retire <id>.")
		return nil
	}

	rv.loop(bufio.NewReader(os.Stdin), stale)
	return nil
}

// loop asks about each stale fact in turn.
func (rv *reviewer) loop(in *bufio.Reader, stale []memory.Memory) {
	defer rv.done()
	for i, m := range stale {
		fmt.Printf("\n[%d/%d] ", i+1, len(stale))
		printStale(m)
		if len(m.Sources) > 0 {
			var where []string
			for _, r := range m.Sources {
				where = append(where, r.Where())
			}
			fmt.Printf("         points at %s\n", strings.Join(where, ", "))
		}
	ask:
		fmt.Print("  [c]onfirm, [e]dit, [r]etire, [s]kip or [q]uit? ")
		line, err := in.ReadString('\n')
		if err != nil && line == "" {
			fmt.Println()
			return // no more input
		}
		switch strings.ToLower(strings.TrimSpace(line)) {
		case "c", "confirm":
			if err := rv.confirm(m); err != nil {
				fmt.Println("  ", err)
			}
		case "e", "edit":
			fmt.Print("  The fact as it is now: ")
			text, _ := in.ReadString('\n')
			text = strings.TrimSpace(text)
			if text == "" {
				fmt.Println("  Nothing typed, so it's unchanged.")
				goto ask
			}
			if err := rv.edit(m, text); err != nil {
				fmt.Println("  ", err)
				goto ask
			}
		case "r", "retire":
			if _, err := rv.a.store.Expire(m.ID); err != nil {
				fmt.Println("  ", err)
				continue
			}
			rv.changed++
			fmt.Println("  Retired. It's kept as history.")
		case "s", "skip", "":
			fmt.Println("  Skipped.")
		case "q", "quit":
			return
		default:
			goto ask
		}
	}
}

type reviewer struct {
	a       *app
	repo    *share.Repo
	changed int
}

// sources are the fact's sources pointing at the code as it is now,
// following renames and moves, or nil outside the repo.
func (rv *reviewer) sources(m memory.Memory) ([]memory.Ref, string) {
	if rv.repo == nil || len(m.Sources) == 0 {
		return nil, ""
	}
	g := ground.Open(rv.repo.Dir)
	refs := m.Sources
	if res := g.Check(m); len(res.Refs) > 0 {
		refs = res.Refs
	}
	return g.Resolve(refs), gitinfo.Head(rv.repo.Dir)
}

func (rv *reviewer) confirm(m memory.Memory) error {
	refs, head := rv.sources(m)
	if rv.repo != nil && len(m.Sources) > 0 && len(refs) == 0 {
		refs = []memory.Ref{} // its code is gone; keep the fact without sources
	}
	if _, err := rv.a.store.Confirm(m.ID, refs, head); err != nil {
		return err
	}
	rv.changed++
	msg := "  Confirmed. It's current again"
	if refs != nil && len(refs) < len(m.Sources) {
		msg += ", without the sources that are gone"
	}
	fmt.Println(msg + ".")
	return nil
}

func (rv *reviewer) edit(m memory.Memory, text string) error {
	refs, head := rv.sources(m)
	if rv.repo != nil && len(m.Sources) > 0 && len(refs) == 0 {
		refs = []memory.Ref{} // its code is gone; don't carry the dead sources over
	}
	_, err := rv.a.store.Replace(m.ID, memory.RememberInput{
		Content: text, Tags: m.Tags, Type: m.Type, Sources: refs, Commit: head,
		Source: memory.SourceUser, Via: cli, Project: m.Project,
	})
	if err != nil {
		return err
	}
	rv.changed++
	fmt.Println("  Saved. The old version is kept as history.")
	return nil
}

func (rv *reviewer) done() {
	if rv.changed > 0 && rv.repo != nil {
		rv.a.syncRepo(rv.repo.Dir)
	}
	rv.a.record("review", nil, rv.changed)
}

func printStale(m memory.Memory) {
	fmt.Printf("stale  %s\n", clip(m.Content, 80))
	why := m.StaleReason
	if why == "" {
		why = "flagged stale"
	}
	if m.StaleCommit != "" {
		why += " in " + m.StaleCommit
	}
	if m.StaleAt != nil {
		why += ", " + ago(*m.StaleAt)
	}
	fmt.Printf("         %s\n", why)
}

func ago(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return plural(int(d.Minutes()), "minute ago", "minutes ago")
	case d < 48*time.Hour:
		return plural(int(d.Hours()), "hour ago", "hours ago")
	}
	return plural(int(d.Hours()/24), "day ago", "days ago")
}

func interactive() bool {
	st, err := os.Stdin.Stat()
	return err == nil && st.Mode()&os.ModeCharDevice != 0
}

// reviewOld lists old guesses that may no longer be true, or keeps one.
// Nothing here expires a memory on its own.
func reviewOld(keep string) error {
	a, err := open()
	if err != nil {
		return err
	}
	if keep != "" {
		m, err := a.store.Get(keep)
		if err != nil {
			return err
		}
		a.store.Touch(m.ID)
		fmt.Println("Kept: " + m.Content)
		return nil
	}
	mems, err := a.store.ReviewQueue(time.Now())
	if err != nil {
		return err
	}
	if len(mems) == 0 {
		fmt.Println("Nothing to check. Every AI guess was used or changed in the last 90 days.")
		return nil
	}
	fmt.Printf("%s an AI guessed that nothing has used in 90 days. Still true?\n\n", plural(len(mems), "memory", "memories"))
	for _, m := range mems {
		fmt.Printf("  %s  %s\n", m.ID, m.Content)
	}
	fmt.Println("\nKeep one with: woolly review --keep <id>")
	fmt.Println("Fix one on the review page: woolly ui")
	fmt.Println("Drop one with: woolly forget <id>")
	return nil
}
