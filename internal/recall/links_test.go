package recall

import (
	"context"
	"testing"

	"github.com/rjcuff/woolly/internal/memory"
)

func TestLinkedMemoryComesAlong(t *testing.T) {
	e, store := newEngine(t)
	bug, _ := store.Remember(memory.RememberInput{Content: "The March outage came from a bad config push"})
	rule, _ := store.Remember(memory.RememberInput{Content: "Every change needs a canary before rollout"})
	if _, _, err := store.Link(bug.Memory.ID, rule.Memory.ID); err != nil {
		t.Fatal(err)
	}
	got, _, _ := e.Recall(context.Background(), "march outage", 5, false)
	if len(got) != 2 || got[1].Memory.ID != rule.Memory.ID || got[1].Parts.LinkedFrom != bug.Memory.ID {
		t.Fatalf("got %+v", got)
	}
	if got[1].Score > got[0].Score/2+0.001 {
		t.Fatalf("linked score %v vs %v", got[1].Score, got[0].Score)
	}
}

func TestLinksStayInScope(t *testing.T) {
	e, store := newEngine(t)
	a, _ := store.Remember(memory.RememberInput{Content: "Alice resets passwords by email", User: "alice"})
	b, _ := store.Remember(memory.RememberInput{Content: "Alice prefers SMS codes", User: "alice"})
	if _, _, err := store.Link(a.Memory.ID, b.Memory.ID); err != nil {
		t.Fatal(err)
	}
	mine, _ := store.Remember(memory.RememberInput{Content: "I reset passwords with 1Password"})
	if _, _, err := store.Link(mine.Memory.ID, a.Memory.ID); err == nil {
		t.Fatal("linked across scopes")
	}
	got, _, _ := e.RecallScoped(context.Background(), "reset passwords", 5, false, memory.Scope{Personal: true})
	for _, r := range got {
		if r.Memory.User != "" {
			t.Fatalf("an app user's memory came in: %+v", r.Memory)
		}
	}
}

func TestForgetRemovesLinks(t *testing.T) {
	_, store := newEngine(t)
	a, _ := store.Remember(memory.RememberInput{Content: "A"})
	b, _ := store.Remember(memory.RememberInput{Content: "B"})
	store.Link(a.Memory.ID, b.Memory.ID)
	if _, err := store.Forget(b.Memory.ID); err != nil {
		t.Fatal(err)
	}
	got, _ := store.Get(a.Memory.ID)
	if len(got.Links) != 0 {
		t.Fatalf("dangling links: %v", got.Links)
	}
}
