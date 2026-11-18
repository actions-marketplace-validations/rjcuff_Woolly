package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rjcuff/woolly/internal/dream"
	"github.com/rjcuff/woolly/internal/memory"
	"github.com/rjcuff/woolly/internal/share"
)

// The recap is what the last dream did, kept in ~/.woolly/recap.json so the
// next session can open with it ("While you were away...") and woolly recap
// can show it. Each project sees it once.

// mammoth is Woolly's pixel mammoth from the website, in half blocks.
const mammoth = `          ▄███▄
     ▄▄▄▄▄██████▄
  ▄███████████████
▄█████████████████
██████████████ ███▄
 █████████████▄████
 █████████████▄██▀
  ▀▀▀▀▀ ▀▀▀▀▀  ▀`

type recapLesson struct {
	Project string `json:"project,omitempty"`
	Text    string `json:"text"`
}

type recapState struct {
	At        time.Time            `json:"at"`
	Merged    int                  `json:"merged"`
	Settled   int                  `json:"settled"`
	Conflicts int                  `json:"conflicts"`
	Lessons   []recapLesson        `json:"lessons,omitempty"`
	Shown     map[string]time.Time `json:"shown,omitempty"` // by project, when a session brief showed it
}

// add folds a dream's report into the recap. project is the project the
// dream was scoped to, which its lessons belong to.
func (r *recapState) add(rep *dream.Report, project string) {
	r.Merged += rep.Merged
	r.Settled += rep.Settled
	r.Conflicts += rep.Conflicts
	for _, st := range rep.Steps {
		if st.Action == "learn" && strings.TrimSpace(st.Memory) != "" {
			r.Lessons = append(r.Lessons, recapLesson{Project: memory.NormalizeProject(project), Text: strings.TrimSpace(st.Memory)})
		}
	}
}

func (r *recapState) empty() bool {
	return r.Merged+r.Settled+r.Conflicts+len(r.Lessons) == 0
}

func (a *app) recapPath() string { return filepath.Join(a.home, "recap.json") }

func (a *app) loadRecap() *recapState {
	r := &recapState{}
	if b, err := os.ReadFile(a.recapPath()); err == nil {
		_ = json.Unmarshal(b, r)
	}
	return r
}

func (a *app) writeRecap(r *recapState) {
	if b, err := json.MarshalIndent(r, "", "  "); err == nil {
		_ = os.WriteFile(a.recapPath(), b, 0o600)
	}
}

// saveRecap keeps a dream's results for the next session. A dream that did
// nothing leaves the last recap alone.
func (a *app) saveRecap(r *recapState) {
	if r.empty() {
		return
	}
	r.At = time.Now().UTC()
	r.Shown = nil
	a.writeRecap(r)
}

// recapText describes the recap for one project: what was tidied, and the
// lessons that belong to it. It returns "" when there's nothing to say.
func recapText(r *recapState, project string) string {
	if r.At.IsZero() {
		return ""
	}
	var did []string
	if r.Merged > 0 {
		did = append(did, "folded "+plural(r.Merged, "repeat", "repeats")+" together")
	}
	if r.Settled > 0 {
		did = append(did, "settled "+plural(r.Settled, "fact that disagreed", "facts that disagreed"))
	}
	if r.Conflicts > 0 {
		did = append(did, "set aside "+plural(r.Conflicts, "conflict", "conflicts")+" for you to review")
	}
	var lessons []string
	for _, l := range r.Lessons {
		if l.Project == "" || l.Project == memory.NormalizeProject(project) {
			lessons = append(lessons, l.Text)
		}
	}
	if len(did) == 0 && len(lessons) == 0 {
		return ""
	}
	var b strings.Builder
	when := r.At.Local().Format("Monday")
	if time.Since(r.At) < 20*time.Hour {
		when = "overnight"
	}
	b.WriteString("While you were away, Woolly dreamed " + when)
	if len(did) > 0 {
		b.WriteString(". It " + joinAnd(did))
	}
	b.WriteString(".")
	if len(lessons) > 0 {
		fmt.Fprintf(&b, " It learned %s from recent sessions:\n", plural(len(lessons), "lesson", "lessons"))
		for _, l := range lessons {
			b.WriteString("- " + l + "\n")
		}
	} else {
		b.WriteString("\n")
	}
	return b.String()
}

func joinAnd(parts []string) string {
	switch len(parts) {
	case 0:
		return ""
	case 1:
		return parts[0]
	}
	return strings.Join(parts[:len(parts)-1], ", ") + " and " + parts[len(parts)-1]
}

// awaySection is the recap for a session brief, once per project per
// dream, and only while it's fresh.
func (a *app) awaySection(project string) string {
	r := a.loadRecap()
	if r.At.IsZero() || time.Since(r.At) > 3*24*time.Hour {
		return ""
	}
	key := memory.NormalizeProject(project)
	if seen, ok := r.Shown[key]; ok && seen.After(r.At) {
		return ""
	}
	text := recapText(r, project)
	if text == "" {
		return ""
	}
	if r.Shown == nil {
		r.Shown = map[string]time.Time{}
	}
	r.Shown[key] = time.Now().UTC()
	a.writeRecap(r)
	return "# While you were away\n\n" + text + "\n"
}

// cmdRecap shows what the last dream did.
func cmdRecap(args []string) error {
	fs := newFlags("recap")
	project := fs.String("project", "", "the project (default: this repo)")
	if _, err := parse(fs, args); err != nil {
		return err
	}
	p := *project
	if p == "" {
		if cwd, err := os.Getwd(); err == nil {
			if r := share.Find(cwd); r != nil {
				p = r.Project
			}
		}
	}
	a, err := open()
	if err != nil {
		return err
	}
	fmt.Println(mammoth)
	fmt.Println()
	text := recapText(a.loadRecap(), p)
	if text == "" {
		fmt.Println("Woolly hasn't dreamed yet. While it's running it dreams once a day, when memory has changed.")
		fmt.Println("To dream now, run: woolly dream")
		return nil
	}
	fmt.Print(text)
	return nil
}
