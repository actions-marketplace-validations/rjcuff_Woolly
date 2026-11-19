package mcp

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rjcuff/woolly/internal/memory"
	"github.com/rjcuff/woolly/internal/recall"
)

// An agent asks whether CLAUDE.md is still right after a rename.
func TestCheckMemoryFiles(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CLAUDE_CONFIG_DIR", t.TempDir())
	t.Setenv("CODEX_HOME", t.TempDir())
	git := func(args ...string) {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	git("init", "-q")
	os.WriteFile(filepath.Join(dir, "auth.go"), []byte("func requireUser() {}\n"), 0o644)
	git("add", ".")
	git("commit", "-q", "-m", "one")
	os.WriteFile(filepath.Join(dir, "auth.go"), []byte("func requireSession() {}\n"), 0o644)
	git("commit", "-q", "-am", "two")
	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("- Every handler calls requireUser first\n"), 0o644)

	st, err := memory.Open(filepath.Join(t.TempDir(), "memories"))
	if err != nil {
		t.Fatal(err)
	}
	s := &Server{Store: st, Engine: recall.New(st, nil, ""), Version: "test", Root: dir, Project: "acme"}
	in := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"check_memory_files","arguments":{}}}` + "\n"
	var out bytes.Buffer
	s.Serve(context.Background(), strings.NewReader(in), &out)
	got := out.String()
	for _, want := range []string{"CLAUDE.md:1", "requireUser was renamed to requireSession in auth.go", "Fix: use requireSession"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %s", want, got)
		}
	}
}
