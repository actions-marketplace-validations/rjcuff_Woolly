package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rjcuff/woolly/internal/audit"
)

// A pull request that renames a function CLAUDE.md names gets a comment
// about it, apart from the line that was already out of date.
func TestMarkdownAuditSplitsThisChangeFromBefore(t *testing.T) {
	dir := t.TempDir()
	git := func(args ...string) string {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	write := func(name, text string) {
		os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644)
	}
	git("init", "-q")
	write("a.go", "func oldHelper() {}\nfunc requireUser() {}\n")
	git("add", ".")
	git("commit", "-q", "-m", "one")
	write("a.go", "func requireUser() {}\n")
	write("CLAUDE.md", "- Use oldHelper for dates\n- Handlers call requireUser first\n")
	git("add", ".")
	git("commit", "-q", "-m", "drop oldHelper")
	base := git("rev-parse", "HEAD")
	write("a.go", "func requireSession() {}\n")
	git("commit", "-q", "-am", "rename")

	t.Setenv("HOME", t.TempDir())
	rep := audit.Run(dir, audit.Options{Env: audit.Env{}, Deadline: time.Now().Add(20 * time.Second)})
	md := markdownAudit(dir, rep, base)
	for _, want := range []string{"**1 line in agent files went stale.**", "`CLAUDE.md:2`", "requireUser was renamed to requireSession", "1 line was already out of date before this change", "oldHelper was removed from a.go"} {
		if !strings.Contains(md, want) {
			t.Errorf("missing %q in:\n%s", want, md)
		}
	}
	if s := auditSection(rep); !strings.Contains(s, "CLAUDE.md:1") || !strings.Contains(s, "CLAUDE.md:2") {
		t.Errorf("session brief:\n%s", s)
	}
}
