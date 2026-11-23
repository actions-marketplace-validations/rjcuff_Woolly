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

// A repo whose CLAUDE.md names one function removed before base and one
// renamed after it, plus a Claude Code memory note outside the repo.
func annotatedRepo(t *testing.T) (dir, base string, rep audit.Report) {
	t.Helper()
	dir = t.TempDir()
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
	write("CLAUDE.md", "- Use oldHelper for dates, then: check\n- Handlers call requireUser first\n")
	git("add", ".")
	git("commit", "-q", "-m", "drop oldHelper")
	base = git("rev-parse", "HEAD")
	write("a.go", "func requireSession() {}\n")
	git("commit", "-q", "-am", "rename")

	t.Setenv("HOME", t.TempDir())
	rep = audit.Run(dir, audit.Options{Env: audit.Env{}, Deadline: time.Now().Add(20 * time.Second)})
	rep.Findings = append(rep.Findings, audit.Finding{
		Line:   audit.Line{Source: "Claude Code memory", File: "Claude memory/notes.md", N: 3, Text: "use requireUser", Path: "/home/me/.claude/projects/x/memory/notes.md"},
		What:   "requireUser",
		Reason: "requireUser was renamed to requireSession",
	})
	return dir, base, rep
}

func TestGitHubAnnotations(t *testing.T) {
	dir, base, rep := annotatedRepo(t)
	out := githubFindings(dir, rep, base)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 annotations (memory outside the repo left out), got:\n%s", out)
	}
	if !strings.HasPrefix(lines[0], "::warning file=CLAUDE.md,line=2,title=Stale agent note::requireUser was renamed to requireSession in ") {
		t.Errorf("new finding:\n%s", lines[0])
	}
	if !strings.Contains(lines[0], "woolly check --fix to change it to requireSession") {
		t.Errorf("no fix hint:\n%s", lines[0])
	}
	if !strings.HasPrefix(lines[1], "::notice file=CLAUDE.md,line=1,title=Agent note already stale::oldHelper was removed") {
		t.Errorf("finding from before base:\n%s", lines[1])
	}
}

func TestGitHubEscaping(t *testing.T) {
	if got := ghProp("a:b,c%\n"); got != "a%3Ab%2Cc%25%0A" {
		t.Errorf("ghProp = %q", got)
	}
	if got := ghData("50%\r\nnext ::warning::x"); got != "50%25%0D%0Anext ::warning::x" {
		t.Errorf("ghData = %q", got)
	}
}

func TestCompactFindings(t *testing.T) {
	_, _, rep := annotatedRepo(t)
	out := compactFindings(rep)
	for _, want := range []string{
		"CLAUDE.md:2: requireUser was renamed to requireSession in ",
		"CLAUDE.md:1: oldHelper was removed from a.go",
		"/home/me/.claude/projects/x/memory/notes.md:3: requireUser was renamed",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestFormatFlagChecks(t *testing.T) {
	for args, want := range map[string]string{
		"--format=xml":                      "must be text, compact or github",
		"--format=github --json":            "can't be combined",
		"--format=compact --markdown":       "can't be combined",
		"--format=github --no-memory-files": "--no-memory-files",
	} {
		err := cmdCheck(strings.Fields(args))
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: err = %v", args, err)
		}
	}
}
