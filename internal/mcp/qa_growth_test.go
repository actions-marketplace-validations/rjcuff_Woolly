package mcp

// QA tests for S3 (woolly mcp --user/--agent), G1 kinds, G4 near copies on
// save and F5 list budgets, written against plans/memory-growth-and-fix.md.

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/rjcuff/woolly/internal/memory"
)

func call(t *testing.T, s *Server, tool, args string) (string, error) {
	t.Helper()
	return s.callTool(context.Background(), tool, json.RawMessage(args))
}

// A server started for one app user must not reach the user's own memories
// by ID: forget, resolve_conflict, link_memories and rate_memory all take an
// ID straight from the agent.
func TestQAUserServerCannotTouchPersonalMemoriesByID(t *testing.T) {
	s, store := trustServer(t)
	s.User = "alice"
	mine, _ := store.Remember(memory.RememberInput{Content: "I live in Austin"})
	other, _ := store.Remember(memory.RememberInput{Content: "Bob likes tea", User: "bob"})
	aliceRes, _ := store.Remember(memory.RememberInput{Content: "Alice likes coffee", User: "alice"})

	for _, id := range []string{mine.Memory.ID, other.Memory.ID} {
		if _, err := call(t, s, "rate_memory", `{"id":"`+id+`","helpful":false}`); err == nil {
			t.Errorf("rate_memory reached %s outside alice's scope", id)
		}
		if _, err := call(t, s, "link_memories", `{"id":"`+aliceRes.Memory.ID+`","to":"`+id+`"}`); err == nil {
			t.Errorf("link_memories reached %s outside alice's scope", id)
		}
		if _, err := call(t, s, "resolve_conflict", `{"keep_id":"`+id+`"}`); err == nil {
			t.Errorf("resolve_conflict reached %s outside alice's scope", id)
		}
		if _, err := call(t, s, "forget", `{"id":"`+id+`"}`); err == nil {
			t.Errorf("forget reached %s outside alice's scope", id)
		}
	}
	if _, err := store.Get(mine.Memory.ID); err != nil {
		t.Errorf("the user's own memory was forgotten by alice's server: %v", err)
	}
	// Alice's own memory is still hers to forget.
	if _, err := call(t, s, "forget", `{"id":"`+aliceRes.Memory.ID+`"}`); err != nil {
		t.Fatal(err)
	}
}

// The default server (the user's own) should not forget an app user's memory
// by ID either: F2 keeps app users' memories out of every read path.
func TestQAPersonalServerCannotForgetAppUserMemory(t *testing.T) {
	s, store := trustServer(t)
	alice, _ := store.Remember(memory.RememberInput{Content: "Alice likes coffee", User: "alice"})
	if _, err := call(t, s, "forget", `{"id":"`+alice.Memory.ID+`"}`); err == nil {
		t.Fatal("the user's MCP server forgot an app user's memory by ID")
	}
}

// handoff and get_profile on a server for alice must stay in alice's scope.
func TestQAUserServerHandoffStaysScoped(t *testing.T) {
	s, store := trustServer(t)
	s.Project = "acme"
	// The user's own handoff for this project.
	store.Remember(memory.RememberInput{Content: "Personal: was refactoring the billing module", Key: "handoff", Project: "acme", Source: memory.SourceAgent})
	alice := &Server{Store: store, Engine: s.Engine, Version: "test", Project: "acme", User: "alice"}
	out, err := alice.callTool(context.Background(), "get_profile", json.RawMessage(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "billing module") {
		t.Errorf("alice's server shows the user's own handoff:\n%s", out)
	}
	// A scoped server refuses handoffs, which belong to the user's own sessions.
	if _, err := alice.callTool(context.Background(), "handoff", json.RawMessage(`{"note":"Alice: fixed the login bug, next is SSO"}`)); err == nil {
		t.Fatal("a scoped server saved a handoff")
	}
	mems, _ := store.List(false)
	for _, m := range mems {
		if strings.Contains(m.Content, "Alice: fixed the login bug") {
			t.Errorf("alice's handoff was saved: %+v", m)
		}
	}
}

func TestQAUserServerDecideAndRememberAreScoped(t *testing.T) {
	s, store := trustServer(t)
	s.Agent = "support-bot"
	if _, err := call(t, s, "decide", `{"decision":"Use SMS for 2FA","why":"users asked"}`); err != nil {
		t.Fatal(err)
	}
	if _, err := call(t, s, "remember", `{"content":"Customers prefer email receipts","scope":"personal"}`); err != nil {
		t.Fatal(err)
	}
	mems, _ := store.List(false)
	for _, m := range mems {
		if m.Agent != "support-bot" {
			t.Fatalf("saved outside the agent scope: %+v", m)
		}
	}
	// The user's own MCP server doesn't see them.
	me, _ := trustServerOn(store)
	out, _ := me.callTool(context.Background(), "list_memories", json.RawMessage(`{}`))
	if strings.Contains(out, "SMS") || strings.Contains(out, "email receipts") {
		t.Fatalf("the user's own server lists the agent's memories:\n%s", out)
	}
}

func trustServerOn(store *memory.Store) (*Server, *memory.Store) {
	s := &Server{Store: store, Version: "test"}
	return s, store
}

func TestQAPreferenceInARepoGoesPersonal(t *testing.T) {
	s, store := trustServer(t)
	s.Project = "acme"
	for _, typ := range []string{"preference", "procedure", "episode"} {
		if _, err := call(t, s, "remember", `{"content":"The user `+typ+` note about how they work","type":"`+typ+`"}`); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := call(t, s, "remember", `{"content":"Builds use make, never bazel","type":"convention"}`); err != nil {
		t.Fatal(err)
	}
	mems, _ := store.List(false)
	for _, m := range mems {
		project := memory.ProjectType(m.Type)
		if project != (m.Project == "acme") {
			t.Fatalf("%s memory went to project %q", m.Type, m.Project)
		}
	}
}

func TestQARecallUnknownTypeFilter(t *testing.T) {
	s, store := trustServer(t)
	store.Remember(memory.RememberInput{Content: "Deploy by running make ship", Type: "procedure"})
	store.Remember(memory.RememberInput{Content: "Likes deploy logs short", Type: "preference"})
	out, err := call(t, s, "recall", `{"query":"deploy","type":["facts"]}`)
	if err != nil {
		return // refusing an unknown type is fine
	}
	if strings.Contains(out, "make ship") || strings.Contains(out, "logs short") {
		t.Fatalf(`recall with type ["facts"] ignored the filter and returned everything:\n%s`, out)
	}
}

// G4: a negation is a new fact, not "Already known".
func TestQARememberNegationIsSaved(t *testing.T) {
	s, store := trustServer(t)
	if _, err := call(t, s, "remember", `{"content":"The team deploys on Fridays","scope":"personal"}`); err != nil {
		t.Fatal(err)
	}
	out, err := call(t, s, "remember", `{"content":"The team does not deploy on Fridays","scope":"personal"}`)
	if err != nil {
		t.Fatal(err)
	}
	mems, _ := store.List(false)
	if strings.Contains(out, "Already known") || len(mems) != 2 {
		t.Fatalf("the correction was dropped as a copy: %q (%d memories)", out, len(mems))
	}
}

func TestQAKeyedRememberBypassesNearCopy(t *testing.T) {
	s, store := trustServer(t)
	call(t, s, "remember", `{"content":"Prefers tabs over spaces","key":"style.indent","scope":"personal"}`)
	out, err := call(t, s, "remember", `{"content":"Prefers tabs over spaces!","key":"style.indent","scope":"personal"}`)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "Already known") {
		t.Fatalf("a keyed fact went through the near-copy check: %s", out)
	}
	mems, _ := store.List(false)
	if len(mems) != 1 {
		t.Fatalf("%d current facts for one key", len(mems))
	}
}

// A near copy in another scope (an app user) or another project must not
// swallow a new fact.
func TestQANearCopyStaysInScope(t *testing.T) {
	s, store := trustServer(t)
	s.Project = "acme"
	store.Remember(memory.RememberInput{Content: "Prefers dark mode", User: "alice"})
	store.Remember(memory.RememberInput{Content: "Prefers dark mode", Project: "other"})
	out, err := call(t, s, "remember", `{"content":"Prefers dark mode","scope":"personal"}`)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "Already known") {
		t.Fatalf("matched a copy outside the scope: %s", out)
	}
}

// A near copy of the user's fact said again by an agent adds tags but never
// lowers the source.
func TestQANearCopyKeepsSource(t *testing.T) {
	s, store := trustServer(t)
	mine, _ := store.Remember(memory.RememberInput{Content: "Prefers short answers"})
	out, _ := call(t, s, "remember", `{"content":"prefers short answers.","scope":"personal","tags":["style"]}`)
	if !strings.Contains(out, "Already known") {
		t.Fatalf("not matched: %s", out)
	}
	got, _ := store.Get(mine.Memory.ID)
	if got.Source != memory.SourceUser {
		t.Fatalf("source = %s", got.Source)
	}
	tagged := false
	for _, tg := range got.Tags {
		tagged = tagged || tg == "style"
	}
	if !tagged {
		t.Fatalf("tags not added: %v", got.Tags)
	}
}

func TestQAListMemoriesLimitEdges(t *testing.T) {
	s, store := trustServer(t)
	for i := 0; i < 55; i++ {
		store.Remember(memory.RememberInput{Content: fmt.Sprintf("Memory number %d", i)})
	}
	out, _ := call(t, s, "list_memories", `{}`)
	if !strings.Contains(out, "(5 more.") {
		t.Fatalf("default limit 50 should leave 5:\n%s", out[len(out)-200:])
	}
	out, _ = call(t, s, "list_memories", `{"limit":55}`)
	if strings.Contains(out, "more.") {
		t.Fatal("exactly at the limit says there are more")
	}
	out, _ = call(t, s, "list_memories", `{"limit":-3}`)
	if !strings.Contains(out, "(5 more.") {
		t.Fatal("a negative limit should fall back to 50")
	}
}

func TestQARecallMaxCharsLogsOnlyShown(t *testing.T) {
	s, store := trustServer(t)
	var ids []string
	for i := 0; i < 5; i++ {
		r, _ := store.Remember(memory.RememberInput{Content: fmt.Sprintf("Kafka topic %d %s", i, strings.Repeat("x", 150))})
		ids = append(ids, r.Memory.ID)
	}
	out, err := call(t, s, "recall", `{"query":"kafka topic","limit":5,"max_chars":200}`)
	if err != nil {
		t.Fatal(err)
	}
	shown := strings.Count(out, "Kafka topic")
	used := 0
	for id, u := range store.Usage() {
		_ = id
		used += u.Uses
	}
	if shown != 1 || used != shown {
		t.Fatalf("shown %d, touched %d:\n%s", shown, used, out)
	}
}

func TestQARateMemoryMissing(t *testing.T) {
	s, _ := trustServer(t)
	if _, err := call(t, s, "rate_memory", `{"id":"nope","helpful":true}`); err == nil {
		t.Fatal("rated a missing memory")
	}
	if _, err := call(t, s, "rate_memory", `{"id":"","helpful":true}`); err == nil {
		t.Fatal("rated an empty id")
	}
}

func TestQALinkMemoriesErrors(t *testing.T) {
	s, store := trustServer(t)
	a, _ := store.Remember(memory.RememberInput{Content: "A"})
	if _, err := call(t, s, "link_memories", `{"id":"`+a.Memory.ID+`","to":"`+a.Memory.ID+`"}`); err == nil {
		t.Fatal("linked to itself")
	}
	if _, err := call(t, s, "link_memories", `{"id":"`+a.Memory.ID+`"}`); err == nil {
		t.Fatal("linked to nothing")
	}
}
