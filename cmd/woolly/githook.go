package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Woolly's lines in a repo's post-commit hook sit between these markers, so
// they can be updated or removed without touching anything else there.
const (
	hookStart = "# woolly:start (added by woolly init --hooks; woolly init --no-hooks removes it)"
	hookEnd   = "# woolly:end"
)

// postCommitPath is where git looks for the repo's post-commit hook,
// honoring core.hooksPath.
func postCommitPath(repo string) (string, error) {
	out, err := exec.Command("git", "-C", repo, "rev-parse", "--git-path", "hooks").Output()
	if err != nil {
		return "", err
	}
	dir := strings.TrimSpace(string(out))
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(repo, dir)
	}
	return filepath.Join(dir, "post-commit"), nil
}

// hookBlock runs a check in the background after each commit, so the
// commit returns right away.
func hookBlock() string {
	bin := "woolly"
	if exe, err := os.Executable(); err == nil && !strings.Contains(exe, "go-build") {
		if p, err := filepath.EvalSymlinks(exe); err == nil {
			exe = p
		}
		bin = "'" + strings.ReplaceAll(filepath.ToSlash(exe), "'", `'\''`) + "'"
	}
	return hookStart + "\n" + bin + " check --hook >/dev/null 2>&1 &\n" + hookEnd + "\n"
}

// withoutBlock removes Woolly's lines from a hook.
func withoutBlock(text string) string {
	i := strings.Index(text, hookStart)
	if i < 0 {
		return text
	}
	j := strings.Index(text[i:], hookEnd)
	if j < 0 {
		return text
	}
	end := i + j + len(hookEnd)
	if end < len(text) && text[end] == '\n' {
		end++
	}
	return text[:i] + text[end:]
}

// installHook adds Woolly's check to the repo's post-commit hook, keeping
// whatever else is there. It reports whether the hook changed.
func installHook(repo string) (bool, error) {
	p, err := postCommitPath(repo)
	if err != nil {
		return false, err
	}
	old, err := os.ReadFile(p)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	text := withoutBlock(string(old))
	if strings.TrimSpace(text) == "" {
		text = "#!/bin/sh\n"
	}
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	text += hookBlock()
	if text == string(old) {
		return false, nil
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return false, err
	}
	if err := os.WriteFile(p, []byte(text), 0o755); err != nil {
		return false, err
	}
	return true, os.Chmod(p, 0o755)
}

// removeHook takes Woolly's lines out of the post-commit hook, and deletes
// the hook if nothing else is left in it.
func removeHook(repo string) (bool, error) {
	p, err := postCommitPath(repo)
	if err != nil {
		return false, err
	}
	old, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	text := withoutBlock(string(old))
	if text == string(old) {
		return false, nil
	}
	if strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(text), "#!/bin/sh")) == "" {
		return true, os.Remove(p)
	}
	return true, os.WriteFile(p, []byte(text), 0o755)
}

// hasHook reports whether the repo checks memory after each commit.
func hasHook(repo string) bool {
	p, err := postCommitPath(repo)
	if err != nil {
		return false
	}
	b, err := os.ReadFile(p)
	return err == nil && strings.Contains(string(b), hookStart)
}
