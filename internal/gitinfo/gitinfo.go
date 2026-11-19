// Package gitinfo asks git which commit a repo is at and what changed since.
package gitinfo

import (
	"bufio"
	"bytes"
	"context"
	"errors"
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

// gitFor runs git with a longer limit, for searches through history.
func gitFor(dir string, limit time.Duration, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()
	out, err := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...).Output()
	return strings.TrimSpace(string(out)), err
}

// Present returns which of words appear as whole words in the repo's files
// right now, tracked or new, leaving out the paths in skip. It reports
// false when git couldn't finish, so nothing should be taken as missing.
func Present(dir string, words, skip []string) (map[string]bool, bool) {
	found := map[string]bool{}
	if len(words) == 0 {
		return found, true
	}
	args := []string{"grep", "--untracked", "-I", "-o", "-h", "-w", "-F"}
	for _, w := range words {
		args = append(args, "-e", w)
	}
	args = append(append(args, "--"), Exclude(skip)...)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...).Output()
	if ctx.Err() != nil {
		return found, false
	}
	// git grep exits 1 when nothing matched; anything else is a failure.
	if err != nil {
		var ee *exec.ExitError
		if !errors.As(err, &ee) || ee.ExitCode() != 1 {
			return found, false
		}
	}
	for _, w := range strings.Split(string(out), "\n") {
		if w = strings.TrimSpace(w); w != "" {
			found[w] = true
		}
	}
	return found, true
}

// Change is a commit that removed or renamed something.
type Change struct {
	Commit string // short hash
	Date   string // YYYY-MM-DD
	Path   string // the file it happened in, when there was one
	To     string // where a renamed file went
}

// Exclude turns paths into pathspecs that leave them out of a search.
func Exclude(paths []string) []string {
	out := []string{"."}
	for _, p := range paths {
		out = append(out, ":(exclude,literal)"+p)
	}
	return out
}

// LastChange finds the last commit in history whose diff changes how many
// times the extended regular expression re matches, within pathspec. When
// the thing re matches is gone now, that's the commit that removed it.
func LastChange(dir, re, since string, pathspec []string) (Change, bool) {
	args := []string{"log", "-1", "--format=%h%x09%cs", "--name-only", "--pickaxe-regex", "-S" + re}
	if since != "" {
		args = append(args, "--since="+since)
	}
	args = append(append(args, "--"), pathspec...)
	out, err := gitFor(dir, 8*time.Second, args...)
	if err != nil || out == "" {
		return Change{}, false
	}
	lines := strings.Split(out, "\n")
	f := strings.SplitN(lines[0], "\t", 2)
	c := Change{Commit: f[0]}
	if len(f) == 2 {
		c.Date = f[1]
	}
	for _, l := range lines[1:] {
		if l = strings.TrimSpace(l); l != "" {
			c.Path = l
			break
		}
	}
	return c, true
}

// PathGone finds the last commit that deleted or renamed path. It reports
// false when git never knew the path.
func PathGone(dir, path, since string) (Change, bool) {
	args := []string{"log", "-1", "--format=%h%x09%cs", "--diff-filter=D"}
	if since != "" {
		args = append(args, "--since="+since)
	}
	out, err := gitFor(dir, 8*time.Second, append(args, "--", path)...)
	if err != nil || out == "" {
		return Change{}, false
	}
	f := strings.SplitN(out, "\t", 2)
	c := Change{Commit: f[0], Path: path}
	if len(f) == 2 {
		c.Date = f[1]
	}
	// Limited to path, git sees a rename as a delete. Ask the whole commit.
	all, _ := gitFor(dir, 8*time.Second, "show", "-M", "--name-status", "--format=", c.Commit)
	for _, l := range strings.Split(all, "\n") {
		p := strings.Split(strings.TrimSpace(l), "\t")
		if len(p) != 3 || !strings.HasPrefix(p[0], "R") {
			continue
		}
		if p[1] == path {
			c.To = p[2]
			break
		}
		// A folder: its files moved together, keeping their names below it.
		if rest, ok := strings.CutPrefix(p[1], path+"/"); ok && strings.HasSuffix(p[2], "/"+rest) {
			c.To = strings.TrimSuffix(p[2], "/"+rest) + "/"
			break
		}
	}
	return c, true
}

// Added returns the lines a commit added to path.
func Added(dir, commit, path string) []string {
	out, err := git(dir, "show", "--format=", "-U0", commit, "--", path)
	if err != nil {
		return nil
	}
	var lines []string
	for _, l := range strings.Split(out, "\n") {
		if strings.HasPrefix(l, "+") && !strings.HasPrefix(l, "+++") {
			lines = append(lines, l[1:])
		}
	}
	return lines
}

// Root is the repo's top folder, or "" outside a repo.
func Root(dir string) string {
	out, err := git(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return ""
	}
	return filepath.Clean(out)
}

// IsAncestor reports whether commit is already in base's history.
func IsAncestor(dir, commit, base string) bool {
	_, err := git(dir, "merge-base", "--is-ancestor", commit, base)
	return err == nil
}

// FirstAdded returns when path was first committed, as YYYY-MM-DD, or "".
func FirstAdded(dir, path string) string {
	out, err := gitFor(dir, 8*time.Second, "log", "--diff-filter=A", "--format=%cs", "--", path)
	if err != nil || out == "" {
		return ""
	}
	lines := strings.Split(out, "\n")
	return lines[len(lines)-1]
}

// Removals reads the history since a date once, newest first, and finds the
// last commit that took each wanted identifier out of a line, within
// pathspec. It stops at deadline and then reports false, so names it didn't
// find may still have been removed earlier.
func Removals(dir, since string, want map[string]bool, pathspec []string, deadline time.Time) (map[string]Change, bool) {
	out := map[string]Change{}
	if len(want) == 0 {
		return out, true
	}
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	args := []string{"-C", dir, "log", "--no-color", "--no-merges", "-p", "-U0", "--no-renames", "--format=@@@%h %cs"}
	if since != "" {
		args = append(args, "--since="+since)
	}
	args = append(append(args, "--"), pathspec...)
	cmd := exec.CommandContext(ctx, "git", args...)
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return out, false
	}
	if err := cmd.Start(); err != nil {
		return out, false
	}
	sc := bufio.NewScanner(pipe)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	var cur Change
	skipFile := false
	for sc.Scan() {
		l := sc.Bytes()
		switch {
		case bytes.HasPrefix(l, []byte("@@@")):
			f := strings.SplitN(string(l[3:]), " ", 2)
			cur = Change{Commit: f[0]}
			if len(f) == 2 {
				cur.Date = f[1]
			}
		case bytes.HasPrefix(l, []byte("diff --git ")):
			// diff --git a/path b/path
			s := string(l)
			if i := strings.LastIndex(s, " b/"); i >= 0 {
				cur.Path = s[i+3:]
			}
			skipFile = false
		case bytes.HasPrefix(l, []byte("Binary files")):
			skipFile = true
		case len(l) > 0 && l[0] == '-' && !bytes.HasPrefix(l, []byte("---")) && !skipFile:
			for i := 1; i < len(l); {
				c := l[i]
				if !(c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9') {
					i++
					continue
				}
				j := i
				for j < len(l) && (l[j] == '_' || l[j] >= 'a' && l[j] <= 'z' || l[j] >= 'A' && l[j] <= 'Z' || l[j] >= '0' && l[j] <= '9') {
					j++
				}
				if id := string(l[i:j]); want[id] {
					if _, seen := out[id]; !seen {
						out[id] = cur
					}
				}
				i = j
			}
		}
	}
	err = cmd.Wait()
	return out, ctx.Err() == nil && err == nil
}
