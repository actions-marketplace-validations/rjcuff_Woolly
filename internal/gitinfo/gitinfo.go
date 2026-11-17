// Package gitinfo asks git which commit a repo is at and what changed since.
package gitinfo

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func git(dir string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...).Output()
	return strings.TrimSpace(string(out)), err
}

// Head returns the repo's current commit, or "" if dir isn't a git repo.
func Head(dir string) string {
	h, err := git(dir, "rev-parse", "--short=12", "HEAD")
	if err != nil {
		return ""
	}
	return h
}

var pathLike = regexp.MustCompile("`?([A-Za-z0-9_.\\-]+(?:/[A-Za-z0-9_.\\-]+)*\\.[A-Za-z0-9]{1,8}|[A-Za-z0-9_\\-]+(?:/[A-Za-z0-9_.\\-]+)+)`?")

// Paths returns the files and folders text mentions that exist in the repo.
func Paths(dir, text string) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range pathLike.FindAllStringSubmatch(text, 20) {
		p := strings.Trim(m[1], "./")
		if p == "" || seen[p] || strings.Contains(p, "..") {
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, p)); err != nil {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	return out
}

// Changes counts the commits since commit that touched any of paths.
func Changes(dir, commit string, paths []string) int {
	if commit == "" || len(paths) == 0 {
		return 0
	}
	out, err := git(dir, append([]string{"rev-list", "--count", commit + "..HEAD", "--"}, paths...)...)
	if err != nil {
		return 0
	}
	n := 0
	for _, c := range out {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}

// Show returns a file as it was at commit, if git has it.
func Show(dir, commit, path string) ([]byte, bool) {
	if commit == "" {
		return nil, false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "git", "-C", dir, "show", commit+":"+path).Output()
	if err != nil {
		return nil, false
	}
	return out, true
}

// LastTouch returns the short hash of the last commit that changed path,
// including the one that deleted it.
func LastTouch(dir, path string) string {
	out, err := git(dir, "log", "-1", "--format=%h", "--", path)
	if err != nil {
		return ""
	}
	return out
}

// Removed returns the short hash of the last commit that added or removed
// text in path (git log -S), or "" if git can't tell.
func Removed(dir, text, path string) string {
	out, err := git(dir, "log", "-1", "--format=%h", "-S"+text, "--", path)
	if err != nil {
		return ""
	}
	return out
}

// RenamedTo returns where git thinks path moved to since commit, or "".
func RenamedTo(dir, commit, path string) string {
	if commit == "" {
		return ""
	}
	out, err := git(dir, "diff", "--name-status", "-M", commit, "--")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(out, "\n") {
		f := strings.Split(line, "\t")
		if len(f) == 3 && strings.HasPrefix(f[0], "R") && f[1] == path {
			return f[2]
		}
	}
	return ""
}

// Files lists the files git tracks in the repo, plus new ones it would.
func Files(dir string) []string {
	out, err := git(dir, "ls-files", "--cached", "--others", "--exclude-standard")
	if err != nil || out == "" {
		return nil
	}
	return strings.Split(out, "\n")
}

// Dirty reports whether path has changes that aren't committed.
func Dirty(dir, path string) bool {
	out, err := git(dir, "status", "--porcelain", "--", path)
	return err == nil && out != ""
}

// Since lists the commits after base that touched path, oldest first, at
// most 100.
func Since(dir, base, path string) []string {
	args := []string{"rev-list", "--reverse", "--abbrev-commit", "HEAD", "--", path}
	if base != "" {
		args = []string{"rev-list", "--reverse", "--abbrev-commit", base + "..HEAD", "--", path}
	}
	out, err := git(dir, args...)
	if err != nil || out == "" {
		return nil
	}
	list := strings.Split(out, "\n")
	if len(list) > 100 {
		list = list[len(list)-100:]
	}
	return list
}
