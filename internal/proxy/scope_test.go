package proxy

import (
	"strings"
	"testing"

	"github.com/rjcuff/woolly/internal/memory"
)

// The proxy serves the user's own chat apps, so what an app built on Woolly
// keeps about its users, agents or sessions must never be injected.
func TestProxyLeavesOutAppScopedMemories(t *testing.T) {
	p, got, store := setup(t)
	for _, in := range []memory.RememberInput{
		{Content: "Alice lives in Paris", User: "alice", Source: memory.SourceAgent},
		{Content: "The support bot lives in Lisbon", Agent: "support-bot", Source: memory.SourceAgent},
		{Content: "This run lives in Tokyo", Session: "run-7", Source: memory.SourceAgent},
	} {
		if _, err := store.Remember(in); err != nil {
			t.Fatal(err)
		}
	}
	rec := post(t, p.Handler(), "/v1/chat/completions",
		`{"model":"m","messages":[{"role":"user","content":"where do I live? Paris Lisbon Tokyo"}]}`, nil)
	if rec.Code != 200 {
		t.Fatalf("status %d", rec.Code)
	}
	msgs := got.body["messages"].([]any)
	sys, _ := msgs[0].(map[string]any)["content"].(string)
	for _, leak := range []string{"Paris", "Lisbon", "Tokyo"} {
		if strings.Contains(sys, leak) {
			t.Fatalf("proxy injected an app-scoped memory (%s):\n%s", leak, sys)
		}
	}
	if !strings.Contains(sys, "Denver") {
		t.Fatalf("proxy lost the user's own memory:\n%s", sys)
	}
}
