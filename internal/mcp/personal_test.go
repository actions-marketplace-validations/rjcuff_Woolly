package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/rjcuff/woolly/internal/memory"
)

func TestAppUsersMemoriesStayOutOfPersonalTools(t *testing.T) {
	s, store := trustServer(t)
	if _, err := store.Remember(memory.RememberInput{Content: "Alice deploys to Fly.io", User: "alice", Source: memory.SourceAgent}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Remember(memory.RememberInput{Content: "I deploy to Vercel"}); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct{ tool, args string }{
		{"recall", `{"query":"deploy"}`},
		{"get_profile", `{}`},
		{"list_memories", `{}`},
	} {
		out, err := s.callTool(context.Background(), c.tool, json.RawMessage(c.args))
		if err != nil {
			t.Fatalf("%s: %v", c.tool, err)
		}
		if strings.Contains(out, "Alice") {
			t.Fatalf("%s leaked an app user's memory:\n%s", c.tool, out)
		}
	}
	out, _ := s.callTool(context.Background(), "recall", json.RawMessage(`{"query":"deploy"}`))
	if !strings.Contains(out, "Vercel") {
		t.Fatalf("recall lost the user's own memory:\n%s", out)
	}
}
