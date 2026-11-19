package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rjcuff/woolly/internal/audit"
	"github.com/rjcuff/woolly/internal/gitinfo"
	"github.com/rjcuff/woolly/internal/ground"
)

// The memory agents already read, like CLAUDE.md, Claude Code's auto
// memory, Codex memories and claude-mem, checked against the code. The last
// result for each project is kept in ~/.woolly/audits so session briefs can
// warn about it without searching history again.

var auditEnv = audit.DefaultEnv

func (a *app) auditPath(project string) string {
	return filepath.Join(a.home, "audits", strings.ReplaceAll(project, string(filepath.Separator), "_")+".json")
}

// auditRepo checks the repo's memory files and keeps the result. A result
// cut short by the deadline isn't kept, so it can't hide older findings.
func (a *app) auditRepo(dir, project string, deadline time.Time) audit.Report {
	rep := audit.Run(dir, audit.Options{Env: auditEnv(), Deadline: deadline})
	if rep.Head == "" || rep.Partial {
		return rep
	}
	if b, err := json.MarshalIndent(rep, "", "  "); err == nil {
		_ = os.MkdirAll(filepath.Dir(a.auditPath(project)), 0o700)
		_ = os.WriteFile(a.auditPath(project), b, 0o600)
	}
	for _, f := range rep.Findings {
		a.noteCaught(project, ground.Result{ID: "file:" + f.Where() + ":" + f.What, Content: f.Text, Reason: f.Reason, Commit: f.Commit})
	}
	return rep
}

func (a *app) lastAudit(project string) (audit.Report, bool) {
	var rep audit.Report
	b, err := os.ReadFile(a.auditPath(project))
	if err != nil || json.Unmarshal(b, &rep) != nil {
		return rep, false
	}
	return rep, true
}

// fileNames lists the memory files an audit read, like "CLAUDE.md, AGENTS.md
// and Claude Code memory (4 files)".
func fileNames(rep audit.Report) string {
	var order []string
	count := map[string]int{}
	for _, f := range rep.Files {
		if count[f.Source] == 0 {
			order = append(order, f.Source)
		}
		count[f.Source]++
	}
	var parts []string
	for _, s := range order {
		if count[s] > 1 {
			s = fmt.Sprintf("%s (%d files)", s, count[s])
		}
		parts = append(parts, s)
	}
	switch len(parts) {
	case 0:
		return ""
	case 1:
		return parts[0]
	}
	return strings.Join(parts[:len(parts)-1], ", ") + " and " + parts[len(parts)-1]
}

func findingWhy(f audit.Finding) string {
	why := f.Reason
	if f.Commit != "" {
		why += " in " + f.Commit
		if f.Date != "" {
			why += " (" + f.Date + ")"
		}
	}
	return why
}

// printAudit is the terminal report for memory files.
func printAudit(rep audit.Report) {
	if len(rep.Files) == 0 {
		return
	}
	fmt.Printf("%s %s in %s against %s.\n", paint(bold, "Checked"), plural(rep.Lines, "line", "lines"), fileNames(rep), rep.Head)
	width := 0
	for _, f := range rep.Findings {
		if n := len(f.Where()); n > width && n <= 32 {
			width = n
		}
	}
	for _, f := range rep.Findings {
		fmt.Printf("  %s  %s\n", paint(red, pad(f.Where(), width)), clip(f.Text, 72))
		fmt.Printf("  %s  %s\n", pad("", width), paint(dim, findingWhy(f)))
	}
	for _, n := range rep.Notes {
		fmt.Printf("  %s %s\n", paint(blue, "note"), n)
	}
	switch {
	case len(rep.Findings) > 0:
		fmt.Printf("%s Your agents still read %s as true.\n", paint(red, plural(len(rep.Findings), "line names code that's gone.", "lines name code that's gone.")), pick(len(rep.Findings), "it", "them"))
	case rep.Partial:
		fmt.Println("Ran out of time before checking every line. Run woolly check again to finish.")
	default:
		fmt.Println(paint(green, "Every function, file and command they name is still in the code."))
	}
}

func pad(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}

// markdownAudit is the memory file part of a pull request comment. With a
// base, it separates lines this change broke from ones that were already
// out of date.
func markdownAudit(dir string, rep audit.Report, base string) string {
	if len(rep.Files) == 0 {
		return ""
	}
	cell := func(s string) string { return strings.ReplaceAll(strings.ReplaceAll(s, "|", "\\|"), "\n", " ") }
	var now, before []audit.Finding
	for _, f := range rep.Findings {
		if base != "" && f.Commit != "" && gitinfo.IsAncestor(dir, f.Commit, base) {
			before = append(before, f)
		} else {
			now = append(now, f)
		}
	}
	var b strings.Builder
	if len(now) == 0 {
		fmt.Fprintf(&b, "\n**Woolly** checked %s in %s. None of them went stale.\n", plural(rep.Lines, "line", "lines"), fileNames(rep))
	} else {
		fmt.Fprintf(&b, "\n**Woolly** checked %s in %s. **%s in agent files went stale.**\n\n", plural(rep.Lines, "line", "lines"), fileNames(rep), plural(len(now), "line", "lines"))
		b.WriteString("Agents read these as true, but the code no longer has what they name.\n\n| Line | Says | What changed |\n| --- | --- | --- |\n")
		for _, f := range now {
			fmt.Fprintf(&b, "| `%s` | %s | %s |\n", cell(f.Where()), cell(clip(f.Text, 120)), cell(findingWhy(f)))
		}
	}
	if len(before) > 0 {
		fmt.Fprintf(&b, "\n<details><summary>%s already out of date before this change</summary>\n\n", plural(len(before), "line was", "lines were"))
		for _, f := range before {
			fmt.Fprintf(&b, "- `%s` %s (%s)\n", cell(f.Where()), cell(clip(f.Text, 100)), cell(findingWhy(f)))
		}
		b.WriteString("\n</details>\n")
	}
	return b.String()
}

// auditSection is what a session brief says about memory files that name
// code that's gone, so the agent doesn't follow them.
func auditSection(rep audit.Report) string {
	if len(rep.Findings) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("# Out of date lines in your memory files\n\n")
	fmt.Fprintf(&b, "Woolly checked %s against the code. These name a function, file or command that's gone, so don't rely on them, and suggest fixing them when they come up:\n", fileNames(rep))
	for i, f := range rep.Findings {
		if i == 8 {
			fmt.Fprintf(&b, "- and %d more. Run woolly check to see them all.\n", len(rep.Findings)-8)
			break
		}
		fmt.Fprintf(&b, "- %s %q (%s)\n", f.Where(), clip(f.Text, 100), findingWhy(f))
	}
	return b.String() + "\n"
}
