package main

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/rjcuff/woolly/internal/memory"
)

// captureStdout runs f and returns what it printed.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	done := make(chan string)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()
	defer func() { os.Stdout = old }()
	f()
	w.Close()
	os.Stdout = old
	return <-done
}

// The startup brief is for the user's own agents: what an app built on
// Woolly keeps about its users must stay out of it.
func TestContextBriefLeavesOutAppScopedMemories(t *testing.T) {
	t.Setenv("WOOLLY_HOME", t.TempDir())
	t.Setenv("WOOLLY_NO_EMBED", "1")
	a, err := open()
	if err != nil {
		t.Fatal(err)
	}
	for _, in := range []memory.RememberInput{
		{Content: "I prefer short answers"},
		{Content: "Alice prefers long answers", User: "alice", Source: memory.SourceAgent},
		{Content: "Support bot speaks French", Agent: "support-bot", Source: memory.SourceAgent},
		{Content: "Run seven uses dark mode", Session: "run-7", Source: memory.SourceAgent},
	} {
		if _, err := a.store.Remember(in); err != nil {
			t.Fatal(err)
		}
	}
	dir := t.TempDir() // not a repo, so no project or audit work
	out := captureStdout(t, func() {
		if err := cmdContext([]string{"--dir", dir}); err != nil {
			t.Errorf("context: %v", err)
		}
	})
	if !strings.Contains(out, "I prefer short answers") {
		t.Fatalf("brief lost the user's own memory:\n%s", out)
	}
	for _, leak := range []string{"Alice", "French", "dark mode"} {
		if strings.Contains(out, leak) {
			t.Fatalf("brief leaked an app-scoped memory (%s):\n%s", leak, out)
		}
	}
}

// With only app-scoped memories the brief prints nothing at all.
func TestContextBriefEmptyWhenOnlyAppMemories(t *testing.T) {
	t.Setenv("WOOLLY_HOME", t.TempDir())
	t.Setenv("WOOLLY_NO_EMBED", "1")
	a, err := open()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.store.Remember(memory.RememberInput{Content: "Alice prefers long answers", User: "alice", Source: memory.SourceAgent}); err != nil {
		t.Fatal(err)
	}
	out := captureStdout(t, func() { _ = cmdContext([]string{"--dir", t.TempDir()}) })
	if strings.Contains(out, "Alice") {
		t.Fatalf("brief leaked:\n%s", out)
	}
}
