package gitinfo

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func run(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.email=t@t", "-c", "user.name=t"}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
}

func TestChangesSinceCommit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	run(t, dir, "init", "-q")
	os.MkdirAll(filepath.Join(dir, "api"), 0o755)
	os.WriteFile(filepath.Join(dir, "api", "ratelimit.go"), []byte("v1"), 0o644)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("hi"), 0o644)
	run(t, dir, "add", ".")
	run(t, dir, "commit", "-qm", "one")
	start := Head(dir)
	if start == "" {
		t.Fatal("no head")
	}
	paths := Paths(dir, "Rate limits live in `api/ratelimit.go`, see notes.txt and api/")
	if len(paths) != 1 || paths[0] != "api/ratelimit.go" {
		t.Fatalf("paths %v", paths)
	}
	if n := Changes(dir, start, []string{"api/ratelimit.go"}); n != 0 {
		t.Fatalf("no changes yet, got %d", n)
	}
	os.WriteFile(filepath.Join(dir, "api", "ratelimit.go"), []byte("v2"), 0o644)
	run(t, dir, "commit", "-qam", "two")
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello"), 0o644)
	run(t, dir, "commit", "-qam", "three")
	if n := Changes(dir, start, []string{"api/ratelimit.go"}); n != 1 {
		t.Fatalf("want 1 change, got %d", n)
	}
	if Head(t.TempDir()) != "" {
		t.Fatal("not a repo")
	}
}

func TestTrackedAndShortCommit(t *testing.T) {
	dir := t.TempDir()
	run(t, dir, "init", "-q")
	for name, text := range map[string]string{".gitignore": "CLAUDE.local.md\n", "CLAUDE.md": "x\n", "CLAUDE.local.md": "mine\n", "notes é.md": "y\n"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	run(t, dir, "add", ".gitignore", "CLAUDE.md", "notes é.md")
	run(t, dir, "commit", "-q", "-m", "one")
	got := map[string]bool{}
	for _, f := range Tracked(dir) {
		got[f] = true
	}
	if !got["CLAUDE.md"] || !got["notes é.md"] || got["CLAUDE.local.md"] {
		t.Errorf("Tracked = %v", got)
	}
	if h := ShortCommit(dir, "HEAD"); len(h) != 12 {
		t.Errorf("ShortCommit(HEAD) = %q", h)
	}
	for _, bad := range []string{"", "nope", "-c", "--output=/tmp/x"} {
		if h := ShortCommit(dir, bad); h != "" {
			t.Errorf("ShortCommit(%q) = %q, want empty", bad, h)
		}
	}
}
