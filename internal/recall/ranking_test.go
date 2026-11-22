package recall

import (
	"context"
	"testing"

	"github.com/rjcuff/woolly/internal/memory"
)

func newEngine(t *testing.T) (*Engine, *memory.Store) {
	t.Helper()
	store, err := memory.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return New(store, nil, ""), store
}

func TestNoOverlapFindsNothing(t *testing.T) {
	e, store := newEngine(t)
	store.Remember(memory.RememberInput{Content: "Lives in Denver"})
	got, _, _ := e.Recall(context.Background(), "quantum chromodynamics", 5, false)
	if len(got) != 0 {
		t.Fatalf("got %v", got)
	}
}

func TestWhatTheUserSaidRanksAboveAGuess(t *testing.T) {
	e, store := newEngine(t)
	store.Remember(memory.RememberInput{Content: "Deploys the web app with Vercel", Source: memory.SourceAgent})
	store.Remember(memory.RememberInput{Content: "Deploys the api with Fly", Source: memory.SourceUser})
	got, _, _ := e.Recall(context.Background(), "deploys", 5, false)
	if len(got) != 2 || got[0].Memory.Source != memory.SourceUser {
		t.Fatalf("order: %+v", got)
	}
	if got[0].Parts == nil || got[0].Parts.Confidence <= got[1].Parts.Confidence {
		t.Fatalf("score parts: %+v %+v", got[0].Parts, got[1].Parts)
	}
}

func TestNearCopiesAreSkipped(t *testing.T) {
	e, store := newEngine(t)
	store.Remember(memory.RememberInput{Content: "Prefers tabs over spaces in Go code", Key: "a"})
	store.Remember(memory.RememberInput{Content: "Prefers tabs over spaces in Go code!", Key: "b"})
	store.Remember(memory.RememberInput{Content: "Uses spaces in Python code", Key: "c"})
	got, _, _ := e.Recall(context.Background(), "tabs spaces code", 3, false)
	if len(got) != 2 {
		t.Fatalf("got %d results: %+v", len(got), got)
	}
}

func TestDownVoteLowersRank(t *testing.T) {
	e, store := newEngine(t)
	a, _ := store.Remember(memory.RememberInput{Content: "The staging database is Postgres"})
	b, _ := store.Remember(memory.RememberInput{Content: "The prod database is Postgres"})
	first := func() string {
		got, _, _ := e.Recall(context.Background(), "database postgres", 2, false)
		return got[0].Memory.ID
	}
	top := first()
	other := a.Memory.ID
	if top == a.Memory.ID {
		other = b.Memory.ID
	}
	for i := 0; i < 3; i++ {
		store.Feedback(top, false)
		store.Feedback(other, true)
	}
	if first() != other {
		t.Fatal("down-voted memory still ranks first")
	}
	if _, err := store.Get(top); err != nil {
		t.Fatal("a down-vote must never delete a memory")
	}
}
