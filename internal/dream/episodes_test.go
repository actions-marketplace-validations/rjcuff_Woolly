package dream

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/rjcuff/woolly/internal/memory"
)

type summaryModel struct{ asked int }

func (s *summaryModel) Chat(_ context.Context, system, user string, _ any) (string, error) {
	if system == summarizePrompt {
		s.asked++
		return `{"summary":"Monday the deploy failed on a bad env var; Wednesday it was fixed by adding DATABASE_URL to Fly secrets."}`, nil
	}
	return `{"relation":"unrelated"}`, nil
}

func TestSummarizesAWeekOfEpisodes(t *testing.T) {
	m := &summaryModel{}
	d, store := setup(t, m)
	a := save(t, store, memory.RememberInput{Content: "Deploy failed: missing env var", Type: "episode"})
	b := save(t, store, memory.RememberInput{Content: "Fixed deploy by adding DATABASE_URL to Fly secrets", Type: "episode"})
	mine := save(t, store, memory.RememberInput{Content: "I paired with Sam on the fix", Type: "episode", Source: memory.SourceUser})
	alone := save(t, store, memory.RememberInput{Content: "Standup was short", Type: "episode", User: "alice"})

	rep := &Report{Steps: []Step{}}
	later := time.Now().Add(EpisodeAge + time.Hour)
	if err := d.summarizeEpisodes(context.Background(), Options{}, rep, map[string]bool{}, later); err != nil {
		t.Fatal(err)
	}
	if m.asked != 1 || len(rep.Steps) != 1 {
		t.Fatalf("asked %d, steps %+v", m.asked, rep.Steps)
	}
	cur := strings.Join(current(t, store, memory.Scope{}), "\n")
	if !strings.Contains(cur, "DATABASE_URL to Fly secrets.") || strings.Contains(cur, "Deploy failed: missing") {
		t.Fatalf("current memories:\n%s", cur)
	}
	for _, id := range []string{a.ID, b.ID} {
		if got, _ := store.Get(id); got.Status != memory.StatusSuperseded {
			t.Fatalf("episode %s is %s, want kept as history", id, got.Status)
		}
	}
	// What a person saved keeps their words; a lone episode has nothing to roll up.
	for _, id := range []string{mine.ID, alone.ID} {
		if got, _ := store.Get(id); got.Status != memory.StatusActive {
			t.Fatalf("%s is %s", id, got.Status)
		}
	}
	// Recent episodes are left alone.
	rep = &Report{Steps: []Step{}}
	m.asked = 0
	save(t, store, memory.RememberInput{Content: "Another deploy today", Type: "episode"})
	save(t, store, memory.RememberInput{Content: "And one more deploy today", Type: "episode"})
	if err := d.summarizeEpisodes(context.Background(), Options{}, rep, map[string]bool{}, time.Now()); err != nil || m.asked != 0 {
		t.Fatalf("recent episodes summarized: %v asked %d", err, m.asked)
	}
}
