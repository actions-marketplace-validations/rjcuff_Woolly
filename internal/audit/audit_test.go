package audit

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func write(t *testing.T, dir, name, text string) {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

// A drone firmware repo whose CLAUDE.md was right when it was written,
// then a refactor renamed a function, dropped a script and a make target,
// and moved a file. Nobody updated CLAUDE.md.
func driftedRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	write(t, dir, "main/imu.c", "int imu_init(void) { return 0; }\nint imu_read_accel(int16_t *out) { return read_raw(out); }\n")
	write(t, dir, "package.json", `{"scripts": {"lint": "eslint .", "dev": "vite"}}`)
	write(t, dir, "Makefile", "build:\n\tgo build\nrelease:\n\t./release.sh\n")
	write(t, dir, "src/flight.ts", "export const legacyMixer = 1;\n")
	write(t, dir, "web/components/Button.tsx", "export function Button() {}\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "first")

	write(t, dir, "main/imu.c", "int imu_init(void) { return 0; }\nint imu_read_accel_g(float *out) { return read_scaled(out); }\n")
	write(t, dir, "package.json", `{"scripts": {"check": "eslint .", "dev": "vite"}}`)
	write(t, dir, "Makefile", "build:\n\tgo build\n")
	if err := os.MkdirAll(filepath.Join(dir, "src", "control"), 0o755); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "mv", "src/flight.ts", "src/control/flight.ts")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "refactor imu and tooling")

	write(t, dir, "CLAUDE.md", strings.Join([]string{
		"# Drone",
		"",
		"## Gotchas",
		"- imu_read_accel returns raw counts; divide by 16384 for g",
		"- Call `imu_init` before anything else",
		"- Run `pnpm lint` before you push",
		"- Cut a release with `make release`",
		"- Flight control lives in src/flight.ts",
		"- Buttons are in components/Button.tsx",
		"- Make sure the battery is unplugged when flashing",
		"- Don't use legacyMixer anymore",
		"- `someHelperWeNeverHad` is a nice idea for later",
		"",
		"<!-- woolly:start -->",
		"- imu_read_accel is woolly's own fact and is checked elsewhere",
		"<!-- woolly:end -->",
	}, "\n"))
	return dir
}

func TestRunFlagsWhatTheCodeNoLongerHas(t *testing.T) {
	dir := driftedRepo(t)
	rep := Run(dir, Options{})
	if rep.Lines == 0 || len(rep.Files) != 1 || rep.Files[0].File != "CLAUDE.md" {
		t.Fatalf("files = %+v, lines = %d", rep.Files, rep.Lines)
	}
	got := map[int]Finding{}
	for _, f := range rep.Findings {
		got[f.N] = f
	}
	want := map[int]string{
		4: "imu_read_accel was renamed to imu_read_accel_g in main/imu.c",
		6: "the lint script is no longer in package.json",
		7: "make release is no longer a target in the Makefile",
		8: "src/flight.ts was moved to src/control/flight.ts",
	}
	for n, reason := range want {
		f, ok := got[n]
		if !ok {
			t.Errorf("line %d wasn't flagged; findings: %+v", n, rep.Findings)
			continue
		}
		if f.Reason != reason {
			t.Errorf("line %d reason = %q, want %q", n, f.Reason, reason)
		}
		if f.Commit == "" || f.Date == "" {
			t.Errorf("line %d has no commit: %+v", n, f)
		}
	}
	if len(rep.Findings) != len(want) {
		t.Errorf("flagged %d lines, want %d: %+v", len(rep.Findings), len(want), rep.Findings)
	}
	if got[4].Where() != "CLAUDE.md:4" {
		t.Errorf("where = %q", got[4].Where())
	}
}

func TestRunReadsClaudeCodeMemory(t *testing.T) {
	dir := driftedRepo(t)
	claude := t.TempDir()
	mem := filepath.Join(claude, "projects", claudeKey(dir), "memory")
	write(t, mem, "imu.md", "---\nname: imu-scale\ndescription: imu_read_accel scaling\n---\n\nimu_read_accel needs dividing by 16384.\n")
	write(t, mem, "fine.md", "imu_init must run first.\n")
	rep := Run(dir, Options{Env: Env{ClaudeDir: claude}})
	var hits []Finding
	for _, f := range rep.Findings {
		if f.Source == "Claude Code memory" {
			hits = append(hits, f)
		}
	}
	if len(hits) != 1 || hits[0].File != "Claude memory/imu.md" || hits[0].N != 6 {
		t.Fatalf("Claude memory findings = %+v", hits)
	}
}

func TestRunReadsCodexMemoryAboutThisRepoOnly(t *testing.T) {
	dir := driftedRepo(t)
	codex := t.TempDir()
	write(t, codex, "memories/this.md", "Working in "+dir+"\n- imu_read_accel gives raw counts\n")
	write(t, codex, "memories/other.md", "Working in /somewhere/else\n- imu_read_accel gives raw counts\n")
	rep := Run(dir, Options{Env: Env{CodexDir: codex}})
	n := 0
	for _, f := range rep.Findings {
		if f.Source == "Codex memory" {
			n++
			if f.File != "Codex memory/this.md" {
				t.Errorf("flagged a memory about another repo: %+v", f)
			}
		}
	}
	if n != 1 {
		t.Fatalf("Codex findings = %d, want 1: %+v", n, rep.Findings)
	}
}

func TestRunReadsClaudeMem(t *testing.T) {
	bin, err := exec.LookPath("sqlite3")
	if err != nil {
		t.Skip("no sqlite3")
	}
	dir := driftedRepo(t)
	db := filepath.Join(t.TempDir(), "claude-mem.db")
	project := filepath.Base(dir)
	sql := "CREATE TABLE observations (id INTEGER PRIMARY KEY, project TEXT NOT NULL, title TEXT, narrative TEXT);" +
		"INSERT INTO observations VALUES (1, '" + project + "', 'IMU scaling', 'imu_read_accel returns raw counts');" +
		"INSERT INTO observations VALUES (2, '" + project + "', 'Init order', 'imu_init runs first');" +
		"INSERT INTO observations VALUES (3, 'other', 'Elsewhere', 'imu_read_accel is fine there');"
	if out, err := exec.Command(bin, db, sql).CombinedOutput(); err != nil {
		t.Fatalf("sqlite3: %v %s", err, out)
	}
	rep := Run(dir, Options{Env: Env{ClaudeMem: db}})
	var hits []Finding
	for _, f := range rep.Findings {
		if f.Source == "claude-mem" {
			hits = append(hits, f)
		}
	}
	if len(hits) != 1 || hits[0].Where() != "claude-mem #1" {
		t.Fatalf("claude-mem findings = %+v", hits)
	}
}

func TestRunOutsideGit(t *testing.T) {
	rep := Run(t.TempDir(), Options{})
	if rep.Head != "" || len(rep.Findings) != 0 {
		t.Fatalf("rep = %+v", rep)
	}
}

func TestNamesPathsAndCommands(t *testing.T) {
	cases := []struct {
		text         string
		names, paths []string
		cmds         []string
	}{
		{"imu_read_accel returns raw counts", []string{"imu_read_accel"}, nil, nil},
		{"Use `Store.SetSources()` to save refs", []string{"SetSources"}, nil, nil},
		{"We ship on macOS and iOS via GitHub", nil, nil, nil},
		{"See https://example.com/a/b.js for HTTP and API docs", nil, nil, nil},
		{"Config is in `config.yaml` and web/app/root.tsx", nil, []string{"config.yaml", "web/app/root.tsx"}, nil},
		{"Use TCP/IP and/or Node.js", nil, nil, nil},
		{"Run `pnpm run test:unit` then `pnpm dev`, then npm test", nil, nil, []string{"npm test:unit", "npm test", "npm dev"}},
		{"Make sure you run `make release` and `pnpm install`", nil, nil, []string{"make release"}},
		{"pnpm dev:web", nil, nil, []string{"npm dev:web"}},
		{"Caches Cargo and pnpm dependencies, then make changes", nil, nil, nil},
		{"Globs sources via `bun scripts/glob.ts` and `pnpm scripts/x.ts`, checks with `bun lint`", nil, []string{"scripts/glob.ts", "scripts/x.ts"}, nil},
		{"Colors: #1a3fd1 on v1.2.3", nil, nil, nil},
		{"Methodology is in [`docs/drive/`](https://github.com/other/repo/tree/main/docs/drive)", nil, nil, nil},
	}
	for _, c := range cases {
		if got := names(c.text); strings.Join(got, ",") != strings.Join(c.names, ",") {
			t.Errorf("names(%q) = %v, want %v", c.text, got, c.names)
		}
		if got := paths(c.text); strings.Join(got, ",") != strings.Join(c.paths, ",") {
			t.Errorf("paths(%q) = %v, want %v", c.text, got, c.paths)
		}
		if got := commands(c.text); strings.Join(got, ",") != strings.Join(c.cmds, ",") {
			t.Errorf("commands(%q) = %v, want %v", c.text, got, c.cmds)
		}
	}
}

func TestClaudeKey(t *testing.T) {
	if got := claudeKey("/Users/ryan/my_app.v2"); got != "-Users-ryan-my-app-v2" {
		t.Fatalf("claudeKey = %q", got)
	}
}

func TestFixRenamesAndMoves(t *testing.T) {
	dir := driftedRepo(t)
	rep := Run(dir, Options{})
	fixed, err := Fix(rep.Findings)
	if err != nil {
		t.Fatal(err)
	}
	if len(fixed) != 2 {
		t.Fatalf("fixed %d lines, want the rename and the move: %+v", len(fixed), fixed)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	text := string(b)
	for _, want := range []string{"- imu_read_accel_g returns raw counts", "- Flight control lives in src/control/flight.ts", "- Run `pnpm lint` before you push", "- imu_read_accel is woolly's own fact"} {
		if !strings.Contains(text, want) {
			t.Errorf("CLAUDE.md lacks %q:\n%s", want, text)
		}
	}
	again := Run(dir, Options{})
	if len(again.Findings) != 2 {
		t.Fatalf("after fixing, the script and make target are still flagged, and only them: %+v", again.Findings)
	}
}

func TestNotes(t *testing.T) {
	dir := driftedRepo(t)
	write(t, dir, "AGENTS.md", strings.Repeat("- keep handlers small and name files after what they export\n", 700))
	claude := t.TempDir()
	write(t, filepath.Join(claude, "projects", claudeKey(dir), "memory"), "MEMORY.md", strings.Repeat("- [a](a.md) a note\n", 250))
	rep := Run(dir, Options{Env: Env{ClaudeDir: claude}})
	all := strings.Join(rep.Notes, "\n")
	for _, want := range []string{"Codex stops reading at 32 KB", "Add a line with @AGENTS.md", "It loads only the first 200"} {
		if !strings.Contains(all, want) {
			t.Errorf("notes lack %q: %v", want, rep.Notes)
		}
	}
	write(t, dir, "CLAUDE.md", "@AGENTS.md\n")
	write(t, dir, "AGENTS.md", "- short\n")
	if rep := Run(dir, Options{}); len(rep.Notes) != 0 {
		t.Fatalf("notes = %v", rep.Notes)
	}
}

// pnpm also runs installed tools, so "pnpm tsx" isn't a missing script.
func TestInstalledToolsAreNotScripts(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q")
	write(t, dir, "package.json", `{"scripts": {"tsx": "tsx"}}`)
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "one")
	write(t, dir, "package.json", `{"scripts": {}, "devDependencies": {"tsx": "^4"}}`)
	write(t, dir, "AGENTS.md", "- Run an example with `pnpm tsx src/basic.ts`\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "two")
	if rep := Run(dir, Options{}); len(rep.Findings) != 0 {
		t.Fatalf("findings = %+v", rep.Findings)
	}
}

// Codex and Claude Code read AGENTS.md and CLAUDE.md in subfolders too.
func TestNestedAgentFiles(t *testing.T) {
	dir := driftedRepo(t)
	write(t, dir, "firmware/AGENTS.md", "- imu_read_accel gives raw counts\n")
	git(t, dir, "add", "firmware/AGENTS.md")
	rep := Run(dir, Options{})
	for _, f := range rep.Findings {
		if f.File == "firmware/AGENTS.md" && f.N == 1 {
			return
		}
	}
	t.Fatalf("nested AGENTS.md wasn't checked: %+v", rep.Files)
}

// Docs aren't code: a name that was only ever in Markdown isn't flagged,
// and a test file's name counts as that name being there.
func TestDocsAndFileNames(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q")
	write(t, dir, "docs/guide.md", "Use `TypeName` links in doc comments.\n")
	write(t, dir, "tests/unit/webcrypto_test.ts", "test('x', () => {})\n")
	write(t, dir, "tests/mod.rs", "mod webcrypto_test;\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "one")
	write(t, dir, "docs/guide.md", "Nothing here now.\n")
	write(t, dir, "tests/mod.rs", "\n")
	write(t, dir, "AGENTS.md", "- PREFER [`TypeName`] references in doc comments\n- Run `cargo test unit::webcrypto_test`\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "two")
	if rep := Run(dir, Options{}); len(rep.Findings) != 0 {
		t.Fatalf("findings = %+v", rep.Findings)
	}
}

func TestMovedFolder(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q")
	write(t, dir, "libs/standard-tests/package.json", `{"name": "standard-tests"}`)
	write(t, dir, "libs/standard-tests/src/index.ts", "export const suite = 1\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "one")
	if err := os.MkdirAll(filepath.Join(dir, "internal"), 0o755); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "mv", "libs/standard-tests", "internal/standard-tests")
	git(t, dir, "commit", "-q", "-m", "move")
	write(t, dir, "AGENTS.md", "| Standard tests | `libs/standard-tests/` |\n")
	rep := Run(dir, Options{})
	if len(rep.Findings) != 1 || rep.Findings[0].Fix != "internal/standard-tests/" {
		t.Fatalf("findings = %+v", rep.Findings)
	}
}

// A file that moved twice is followed to where it is now.
func TestMovedTwice(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q")
	write(t, dir, "src/api.ts", "export const api = 1\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "one")
	os.MkdirAll(filepath.Join(dir, "apps", "web"), 0o755)
	git(t, dir, "mv", "src", "apps/web/src")
	git(t, dir, "commit", "-q", "-m", "two")
	os.MkdirAll(filepath.Join(dir, "packages"), 0o755)
	git(t, dir, "mv", "apps/web/src/api.ts", "packages/api.ts")
	git(t, dir, "commit", "-q", "-m", "three")
	write(t, dir, "AGENTS.md", "- The client lives in src/api.ts\n")
	rep := Run(dir, Options{})
	if len(rep.Findings) != 1 || rep.Findings[0].Fix != "packages/api.ts" {
		t.Fatalf("findings = %+v", rep.Findings)
	}
}

// Claude Code rules, skills, commands and subagents, Copilot's path
// instructions, Windsurf, Cline and Kiro rules are agent files too.
func TestMoreAgentFileFormats(t *testing.T) {
	dir := driftedRepo(t)
	files := map[string]string{
		".claude/rules/firmware.md":              "---\npaths: main/**\n---\n- imu_read_accel gives raw counts\n",
		".claude/skills/flash/SKILL.md":          "---\nname: flash\n---\n- imu_read_accel gives raw counts\n",
		".claude/commands/check.md":              "- imu_read_accel gives raw counts\n",
		".claude/agents/reviewer.md":             "- imu_read_accel gives raw counts\n",
		".github/instructions/c.instructions.md": "- imu_read_accel gives raw counts\n",
		".windsurf/rules/imu.md":                 "- imu_read_accel gives raw counts\n",
		".clinerules/imu.md":                     "- imu_read_accel gives raw counts\n",
		".kiro/steering/imu.md":                  "- imu_read_accel gives raw counts\n",
		"AGENT.md":                               "- imu_read_accel gives raw counts\n",
		"WARP.md":                                "- imu_read_accel gives raw counts\n",
	}
	for name, text := range files {
		write(t, dir, name, text)
	}
	git(t, dir, "add", ".")
	rep := Run(dir, Options{})
	flagged := map[string]string{}
	for _, f := range rep.Findings {
		if f.What == "imu_read_accel" {
			flagged[f.File] = f.Source
		}
	}
	want := map[string]string{
		".claude/rules/firmware.md":              "Claude Code rules",
		".claude/skills/flash/SKILL.md":          "Claude Code skills",
		".claude/commands/check.md":              "Claude Code commands",
		".claude/agents/reviewer.md":             "Claude Code subagents",
		".github/instructions/c.instructions.md": "Copilot instructions",
		".windsurf/rules/imu.md":                 "Windsurf rules",
		".clinerules/imu.md":                     "Cline rules",
		".kiro/steering/imu.md":                  "Kiro steering",
		"AGENT.md":                               "AGENT.md",
		"WARP.md":                                "WARP.md",
	}
	for file, source := range want {
		if flagged[file] != source {
			t.Errorf("%s: flagged as %q, want %q", file, flagged[file], source)
		}
	}
	// The frontmatter line in the rule file isn't a note.
	for _, f := range rep.Findings {
		if f.File == ".claude/rules/firmware.md" && f.N != 4 {
			t.Errorf("rule finding on line %d, want 4", f.N)
		}
	}
}

// Files CLAUDE.md imports with @path are read, and their own imports too,
// but not emails, code spans, home paths or paths outside the repo.
func TestClaudeImports(t *testing.T) {
	dir := driftedRepo(t)
	b, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	write(t, dir, "CLAUDE.md", string(b)+"\nSee @docs/firmware.md and mail ryan@example.md, not `@docs/skip.md`.\n")
	write(t, dir, "docs/firmware.md", "- Use @deeper.md for more\n")
	write(t, dir, "docs/deeper.md", "- imu_read_accel gives raw counts\n")
	write(t, dir, "docs/skip.md", "- imu_read_accel gives raw counts\n")
	git(t, dir, "add", ".")
	rep := Run(dir, Options{})
	got := map[string]bool{}
	for _, f := range rep.Files {
		got[f.File] = true
	}
	if !got["docs/firmware.md"] || !got["docs/deeper.md"] {
		t.Fatalf("imports not read: %+v", rep.Files)
	}
	if got["docs/skip.md"] || got["example.md"] {
		t.Fatalf("read a file that wasn't imported: %+v", rep.Files)
	}
}

func TestImports(t *testing.T) {
	text := "@AGENTS.md\nSee @../../outside.md and @/etc/x.md and @~/x.md\n```\n@in-code.md\n```\n- read `@span.md` or @sub/./style.md\n"
	got := strings.Join(imports("pkg/CLAUDE.md", text), ",")
	if got != "pkg/AGENTS.md,pkg/sub/style.md" {
		t.Fatalf("imports = %q", got)
	}
}
