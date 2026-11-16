package learn

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/rjcuff/woolly/internal/memory"
)

// scriptModel returns the given facts when asked to extract, and the given
// decision for each new fact when asked to decide.
type scriptModel struct {
	facts     []string
	decisions map[string]string // new fact -> decision JSON
	asked     []string          // the decide prompts it saw
}

func (s *scriptModel) Chat(_ context.Context, system, user string, _ any) (string, error) {
	switch system {
	case extractPrompt:
		var out struct {
			Facts []fact `json:"facts"`
		}
		for _, f := range s.facts {
			out.Facts = append(out.Facts, fact{Fact: f})
		}
		b, _ := json.Marshal(out)
		return string(b), nil
	case decidePrompt:
		s.asked = append(s.asked, user)
		for f, d := range s.decisions {
			if strings.HasSuffix(user, "New fact: "+f) {
				return d, nil
			}
		}
		return `{"event":"ADD","target":-1}`, nil
	}
	return `{"topic":"none"}`, nil
}

func say(text string) []Message { return []Message{{Role: "user", Content: text}} }

func events(evs []Event) string {
	var out []string
	for _, e := range evs {
		out = append(out, e.Event)
	}
	return strings.Join(out, ",")
}

func TestAddUpdatesAndKeepsHistory(t *testing.T) {
	m := &scriptModel{facts: []string{"Lives in Denver"}}
	l, store := setup(t, m)
	alice := AddOptions{Scope: memory.Scope{User: "alice"}, Infer: true}

	evs, err := l.Add(context.Background(), say("I live in Denver"), alice)
	if err != nil || events(evs) != "ADD" {
		t.Fatalf("first add: %v %v", events(evs), err)
	}
	first := evs[0].ID

	m.facts = []string{"Lives in Austin"}
	m.decisions = map[string]string{"Lives in Austin": `{"event":"UPDATE","target":0,"text":"Lives in Austin"}`}
	evs, err = l.Add(context.Background(), say("I moved to Austin"), alice)
	if err != nil || events(evs) != "UPDATE" {
		t.Fatalf("update: %v %v", events(evs), err)
	}
	if evs[0].PreviousMemory != "Lives in Denver" || evs[0].Memory != "Lives in Austin" {
		t.Fatalf("update event: %+v", evs[0])
	}

	old, _ := store.Get(first)
	if old.Status != memory.StatusSuperseded || old.InvalidAt == nil || old.SupersededBy != evs[0].ID {
		t.Fatalf("old memory should be superseded with a time: %+v", old)
	}
	cur, _ := store.ListIn(memory.Scope{User: "alice"}, false)
	if len(cur) != 1 || cur[0].Content != "Lives in Austin" || cur[0].User != "alice" {
		t.Fatalf("current memories: %+v", cur)
	}
	hist, err := store.History(evs[0].ID)
	if err != nil || len(hist) != 2 || hist[0].ID != first || hist[1].ID != evs[0].ID {
		t.Fatalf("history: %+v %v", hist, err)
	}
}

func TestAddDeleteExpires(t *testing.T) {
	m := &scriptModel{facts: []string{"Uses Vim"}}
	l, store := setup(t, m)
	o := AddOptions{Scope: memory.Scope{User: "alice"}, Infer: true}
	evs, _ := l.Add(context.Background(), say("I use Vim"), o)
	id := evs[0].ID

	m.facts = []string{"No longer uses Vim"}
	m.decisions = map[string]string{"No longer uses Vim": `{"event":"DELETE","target":0}`}
	evs, err := l.Add(context.Background(), say("I stopped using Vim"), o)
	if err != nil || events(evs) != "DELETE" || evs[0].ID != id {
		t.Fatalf("delete: %+v %v", evs, err)
	}
	got, _ := store.Get(id)
	if got.Status != memory.StatusExpired || got.InvalidAt == nil {
		t.Fatalf("should be expired with a time: %+v", got)
	}
	if cur, _ := store.ListIn(o.Scope, false); len(cur) != 0 {
		t.Fatalf("an expired memory isn't current: %+v", cur)
	}
}

func TestAddNeverOverrulesWhatYouSaid(t *testing.T) {
	m := &scriptModel{facts: []string{"Lives in Austin"}, decisions: map[string]string{
		"Lives in Austin": `{"event":"UPDATE","target":0,"text":"Lives in Austin"}`,
	}}
	l, store := setup(t, m)
	said, _ := store.Remember(memory.RememberInput{Content: "Lives in Denver", Source: memory.SourceUser})

	evs, err := l.Add(context.Background(), say("guess I'm in Austin now"), AddOptions{Infer: true})
	if err != nil || events(evs) != "CONFLICT" {
		t.Fatalf("want a conflict: %+v %v", evs, err)
	}
	kept, _ := store.Get(said.Memory.ID)
	if kept.Status != memory.StatusActive || kept.Content != "Lives in Denver" {
		t.Fatalf("what you said should stand: %+v", kept)
	}
	if c, _ := store.Conflicts(); len(c) != 1 || c[0].Content != "Lives in Austin" {
		t.Fatalf("the guess should wait for review: %+v", c)
	}
}

func TestAddKeepsUsersApart(t *testing.T) {
	m := &scriptModel{facts: []string{"Lives in Denver"}}
	l, store := setup(t, m)
	l.Add(context.Background(), say("I live in Denver"), AddOptions{Scope: memory.Scope{User: "alice"}, Infer: true})

	// Bob's fact would update anything it was shown, but Alice's memory is never shown.
	m.facts = []string{"Lives in Austin"}
	m.decisions = map[string]string{"Lives in Austin": `{"event":"UPDATE","target":0,"text":"Lives in Austin"}`}
	evs, _ := l.Add(context.Background(), say("I live in Austin"), AddOptions{Scope: memory.Scope{User: "bob"}, Infer: true})
	if events(evs) != "ADD" || len(m.asked) != 0 {
		t.Fatalf("bob should start fresh: %+v, asked %d", evs, len(m.asked))
	}
	for user, want := range map[string]string{"alice": "Lives in Denver", "bob": "Lives in Austin"} {
		cur, _ := store.ListIn(memory.Scope{User: user}, false)
		if len(cur) != 1 || cur[0].Content != want {
			t.Fatalf("%s: %+v", user, cur)
		}
	}
}

func TestAddSkipsWhatItKnowsAndSurvivesBadAnswers(t *testing.T) {
	m := &scriptModel{facts: []string{"Prefers dark mode"}}
	l, _ := setup(t, m)
	o := AddOptions{Scope: memory.Scope{Agent: "helper"}, Infer: true}
	l.Add(context.Background(), say("I like dark mode"), o)

	evs, _ := l.Add(context.Background(), say("dark mode please"), o)
	if events(evs) != "NONE" {
		t.Fatalf("a repeat is NONE: %+v", evs)
	}

	// A target that isn't on the list falls back to adding.
	m.facts = []string{"Prefers tabs"}
	m.decisions = map[string]string{"Prefers tabs": `{"event":"UPDATE","target":9,"text":"x"}`}
	evs, _ = l.Add(context.Background(), say("tabs"), o)
	if events(evs) != "ADD" || evs[0].Memory != "Prefers tabs" {
		t.Fatalf("bad target should add: %+v", evs)
	}
}

func TestAddWithoutInferSavesUserMessages(t *testing.T) {
	l, store := setup(t, nil)
	msgs := []Message{{Role: "user", Content: "Ship on Fridays"}, {Role: "assistant", Content: "Noted"}}
	evs, err := l.Add(context.Background(), msgs, AddOptions{Scope: memory.Scope{Session: "run-1"}, Metadata: map[string]any{"team": "web"}})
	if err != nil || events(evs) != "ADD" {
		t.Fatalf("raw add: %+v %v", evs, err)
	}
	got, _ := store.Get(evs[0].ID)
	if got.Session != "run-1" || got.Metadata["team"] != "web" || got.Source != memory.SourceAgent {
		t.Fatalf("saved: %+v", got)
	}
	if _, err := l.Add(context.Background(), msgs, AddOptions{Infer: true}); !errors.Is(err, ErrNoModel) {
		t.Fatalf("infer without a model: %v", err)
	}
}
