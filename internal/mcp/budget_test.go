package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/rjcuff/woolly/internal/memory"
)

func TestListMemoriesHasALimit(t *testing.T) {
	s, store := trustServer(t)
	for i := 0; i < 60; i++ {
		store.Remember(memory.RememberInput{Content: fmt.Sprintf("Fact number %d", i)})
	}
	out, err := s.callTool(context.Background(), "list_memories", json.RawMessage(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(out, "Fact number"); n != 50 || !strings.Contains(out, "10 more") {
		t.Fatalf("%d facts listed:\n%s", n, out)
	}
	out, _ = s.callTool(context.Background(), "list_memories", json.RawMessage(`{"limit":5}`))
	if n := strings.Count(out, "Fact number"); n != 5 {
		t.Fatalf("limit 5 listed %d", n)
	}
}

func TestRecallMaxChars(t *testing.T) {
	s, store := trustServer(t)
	for i := 0; i < 10; i++ {
		store.Remember(memory.RememberInput{Content: fmt.Sprintf("Deploy note %d: %s", i, strings.Repeat("x", 80))})
	}
	out, err := s.callTool(context.Background(), "recall", json.RawMessage(`{"query":"deploy note","limit":10,"max_chars":300}`))
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(out, "Deploy note"); n == 0 || n > 3 {
		t.Fatalf("%d notes for 300 characters:\n%s", n, out)
	}
}
