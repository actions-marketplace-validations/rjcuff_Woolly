package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rjcuff/woolly/internal/activity"
	"github.com/rjcuff/woolly/internal/dream"
	"github.com/rjcuff/woolly/internal/handoff"
	"github.com/rjcuff/woolly/internal/memory"
)

// dreamer is shared by the API and the nightly dream, so they never run at once.
func (a *app) dreamer() *dream.Dreamer {
	a.dreamOnce.Do(func() {
		a.dream = &dream.Dreamer{Store: a.store, Engine: a.engine, Model: a.ollama, Learner: a.learner()}
	})
	return a.dream
}

func cmdDream(args []string) error {
	fs := newFlags("dream")
	project := fs.String("project", "", "only this project, and write lessons from its past sessions")
	user := fs.String("user", "", "only this user_id")
	dryRun := fs.Bool("dry-run", false, "show what would change without changing anything")
	noLearn := fs.Bool("no-learn", false, "don't write lessons")
	asJSON := fs.Bool("json", false, "print JSON")
	if _, err := parse(fs, args); err != nil {
		return err
	}
	a, err := open()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	d := a.dreamer()
	if ok, err := a.ollama.HasModel(ctx, a.ollama.ChatModel); err != nil || !ok {
		// Without a model, exact repeats can still be folded together.
		d = &dream.Dreamer{Store: a.store, Engine: a.engine}
	}
	st := loadDreamState(a.home)
	rep, err := d.Dream(ctx, dream.Options{
		Scope: memory.Scope{Project: *project, User: *user}, DryRun: *dryRun, Learn: !*noLearn, Unrelated: st.set(),
	})
	if err != nil {
		return err
	}
	if !*dryRun {
		st.remember(rep)
		_ = st.save(a.home)
		a.log.Add(activity.Entry{App: cli, Action: "dream", IDs: results(rep), Count: len(rep.Steps)})
	}
	if *asJSON {
		return printJSON(rep)
	}
	printDream(rep, *dryRun)
	if rep.Note != "" {
		fmt.Println("Note:", rep.Note)
		if d.Model == nil {
			fmt.Printf("For the rest, run: ollama pull %s\n", a.ollama.ChatModel)
		}
	}
	return nil
}

func printDream(rep *dream.Report, dryRun bool) {
	verb := map[string]string{"merge": "Merged", "settle": "Settled", "conflict": "Needs review", "learn": "Learned"}
	if dryRun {
		verb = map[string]string{"merge": "Would merge", "settle": "Would settle", "conflict": "Would flag", "learn": "Would learn"}
	}
	for _, st := range rep.Steps {
		line := fmt.Sprintf("%s: %s", verb[st.Action], st.Memory)
		if st.Reason != "" && st.Action != "learn" {
			line += " (" + st.Reason + ")"
		}
		fmt.Println(line)
	}
	if len(rep.Steps) == 0 {
		fmt.Println("Nothing to tidy. Memory is in good shape.")
	}
	fmt.Printf("\nCompared %d pairs: %d merged, %d settled, %d to review, %d lessons.\n",
		rep.Compared, rep.Merged, rep.Settled, rep.Conflicts, rep.Learned)
}

func results(rep *dream.Report) []string {
	var ids []string
	for _, st := range rep.Steps {
		if st.Result != "" {
			ids = append(ids, st.Result)
		}
	}
	return ids
}

// dreamState is what the nightly dream remembers between runs.
type dreamState struct {
	Last      time.Time `json:"last"`
	Unrelated []string  `json:"unrelated,omitempty"`
}

func dreamStatePath(home string) string { return filepath.Join(home, "dream.json") }

func loadDreamState(home string) *dreamState {
	st := &dreamState{}
	if data, err := os.ReadFile(dreamStatePath(home)); err == nil {
		_ = json.Unmarshal(data, st)
	}
	return st
}

func (st *dreamState) set() map[string]bool {
	out := make(map[string]bool, len(st.Unrelated))
	for _, k := range st.Unrelated {
		out[k] = true
	}
	return out
}

// remember notes the time and the pairs found unrelated, keeping the newest 5000.
func (st *dreamState) remember(rep *dream.Report) {
	st.Last = time.Now().UTC()
	st.Unrelated = append(st.Unrelated, rep.Unrelated...)
	if n := len(st.Unrelated); n > 5000 {
		st.Unrelated = st.Unrelated[n-5000:]
	}
}

func (st *dreamState) save(home string) error {
	data, err := json.Marshal(st)
	if err != nil {
		return err
	}
	return os.WriteFile(dreamStatePath(home), data, 0o600)
}

// dreamInBackground dreams at most once a day while Woolly is running, when
// memory has changed and a chat model is installed. WOOLLY_DREAM=off stops it.
func (a *app) dreamInBackground() {
	if os.Getenv("WOOLLY_DREAM") == "off" {
		return
	}
	time.Sleep(10 * time.Minute) // let a fresh start settle first
	for {
		a.dreamIfDue()
		time.Sleep(30 * time.Minute)
	}
}

func (a *app) dreamIfDue() {
	st := loadDreamState(a.home)
	if time.Since(st.Last) < 20*time.Hour {
		return
	}
	all, err := a.store.List(true)
	if err != nil {
		return
	}
	changed := false
	projects := map[string]bool{}
	for _, m := range all {
		if m.UpdatedAt.After(st.Last) {
			changed = true
			// Only projects with a new session have new lessons to find.
			if m.Key == handoff.Key && m.Project != "" {
				projects[m.Project] = true
			}
		}
	}
	if !changed {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	if ok, err := a.ollama.HasModel(ctx, a.ollama.ChatModel); err != nil || !ok {
		return
	}
	d := a.dreamer()
	total := &dream.Report{}
	rep, err := d.Dream(ctx, dream.Options{Unrelated: st.set()})
	if err != nil {
		return // busy, or the store failed; try again later
	}
	st.remember(rep)
	merge(total, rep)
	for p := range projects {
		if r, err := d.Dream(ctx, dream.Options{Scope: memory.Scope{Project: p}, Learn: true, OnlyLessons: true}); err == nil {
			merge(total, r)
		}
	}
	_ = st.save(a.home)
	if len(total.Steps) > 0 {
		a.log.Add(activity.Entry{App: dream.Via, Action: "dream", IDs: results(total), Count: len(total.Steps)})
		fmt.Fprintf(os.Stderr, "woolly: dreamed: %d merged, %d settled, %d to review, %d lessons\n",
			total.Merged, total.Settled, total.Conflicts, total.Learned)
	}
}

func merge(total, r *dream.Report) {
	total.Steps = append(total.Steps, r.Steps...)
	total.Merged += r.Merged
	total.Settled += r.Settled
	total.Conflicts += r.Conflicts
	total.Learned += r.Learned
}
