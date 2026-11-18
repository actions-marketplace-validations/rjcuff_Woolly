package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/rjcuff/woolly/internal/share"
	"github.com/rjcuff/woolly/internal/why"
)

// cmdWhy shows the decisions, conventions and gotchas behind a file, a line
// or a function, and whether the code still backs each one up.
func cmdWhy(args []string) error {
	fs := newFlags("why")
	asJSON := fs.Bool("json", false, "print the answer as JSON")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 1 {
		return errors.New("say which code, like: woolly why src/auth.ts, woolly why src/auth.ts:42 or woolly why src/auth.ts#requireUser")
	}
	cwd, _ := os.Getwd()
	r := share.Find(cwd)
	if r == nil {
		return errors.New("run this inside a git repo")
	}
	t, err := why.ParseTarget(pos[0], r.Dir, cwd)
	if err != nil {
		return err
	}
	a, err := open()
	if err != nil {
		return err
	}
	a.syncRepo(r.Dir)
	mems, err := a.store.List(true)
	if err != nil {
		return err
	}
	ans := why.Find(mems, r.Project, r.Dir, t)
	var ids []string
	for _, m := range ans.Matches {
		ids = append(ids, m.Memory.ID)
	}
	a.record("why", ids, len(ids))
	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(ans)
	}
	fmt.Print(strings.TrimRight(why.Text(ans), "\n") + "\n")
	return nil
}
