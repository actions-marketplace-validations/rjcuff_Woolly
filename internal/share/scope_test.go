package share

import (
	"strings"
	"testing"

	"github.com/rjcuff/woolly/internal/memory"
)

func TestAppUsersMemoriesNeverGoInTheTeamFile(t *testing.T) {
	r := repo(t)
	me := store(t)
	me.Remember(memory.RememberInput{Content: "acme uses Postgres 17", Project: "acme"})
	me.Remember(memory.RememberInput{Content: "Alice deploys acme to Fly.io", Project: "acme", User: "alice", Source: memory.SourceAgent})
	if _, err := r.Init(me, Options{}); err != nil {
		t.Fatal(err)
	}
	me.Remember(memory.RememberInput{Content: "Bob prefers dark mode in acme", Project: "acme", Agent: "support-bot", Source: memory.SourceAgent})
	if _, err := r.Sync(me); err != nil {
		t.Fatal(err)
	}
	text := read(t, r)
	if !strings.Contains(text, "Postgres") {
		t.Fatalf("project note missing:\n%s", text)
	}
	for _, never := range []string{"Alice", "Bob"} {
		if strings.Contains(text, never) {
			t.Fatalf("%q must not be in the repo:\n%s", never, text)
		}
	}
}
