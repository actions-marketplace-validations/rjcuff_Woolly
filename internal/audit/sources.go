// Package audit checks the memory a repo's agents already read, like
// CLAUDE.md, AGENTS.md, Cursor rules, Claude Code's auto memory, Codex
// memories and claude-mem, against the code. It flags lines that name a
// function, file or command the code no longer has. It needs no setup and
// no model: just git.
package audit

import (
	"github.com/rjcuff/woolly/internal/gitinfo"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Line is one line of memory an agent reads.
type Line struct {
	Source string `json:"source"` // which tool keeps it, like "CLAUDE.md" or "Claude Code memory"
	File   string `json:"file"`   // where it is, shown to people
	N      int    `json:"line"`   // its line number, or claude-mem's observation id
	Text   string `json:"text"`
	Path   string `json:"-"` // the file on disk, when it's a file
}

// File is one memory file and how many lines were read from it.
type File struct {
	Source string `json:"source"`
	File   string `json:"file"`
	Lines  int    `json:"lines"`
}

// repoFiles are the agent files read from the repo itself, relative to it.
var repoFiles = []string{"CLAUDE.md", "CLAUDE.local.md", "AGENTS.md", "AGENTS.local.md", "AGENT.md", "GEMINI.md", "WARP.md", ".cursorrules", ".windsurfrules", ".clinerules", ".github/copilot-instructions.md"}

// ruleGlobs are folders of agent rules, skills, commands and prompts that
// agents load as instructions, relative to the repo.
var ruleGlobs = []string{
	".cursor/rules/*.md", ".cursor/rules/*.mdc",
	".claude/rules/*.md", ".claude/rules/*/*.md",
	".claude/commands/*.md", ".claude/commands/*/*.md",
	".claude/agents/*.md",
	".claude/skills/*/SKILL.md",
	".github/instructions/*.instructions.md",
	".windsurf/rules/*.md",
	".clinerules/*.md",
	".kiro/steering/*.md",
}

// maxRuleFiles caps how many files ruleGlobs and imports add, so a repo
// with hundreds of prompts doesn't slow every check down.
const maxRuleFiles = 200

// maxImportDepth is how many hops of @imports are followed, like Claude
// Code does.
const maxImportDepth = 5

// Env says where each tool keeps its memory. Tests point it at temporary
// folders.
type Env struct {
	Home      string // the user's home folder
	ClaudeDir string // CLAUDE_CONFIG_DIR, or ~/.claude
	CodexDir  string // CODEX_HOME, or ~/.codex
	ClaudeMem string // claude-mem's database
}

// DefaultEnv reads the folders from the environment.
func DefaultEnv() Env {
	home, _ := os.UserHomeDir()
	e := Env{Home: home, ClaudeDir: os.Getenv("CLAUDE_CONFIG_DIR"), CodexDir: os.Getenv("CODEX_HOME")}
	if e.ClaudeDir == "" && home != "" {
		e.ClaudeDir = filepath.Join(home, ".claude")
	}
	if e.CodexDir == "" && home != "" {
		e.CodexDir = filepath.Join(home, ".codex")
	}
	if home != "" {
		e.ClaudeMem = filepath.Join(home, ".claude-mem", "claude-mem.db")
	}
	return e
}

const maxRead = 512 << 10

// Read collects every memory line the repo's agents read, and the repo
// files they came from (so a search can leave them out). also lists other
// spellings of the repo's path, like one before symlinks were resolved,
// that tools outside the repo may have saved memory under.
func Read(dir string, env Env, also ...string) (lines []Line, files []File, inRepo []string) {
	dirs := []string{dir}
	for _, d := range also {
		if d != "" && filepath.Clean(d) != filepath.Clean(dir) {
			dirs = append(dirs, filepath.Clean(d))
		}
	}
	add := func(source, shown string, ls []Line) {
		if len(ls) == 0 {
			return
		}
		lines = append(lines, ls...)
		files = append(files, File{Source: source, File: shown, Lines: len(ls)})
	}

	names := append([]string{}, repoFiles...)
	globbed := 0
	for _, pat := range ruleGlobs {
		found, err := filepath.Glob(filepath.Join(dir, filepath.FromSlash(pat)))
		if err != nil {
			continue
		}
		sort.Strings(found)
		for _, p := range found {
			if globbed >= maxRuleFiles {
				break
			}
			if rel, err := filepath.Rel(dir, p); err == nil {
				names = append(names, filepath.ToSlash(rel))
				globbed++
			}
		}
	}
	// Codex, Claude Code and others also read these files in subfolders.
	n := 0
	for _, f := range gitinfo.Files(dir) {
		base := f[strings.LastIndex(f, "/")+1:]
		if f == base || base != "AGENTS.md" && base != "CLAUDE.md" && base != "GEMINI.md" || strings.Contains(f, "node_modules/") {
			continue
		}
		if n++; n > 50 {
			break
		}
		names = append(names, f)
	}
	// Files that CLAUDE.md pulls in with @path are read too, as Claude Code
	// does, up to maxImportDepth hops.
	seenName := map[string]bool{}
	for _, n := range names {
		seenName[n] = true
	}
	depth := map[string]int{}
	for i := 0; i < len(names); i++ {
		name := names[i]
		b, ok := readSmall(filepath.Join(dir, filepath.FromSlash(name)))
		if !ok {
			continue
		}
		inRepo = append(inRepo, name)
		add(sourceOf(name), name, splitLines(sourceOf(name), name, filepath.Join(dir, filepath.FromSlash(name)), string(b)))
		base := name[strings.LastIndex(name, "/")+1:]
		if base != "CLAUDE.md" && base != "CLAUDE.local.md" && depth[name] == 0 {
			continue
		}
		if depth[name] >= maxImportDepth {
			continue
		}
		for _, imp := range imports(name, string(b)) {
			if seenName[imp] || globbed >= maxRuleFiles {
				continue
			}
			seenName[imp] = true
			depth[imp] = depth[name] + 1
			names = append(names, imp)
			globbed++
		}
	}
	// Agent settings, skills and prompts aren't the code the notes are about.
	inRepo = append(inRepo, ".woolly", ".agents", ".claude", ".cursor", ".codex", ".gemini", ".windsurf", ".clinerules", ".kiro", ".github/instructions")

	// Claude Code's auto memory for this folder.
	if env.ClaudeDir != "" {
		var paths []string
		seen := map[string]bool{}
		for _, d := range dirs {
			if k := claudeKey(d); !seen[k] {
				seen[k] = true
				found, _ := filepath.Glob(filepath.Join(env.ClaudeDir, "projects", k, "memory", "*.md"))
				sort.Strings(found)
				paths = append(paths, found...)
			}
		}
		for _, p := range paths {
			b, ok := readSmall(p)
			if !ok {
				continue
			}
			shown := "Claude memory/" + filepath.Base(p)
			add("Claude Code memory", shown, splitLines("Claude Code memory", shown, p, string(b)))
		}
	}

	// Codex memories that mention this repo.
	if env.CodexDir != "" {
		root := filepath.Join(env.CodexDir, "memories")
		_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(p, ".md") {
				return nil
			}
			b, ok := readSmall(p)
			if !ok || !mentionsRepo(string(b), dirs) {
				return nil
			}
			rel, _ := filepath.Rel(root, p)
			shown := "Codex memory/" + filepath.ToSlash(rel)
			add("Codex memory", shown, splitLines("Codex memory", shown, p, string(b)))
			return nil
		})
	}

	// claude-mem's observations for this project.
	if obs := claudeMem(env.ClaudeMem, filepath.Base(dir)); len(obs) > 0 {
		lines = append(lines, obs...)
		files = append(files, File{Source: "claude-mem", File: "claude-mem", Lines: len(obs)})
	}
	return lines, files, inRepo
}

func readSmall(p string) ([]byte, bool) {
	st, err := os.Stat(p)
	if err != nil || !st.Mode().IsRegular() || st.Size() > maxRead {
		return nil, false
	}
	b, err := os.ReadFile(p)
	return b, err == nil
}

func sourceOf(name string) string {
	switch {
	case strings.HasPrefix(name, ".cursor/") || name == ".cursorrules":
		return "Cursor rules"
	case strings.HasPrefix(name, ".claude/rules/"):
		return "Claude Code rules"
	case strings.HasPrefix(name, ".claude/commands/"):
		return "Claude Code commands"
	case strings.HasPrefix(name, ".claude/agents/"):
		return "Claude Code subagents"
	case strings.HasPrefix(name, ".claude/skills/"):
		return "Claude Code skills"
	case strings.HasPrefix(name, ".github/instructions/"):
		return "Copilot instructions"
	case strings.HasPrefix(name, ".windsurf/"):
		return "Windsurf rules"
	case name == ".clinerules" || strings.HasPrefix(name, ".clinerules/"):
		return "Cline rules"
	case strings.HasPrefix(name, ".kiro/"):
		return "Kiro steering"
	}
	// AGENTS.md files in subfolders count as more AGENTS.md.
	switch base := name[strings.LastIndex(name, "/")+1:]; base {
	case "AGENTS.md", "CLAUDE.md", "GEMINI.md":
		return base
	}
	return name
}

var notAlnum = regexp.MustCompile(`[^A-Za-z0-9]`)

// importRe finds @path imports, like "@AGENTS.md" or "See @docs/style.md".
// The @ has to start the line or follow a space, so emails don't count.
var importRe = regexp.MustCompile(`(?:^|\s)@([A-Za-z0-9_./-]+\.md)\b`)

// imports lists the Markdown files a memory file pulls in with @path,
// relative to the repo. Paths outside the repo, home paths and ones inside
// code blocks or code spans are left out.
func imports(name, text string) []string {
	from := ""
	if i := strings.LastIndex(name, "/"); i >= 0 {
		from = name[:i]
	}
	var out []string
	fence := false
	for _, ln := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		t := strings.TrimSpace(ln)
		if strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~") {
			fence = !fence
			continue
		}
		if fence {
			continue
		}
		t = codeSpan.ReplaceAllString(t, "")
		for _, m := range importRe.FindAllStringSubmatch(t, -1) {
			p := m[1]
			if strings.HasPrefix(p, "/") {
				continue
			}
			p = filepath.ToSlash(filepath.Clean(filepath.Join(from, p)))
			if p == ".." || strings.HasPrefix(p, "../") {
				continue
			}
			out = append(out, p)
		}
	}
	return out
}

var codeSpan = regexp.MustCompile("`[^`]*`")

// claudeKey is the folder name Claude Code keeps a project's memory under:
// its path with every character that isn't a letter or digit made a dash.
func claudeKey(dir string) string {
	return notAlnum.ReplaceAllString(filepath.Clean(dir), "-")
}

func mentionsRepo(text string, dirs []string) bool {
	for _, d := range dirs {
		if strings.Contains(text, d) {
			return true
		}
	}
	return false
}

const (
	markStart = "<!-- woolly:start -->"
	markEnd   = "<!-- woolly:end -->"
)

// splitLines keeps the lines worth checking: not blank, not headings, not
// frontmatter, and not Woolly's own section, which is checked as facts.
func splitLines(source, shown, path, text string) []Line {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	all := strings.Split(text, "\n")
	var out []Line
	front, woolly := false, false
	for i, ln := range all {
		t := strings.TrimSpace(ln)
		if i == 0 && t == "---" {
			front = true
			continue
		}
		if front {
			if t == "---" {
				front = false
			}
			continue
		}
		if strings.Contains(t, markStart) {
			woolly = true
		}
		if woolly {
			if strings.Contains(t, markEnd) {
				woolly = false
			}
			continue
		}
		if t == "" || strings.HasPrefix(t, "#") || strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~") || strings.HasPrefix(t, "<!--") {
			continue
		}
		out = append(out, Line{Source: source, File: shown, N: i + 1, Text: t, Path: path})
	}
	return out
}

// claudeMem reads claude-mem's most recent observations for a project with
// the sqlite3 command, when both are there. Its columns have changed
// between versions, so it asks which ones exist first.
func claudeMem(db, project string) []Line {
	if db == "" || project == "" {
		return nil
	}
	if _, err := os.Stat(db); err != nil {
		return nil
	}
	bin, err := exec.LookPath("sqlite3")
	if err != nil {
		return nil
	}
	cols, err := exec.Command(bin, "-readonly", db, "SELECT name FROM pragma_table_info('observations');").Output()
	if err != nil {
		return nil
	}
	have := map[string]bool{}
	for _, c := range strings.Fields(string(cols)) {
		have[c] = true
	}
	if !have["id"] || !have["project"] {
		return nil
	}
	var text []string
	for _, c := range []string{"title", "subtitle", "text", "narrative", "facts"} {
		if have[c] {
			text = append(text, "coalesce("+c+",'')")
		}
	}
	if len(text) == 0 {
		return nil
	}
	q := "SELECT id, replace(replace(" + strings.Join(text, "||' '||") + ", char(10), ' '), char(13), ' ') FROM observations WHERE project = '" +
		strings.ReplaceAll(project, "'", "''") + "' ORDER BY id DESC LIMIT 500;"
	out, err := exec.Command(bin, "-readonly", "-separator", "\t", db, q).Output()
	if err != nil {
		return nil
	}
	var lines []Line
	for _, row := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		f := strings.SplitN(row, "\t", 2)
		if len(f) != 2 {
			continue
		}
		id, err := strconv.Atoi(f[0])
		if err != nil || strings.TrimSpace(f[1]) == "" {
			continue
		}
		lines = append(lines, Line{Source: "claude-mem", File: "claude-mem #" + f[0], N: id, Text: strings.TrimSpace(f[1])})
	}
	return lines
}
