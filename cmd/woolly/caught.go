package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rjcuff/woolly/internal/ground"
	"github.com/rjcuff/woolly/internal/memory"
	"github.com/rjcuff/woolly/internal/share"
)

// Every fact Woolly flags is a fact an agent would otherwise have followed.
// They're kept in ~/.woolly/caught.json, by project, so Woolly can say how
// many it has caught.

type caughtFact struct {
	ID      string    `json:"id"`
	Content string    `json:"content"`
	Reason  string    `json:"reason"`
	Commit  string    `json:"commit,omitempty"`
	At      time.Time `json:"at"`
}

func (a *app) caughtPath() string { return filepath.Join(a.home, "caught.json") }

func (a *app) loadCaught() map[string][]caughtFact {
	all := map[string][]caughtFact{}
	if b, err := os.ReadFile(a.caughtPath()); err == nil {
		_ = json.Unmarshal(b, &all)
	}
	return all
}

// noteCaught records a fact a check just flagged. The same fact going stale
// again at the same commit counts once.
func (a *app) noteCaught(project string, res ground.Result) {
	all := a.loadCaught()
	for _, c := range all[project] {
		if c.ID == res.ID && c.Commit == res.Commit {
			return
		}
	}
	all[project] = append(all[project], caughtFact{ID: res.ID, Content: res.Content, Reason: res.Reason, Commit: res.Commit, At: time.Now().UTC()})
	if b, err := json.MarshalIndent(all, "", "  "); err == nil {
		_ = os.WriteFile(a.caughtPath(), b, 0o600)
	}
}

func (a *app) caught(project string) []caughtFact { return a.loadCaught()[project] }

// badgeURL is a shields.io badge with the count, in Woolly blue.
func badgeURL(n int) string {
	label := plural(n, "stale fact caught", "stale facts caught")
	return "https://img.shields.io/badge/woolly-" + strings.ReplaceAll(url.PathEscape(label), "-", "--") + "-1a3fd1"
}

// cmdCaught lists the stale facts Woolly has caught in this repo.
func cmdCaught(args []string) error {
	fs := newFlags("caught")
	badge := fs.Bool("badge", false, "print a README badge with the count")
	asJSON := fs.Bool("json", false, "print them as JSON")
	project := fs.String("project", "", "the project (default: this repo)")
	if _, err := parse(fs, args); err != nil {
		return err
	}
	p := memory.NormalizeProject(*project)
	if p == "" {
		cwd, _ := os.Getwd()
		r := share.Find(cwd)
		if r == nil {
			return errors.New("run this inside a git repo, or pass --project")
		}
		p = r.Project
	}
	a, err := open()
	if err != nil {
		return err
	}
	list := a.caught(p)
	switch {
	case *asJSON:
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if list == nil {
			list = []caughtFact{}
		}
		return enc.Encode(list)
	case *badge:
		fmt.Printf("[![%s](%s)](https://usewoolly.com)\n", plural(len(list), "stale fact caught by Woolly", "stale facts caught by Woolly"), badgeURL(len(list)))
		return nil
	}
	if len(list) == 0 {
		fmt.Printf("Woolly hasn't caught a stale fact in %s yet.\n", p)
		fmt.Println("When a commit changes code a fact points at, woolly check flags it and keeps it away from your agents.")
		return nil
	}
	fmt.Printf("Woolly has caught %s in %s. Your agents would have followed each one.\n\n", plural(len(list), "stale fact", "stale facts"), p)
	for i := len(list) - 1; i >= 0; i-- {
		c := list[i]
		why := c.Reason
		if c.Commit != "" {
			why += " in " + c.Commit
		}
		fmt.Printf("  %s  %s\n             %s\n", c.At.Local().Format("Jan 02"), clip(c.Content, 72), why)
	}
	fmt.Println("\nAdd the count to your README with: woolly caught --badge")
	return nil
}
