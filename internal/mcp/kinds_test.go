package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/rjcuff/woolly/internal/memory"
)

func TestKindsAreSavedAndFiltered(t *testing.T) {
	s, store := trustServer(t)
	s.Project = "acme"
	ctx := context.Background()
	if _, err := s.callTool(ctx, "remember", json.RawMessage(`{"content":"Likes deploy summaries as bullet lists","type":"preference"}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.callTool(ctx, "remember", json.RawMessage(`{"content":"To deploy: run make build, then fly deploy","type":"procedure","scope":"personal"}`)); err != nil {
		t.Fatal(err)
	}
	mems, _ := store.List(false)
	for _, m := range mems {
		if m.Project != "" {
			t.Fatalf("a preference or procedure went to project memory: %+v", m)
		}
	}
	out, err := s.callTool(ctx, "recall", json.RawMessage(`{"query":"deploy","type":["procedure"]}`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "fly deploy") || strings.Contains(out, "bullet lists") {
		t.Fatalf("type filter:\n%s", out)
	}
}

func TestKindRoundTrips(t *testing.T) {
	_, store := trustServer(t)
	res, err := store.Remember(memory.RememberInput{Content: "Standup ran long; moved planning to Friday", Type: "episodes"})
	if err != nil {
		t.Fatal(err)
	}
	got, _ := store.Get(res.Memory.ID)
	if got.Type != memory.TypeEpisode {
		t.Fatalf("type = %q", got.Type)
	}
}
