package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/rjcuff/woolly/internal/memory"
)

func TestServerForOneUser(t *testing.T) {
	s, store := trustServer(t)
	s.User, s.Agent = "alice", "support-bot"
	ctx := context.Background()
	store.Remember(memory.RememberInput{Content: "I deploy to Vercel"})
	store.Remember(memory.RememberInput{Content: "Bob deploys to Render", User: "bob"})
	if _, err := s.callTool(ctx, "remember", json.RawMessage(`{"content":"Alice deploys to Fly","scope":"personal"}`)); err != nil {
		t.Fatal(err)
	}
	mems, _ := store.ListIn(memory.Scope{User: "alice"}, false)
	if len(mems) != 1 || mems[0].Agent != "support-bot" {
		t.Fatalf("alice's memories: %+v", mems)
	}
	for _, tool := range []string{"recall", "get_profile", "list_memories"} {
		out, err := s.callTool(ctx, tool, json.RawMessage(`{"query":"deploys"}`))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out, "Fly") || strings.Contains(out, "Vercel") || strings.Contains(out, "Render") {
			t.Fatalf("%s for alice:\n%s", tool, out)
		}
	}
}
