package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rjcuff/woolly/internal/gitinfo"
)

// Finding is a memory line that names something the code no longer has.
type Finding struct {
	Line
	What   string `json:"what"`   // the name, path or command that's gone
	Kind   string `json:"kind"`   // "name", "path" or "command"
	Reason string `json:"reason"` // what happened to it
	Commit string `json:"commit,omitempty"`
	Date   string `json:"date,omitempty"`
	// Fix is what What became, when Woolly can tell: a function's new name
	// or a file's new path. check --fix writes it into the line.
	Fix string `json:"fix,omitempty"`
	// State is "new" when the change being checked made the line stale and
	// "before" when it already was on the base commit. The check command
	// fills it in.
	State string `json:"state,omitempty"`
}

// Where is the line's place, like CLAUDE.md:42.
func (f Finding) Where() string {
	if f.Source == "claude-mem" {
		return f.File
	}
	return fmt.Sprintf("%s:%d", f.File, f.N)
}

// Report is what one audit found.
type Report struct {
	Head     string    `json:"head"`
	Files    []File    `json:"files"`
	Lines    int       `json:"lines"`
	Findings []Finding `json:"findings"`
	// Notes are problems with the files themselves, like one too long for
	// an agent to read whole.
	Notes   []string `json:"notes,omitempty"`
	Partial bool     `json:"partial,omitempty"` // it ran out of time
}

// Options tune one audit.
type Options struct {
	Env      Env
	Deadline time.Time // stop looking through history after this; zero means a minute
}

// Run audits the memory the repo's agents read.
func Run(dir string, o Options) Report {
	orig := dir
	if root := gitinfo.Root(dir); root != "" {
		dir = root
	}
	rep := Report{Head: gitinfo.Head(dir), Findings: []Finding{}}
	if len(rep.Head) > 7 {
		rep.Head = rep.Head[:7]
	}
	if rep.Head == "" {
		return rep
	}
	if o.Deadline.IsZero() {
		o.Deadline = time.Now().Add(60 * time.Second)
	}
	lines, files, skip := Read(dir, o.Env, orig)
	rep.Files, rep.Lines = files, len(lines)
	rep.Notes = notes(dir, lines)

	// What each line mentions.
	type mention struct {
		kind, what string
	}
	per := make([][]mention, len(lines))
	var words []string
	seenWord := map[string]bool{}
	for i, l := range lines {
		if negated(l.Text) {
			continue
		}
		for _, n := range names(l.Text) {
			per[i] = append(per[i], mention{"name", n})
			if !seenWord[n] {
				seenWord[n] = true
				words = append(words, n)
			}
		}
		for _, p := range paths(l.Text) {
			per[i] = append(per[i], mention{"path", p})
		}
		for _, c := range commands(l.Text) {
			per[i] = append(per[i], mention{"command", c})
		}
	}

	// Which names are in the code now. Docs and text files don't count: a
	// name only in Markdown isn't code, and neither is one only removed
	// from Markdown.
	files0 := gitinfo.Files(dir)
	spec := append(gitinfo.Exclude(skip), ":(exclude,glob)**/*.md", ":(exclude,glob)**/*.mdx", ":(exclude,glob)**/*.txt")
	present, ok := scanPresent(dir, files0, words, skip, o.Deadline)
	if !ok {
		// The search didn't finish, so any name could still be there.
		rep.Partial = true
		for _, w := range words {
			present[w] = true
		}
	}
	since := sinceOf(dir, files, lines)

	// When each missing name was taken out. A few are looked up one by one,
	// which is quick; many at once are found in one pass through history.
	var missing []string
	for _, w := range words {
		if !present[w] {
			missing = append(missing, w)
		}
	}
	removed := map[string]gitinfo.Change{}
	if len(missing) <= 40 {
		var mu sync.Mutex
		var wg sync.WaitGroup
		jobs := make(chan string)
		for i := 0; i < min(8, runtime.NumCPU()); i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for w := range jobs {
					c, ok := gitinfo.LastChange(dir, wordRe(w), since, spec)
					mu.Lock()
					if ok {
						removed[w] = c
					}
					mu.Unlock()
				}
			}()
		}
		for _, w := range missing {
			if time.Now().After(o.Deadline) {
				rep.Partial = true
				break
			}
			jobs <- w
		}
		close(jobs)
		wg.Wait()
	} else {
		want := map[string]bool{}
		for _, w := range missing {
			want[w] = true
		}
		var complete bool
		removed, complete = gitinfo.Removals(dir, since, want, spec, o.Deadline)
		if !complete {
			rep.Partial = true
		}
	}
	scripts, hasPkg := packageScripts(dir, files0)
	targets, hasMake := makeTargets(dir)

	// Each missing thing is looked up in history once.
	type verdict struct {
		gone bool
		f    Finding
	}
	memo := map[mention]verdict{}
	look := func(m mention) verdict {
		if v, ok := memo[m]; ok {
			return v
		}
		v := verdict{}
		if time.Now().After(o.Deadline) {
			rep.Partial = true
			return v // don't remember it: it's unknown, not fine
		}
		switch m.kind {
		case "name":
			if present[m.what] {
				break
			}
			if c, ok := removed[m.what]; ok {
				v.gone = true
				v.f = Finding{What: m.what, Kind: "name", Commit: c.Commit, Date: c.Date}
				v.f.Reason = m.what + " is no longer in the code"
				if c.Path != "" {
					v.f.Reason = m.what + " was removed from " + c.Path
					if to := renamedTo(dir, c, m.what, present); to != "" {
						v.f.Reason = m.what + " was renamed to " + to + " in " + c.Path
						v.f.Fix = to
					}
				}
			}
		case "path":
			if pathExists(dir, m.what, files0) {
				break
			}
			if c, ok := gitinfo.PathGone(dir, strings.TrimSuffix(m.what, "/"), since); ok {
				v.gone = true
				v.f = Finding{What: m.what, Kind: "path", Commit: c.Commit, Date: c.Date, Reason: m.what + " was deleted"}
				// Follow a file that moved more than once to where it is now.
				to := c.To
				for hop := 0; to != "" && !pathExists(dir, to, nil) && hop < 4; hop++ {
					next, ok := gitinfo.PathGone(dir, strings.TrimSuffix(to, "/"), since)
					if !ok {
						to = ""
						break
					}
					to = next.To
				}
				switch {
				case to != "":
					v.f.Reason = m.what + " was moved to " + to
					v.f.Fix = to
				case c.To != "":
					v.f.Reason = m.what + " was moved to " + c.To + ", which is gone now too"
				}
			}
		case "command":
			tool, name, _ := strings.Cut(m.what, " ")
			if tool == "make" {
				if !hasMake || targets[name] {
					break
				}
				if c, ok := gitinfo.LastChange(dir, "^"+regexp.QuoteMeta(name)+"[[:space:]]*:", since, []string{"Makefile"}); ok {
					v.gone = true
					v.f = Finding{What: m.what, Kind: "command", Commit: c.Commit, Date: c.Date, Reason: "make " + name + " is no longer a target in the Makefile"}
				}
				break
			}
			if !hasPkg || scripts[name] {
				break
			}
			if c, ok := gitinfo.LastChange(dir, `"`+regexp.QuoteMeta(name)+`"[[:space:]]*:`, since, []string{"package.json", ":(glob)**/package.json"}); ok {
				v.gone = true
				v.f = Finding{What: m.what, Kind: "command", Commit: c.Commit, Date: c.Date, Reason: "the " + name + " script is no longer in package.json"}
			}
		}
		memo[m] = v
		return v
	}

	for i, l := range lines {
		for _, m := range per[i] {
			v := look(m)
			if !v.gone {
				continue
			}
			f := v.f
			f.Line = l
			rep.Findings = append(rep.Findings, f)
			break // one reason per line is enough
		}
	}
	return rep
}

var identRe = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*`)

// wordRe matches a name as a whole word, in git's extended regex.
func wordRe(name string) string {
	return "(^|[^A-Za-z0-9_])" + regexp.QuoteMeta(name) + "($|[^A-Za-z0-9_])"
}

// renamedTo guesses what a removed name became: a name the same commit
// added to the same file, that's still in the code and looks like it.
func renamedTo(dir string, c gitinfo.Change, name string, present map[string]bool) string {
	best, score := "", 0
	seen := map[string]bool{}
	for _, l := range gitinfo.Added(dir, c.Commit, c.Path) {
		for _, n := range identRe.FindAllString(l, -1) {
			if n == name || seen[n] || len(n) < 3 {
				continue
			}
			seen[n] = true
			if s := similar(name, n); s > score {
				best, score = n, s
			}
		}
	}
	if best == "" || score*10 < len(name)*6 {
		return ""
	}
	if found, _ := gitinfo.Present(dir, []string{best}, nil); !present[best] && !found[best] {
		return ""
	}
	return best
}

// similar scores two names by their shared start and end.
func similar(a, b string) int {
	la, lb := strings.ToLower(a), strings.ToLower(b)
	p := 0
	for p < len(la) && p < len(lb) && la[p] == lb[p] {
		p++
	}
	s := 0
	for s < len(la)-p && s < len(lb)-p && la[len(la)-1-s] == lb[len(lb)-1-s] {
		s++
	}
	return p + s
}

// negated lines talk about something being gone or forbidden, so naming
// it isn't a mistake.
var negations = []string{"no longer", "removed", "remove ", "deprecated", "renamed", "used to", "instead of", "replaced", "don't", "do not", "never", "avoid", "not use", "legacy", "was dropped", "got rid of", "clear out", "clears out", "delete", "migrat"}

func negated(text string) bool {
	t := strings.ToLower(text)
	for _, n := range negations {
		if strings.Contains(t, n) {
			return true
		}
	}
	return false
}

var (
	backtickRe = regexp.MustCompile("`([^`]+)`")
	// A link to another site, like [docs/drive/](https://github.com/other/repo),
	// names paths in that repo, not this one.
	extLinkRe = regexp.MustCompile(`\[[^\]]*\]\([a-z][a-z0-9+.-]*://[^)]*\)`)
	urlRe     = regexp.MustCompile(`[a-z][a-z0-9+.-]*://\S+`)
	codeName  = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(?:(?:\.|::|#)[A-Za-z_][A-Za-z0-9_]*)*(?:\(\))?$`)
	trimPunct = "\"'.,;:!?()[]{}<>*"
)

// Words that look like code but are names of products and platforms.
var notCode = map[string]bool{"macOS": true, "iOS": true, "iPadOS": true, "watchOS": true, "tvOS": true, "visionOS": true, "iPhone": true, "iPad": true, "iCloud": true, "eBay": true, "npm": true, "pnpm": true}

// names are the code names a line mentions: anything in backticks that
// looks like an identifier, and bare words that can only be code, like
// snake_case, camelCase or names with digits.
func names(text string) []string {
	text = urlRe.ReplaceAllString(extLinkRe.ReplaceAllString(text, " "), " ")
	var out []string
	seen := map[string]bool{}
	add := func(n string) {
		n = strings.TrimSuffix(n, "()")
		if i := strings.LastIndexAny(n, ".:#"); i >= 0 {
			n = n[i+1:]
		}
		if len(n) < 4 || seen[n] || notCode[n] || strings.ToUpper(n) == n && !strings.Contains(n, "_") {
			return
		}
		seen[n] = true
		out = append(out, n)
	}
	for _, m := range backtickRe.FindAllStringSubmatch(text, -1) {
		inner := strings.TrimSpace(m[1])
		if codeName.MatchString(inner) && !isFileName(inner) {
			add(inner)
		}
	}
	bare := backtickRe.ReplaceAllString(text, " ")
	for _, w := range strings.Fields(bare) {
		w = strings.Trim(w, trimPunct)
		if !codeName.MatchString(w) || isFileName(w) {
			continue
		}
		last := w
		if i := strings.LastIndexAny(last, ".:#"); i >= 0 {
			last = last[i+1:]
		}
		if bareCode(strings.TrimSuffix(last, "()")) || strings.HasSuffix(w, "()") {
			add(w)
		}
	}
	return out
}

// bareCode is a word that can't be English: it has an underscore or digit
// inside, or starts lowercase and has a capital later (camelCase).
func bareCode(w string) bool {
	if len(w) < 4 {
		return false
	}
	if strings.Contains(strings.Trim(w, "_"), "_") {
		return true
	}
	if strings.ContainsAny(w[1:], "0123456789") && !strings.ContainsAny(w[:1], "0123456789") {
		return w[0] >= 'a' && w[0] <= 'z' || strings.Contains(w, "_")
	}
	if w[0] >= 'a' && w[0] <= 'z' {
		for _, r := range w[1:] {
			if r >= 'A' && r <= 'Z' {
				return true
			}
		}
	}
	return false
}

var (
	pathRe = regexp.MustCompile(`^[A-Za-z0-9_.\-]+(?:/[A-Za-z0-9_.\-]+)*/?$`)
	extRe  = regexp.MustCompile(`\.[A-Za-z][A-Za-z0-9]{0,7}$`)
)

// paths are the repo paths a line mentions: anything with a slash, and
// file names in backticks.
func paths(text string) []string {
	text = urlRe.ReplaceAllString(extLinkRe.ReplaceAllString(text, " "), " ")
	var out []string
	seen := map[string]bool{}
	add := func(p string, ticked bool) {
		p = strings.TrimPrefix(p, "./")
		if p == "" || seen[p] || strings.Contains(p, "..") || strings.HasPrefix(p, "-") || !pathRe.MatchString(p) {
			return
		}
		slash := strings.Contains(strings.TrimSuffix(p, "/"), "/")
		if !slash && !(ticked && isFileName(p)) {
			return
		}
		if slash && !extRe.MatchString(p) && !strings.HasSuffix(p, "/") && !ticked {
			return // like "and/or", or "TCP/IP"
		}
		seen[p] = true
		out = append(out, p)
	}
	for _, m := range backtickRe.FindAllStringSubmatch(text, -1) {
		for _, w := range strings.Fields(m[1]) {
			add(strings.Trim(w, trimPunct), true)
		}
	}
	for _, w := range strings.Fields(backtickRe.ReplaceAllString(text, " ")) {
		add(strings.Trim(w, trimPunct), false)
	}
	return out
}

// Extensions that make a dotted word a file name, not a method.
var fileExts = map[string]bool{}

func init() {
	for _, e := range strings.Fields("md mdc txt json jsonc yaml yml toml go ts tsx js jsx mjs cjs py rb rs c h cpp hpp cc java kt swift sh bash zsh css scss html sql env lock ini cfg conf xml vue svelte proto graphql gradle mod sum dart ex exs php cs fs zig lua tf hcl plist") {
		fileExts[e] = true
	}
}

func isFileName(w string) bool {
	i := strings.LastIndex(w, ".")
	return i > 0 && fileExts[w[i+1:]]
}

func pathExists(dir, p string, files []string) bool {
	p = strings.TrimSuffix(p, "/")
	if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(p))); err == nil {
		return true
	}
	// People often write paths from a subfolder, like components/Button.tsx.
	for _, f := range files {
		if strings.HasSuffix(f, "/"+p) || strings.Contains(f, "/"+p+"/") || strings.HasPrefix(f, p+"/") {
			return true
		}
	}
	return false
}

var (
	runRe = regexp.MustCompile(`\b(npm|pnpm|yarn|bun) run ([A-Za-z0-9_:.\-]+)`)
	// bun has too many commands of its own, and runs files, so only pnpm and
	// yarn shorthand is read as a script.
	shortRe = regexp.MustCompile(`\b(pnpm|yarn) ([A-Za-z0-9_:.\-]+)(/?)`)
	npmRe   = regexp.MustCompile(`\bnpm (test|start)\b`)
	makeRe  = regexp.MustCompile(`\bmake ([A-Za-z0-9_.\-]+)`)
)

// Commands package managers run themselves, not scripts.
var builtin = map[string]bool{
	"install": true, "i": true, "add": true, "remove": true, "rm": true, "update": true, "up": true, "upgrade": true, "dlx": true, "exec": true, "x": true,
	"create": true, "init": true, "why": true, "list": true, "ls": true, "link": true, "unlink": true, "publish": true, "pack": true, "audit": true,
	"outdated": true, "store": true, "run": true, "config": true, "import": true, "patch": true, "global": true, "info": true, "cache": true, "login": true,
	"workspace": true, "workspaces": true, "version": true, "env": true, "setup": true, "help": true, "-v": true, "--version": true, "install-deps": true,
	"node": true, "dedupe": true, "prune": true, "approve-builds": true, "rebuild": true, "fetch": true, "deploy": true, "licenses": true, "bin": true, "root": true,
	"is": true, "and": true, "or": true, "the": true, "for": true, "to": true, "with": true, "not": true, "build-only": true, "a": true, "an": true,
}

// commands are the package scripts and make targets a line tells you to
// run, written as "npm lint" or "make release".
func commands(text string) []string {
	var out []string
	seen := map[string]bool{}
	add := func(c string) {
		if !seen[c] {
			seen[c] = true
			out = append(out, c)
		}
	}
	for _, m := range runRe.FindAllStringSubmatch(text, -1) {
		add("npm " + m[2])
	}
	for _, m := range npmRe.FindAllStringSubmatch(text, -1) {
		add("npm " + m[1])
	}
	// "pnpm dev" and "make release" read as commands only in code: in
	// backticks, or a line that starts with the command. In prose, "pnpm
	// dependencies" or "make changes" are just words.
	var code []string
	for _, m := range backtickRe.FindAllStringSubmatch(text, -1) {
		code = append(code, m[1])
	}
	line := strings.TrimLeft(strings.TrimSpace(text), "-*+>$ ")
	if strings.HasPrefix(line, "pnpm ") || strings.HasPrefix(line, "yarn ") || strings.HasPrefix(line, "bun ") || strings.HasPrefix(line, "make ") {
		code = append(code, line)
	}
	inCode := strings.Join(code, "\n")
	for _, m := range shortRe.FindAllStringSubmatch(inCode, -1) {
		if m[3] == "" && !builtin[m[2]] && !strings.HasPrefix(m[2], "-") && !strings.Contains(m[2], ".") {
			add("npm " + m[2])
		}
	}
	for _, m := range makeRe.FindAllStringSubmatch(inCode, -1) {
		if !builtin[m[1]] && !strings.HasPrefix(m[1], "-") && m[1] != "sure" && m[1] != "it" && m[1] != "this" && m[1] != "that" {
			add("make " + m[1])
		}
	}
	return out
}

// packageScripts are the scripts in every package.json in the repo, since
// monorepos keep them in each package.
func packageScripts(dir string, files []string) (map[string]bool, bool) {
	out := map[string]bool{}
	found := false
	n := 0
	for _, f := range files {
		if f != "package.json" && !strings.HasSuffix(f, "/package.json") || strings.Contains(f, "node_modules/") {
			continue
		}
		if n++; n > 300 {
			break
		}
		b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(f)))
		if err != nil {
			continue
		}
		// pnpm, yarn and bun also run installed tools, like "pnpm tsx", so
		// dependencies count too.
		var pkg struct {
			Scripts map[string]string `json:"scripts"`
			Deps    map[string]any    `json:"dependencies"`
			Dev     map[string]any    `json:"devDependencies"`
			Peer    map[string]any    `json:"peerDependencies"`
			Opt     map[string]any    `json:"optionalDependencies"`
			Bin     json.RawMessage   `json:"bin"`
		}
		if json.Unmarshal(b, &pkg) != nil {
			continue
		}
		found = true
		for k := range pkg.Scripts {
			out[k] = true
		}
		for _, deps := range []map[string]any{pkg.Deps, pkg.Dev, pkg.Peer, pkg.Opt} {
			for k := range deps {
				out[k] = true
				out[k[strings.LastIndex(k, "/")+1:]] = true
			}
		}
		var bins map[string]string
		if json.Unmarshal(pkg.Bin, &bins) == nil {
			for k := range bins {
				out[k] = true
			}
		}
	}
	return out, found
}

var targetRe = regexp.MustCompile(`(?m)^([A-Za-z0-9_.\-][A-Za-z0-9_.\- ]*?)\s*::?(?:[^=]|$)`)

func makeTargets(dir string) (map[string]bool, bool) {
	b, err := os.ReadFile(filepath.Join(dir, "Makefile"))
	if err != nil {
		return nil, false
	}
	out := map[string]bool{}
	for _, m := range targetRe.FindAllStringSubmatch(string(b), -1) {
		for _, t := range strings.Fields(m[1]) {
			out[t] = true
		}
	}
	// Makefiles that include others or build targets from patterns can
	// have targets this can't see.
	if strings.Contains(string(b), "include ") || strings.Contains(string(b), "%:") || strings.Contains(string(b), "$(eval") {
		return nil, false
	}
	return out, true
}

// Sort orders findings by file, then line.
func Sort(fs []Finding) {
	sort.SliceStable(fs, func(i, j int) bool {
		if fs[i].File != fs[j].File {
			return fs[i].File < fs[j].File
		}
		return fs[i].N < fs[j].N
	})
}

// Fix writes each finding's fix into its line, in files on disk. It returns
// the findings it fixed. Lines that changed since the audit are left alone.
func Fix(findings []Finding) ([]Finding, error) {
	byPath := map[string][]Finding{}
	var order []string
	for _, f := range findings {
		if f.Fix == "" || f.Path == "" {
			continue
		}
		if byPath[f.Path] == nil {
			order = append(order, f.Path)
		}
		byPath[f.Path] = append(byPath[f.Path], f)
	}
	var fixed []Finding
	for _, p := range order {
		b, err := os.ReadFile(p)
		if err != nil {
			return fixed, err
		}
		crlf := strings.Contains(string(b), "\r\n")
		lines := strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n")
		changed := false
		for _, f := range byPath[p] {
			i := f.N - 1
			if i < 0 || i >= len(lines) || strings.TrimSpace(lines[i]) != f.Text {
				continue
			}
			var next string
			if f.Kind == "name" {
				next = replaceWord(lines[i], f.What, f.Fix)
			} else {
				next = strings.ReplaceAll(lines[i], f.What, f.Fix)
			}
			if next != lines[i] {
				lines[i], changed = next, true
				fixed = append(fixed, f)
			}
		}
		if !changed {
			continue
		}
		out := strings.Join(lines, "\n")
		if crlf {
			out = strings.ReplaceAll(out, "\n", "\r\n")
		}
		st, err := os.Stat(p)
		if err != nil {
			return fixed, err
		}
		if err := os.WriteFile(p, []byte(out), st.Mode().Perm()); err != nil {
			return fixed, err
		}
	}
	return fixed, nil
}

func replaceWord(line, old, repl string) string {
	re := regexp.MustCompile(`(^|[^A-Za-z0-9_])` + regexp.QuoteMeta(old) + `($|[^A-Za-z0-9_])`)
	// Run twice, since a match eats the character before the next one.
	for i := 0; i < 2; i++ {
		line = re.ReplaceAllString(line, "${1}"+strings.ReplaceAll(repl, "$", "$$")+"${2}")
	}
	return line
}

// Limits the tools put on what they read. Codex stops reading AGENTS.md at
// 32 KiB, and Claude Code loads the first 200 lines of its MEMORY.md.
const (
	codexLimit  = 32 << 10
	claudeIndex = 200
)

func notes(dir string, lines []Line) []string {
	var out []string
	if b, err := os.ReadFile(filepath.Join(dir, "AGENTS.md")); err == nil {
		if len(b) > codexLimit {
			at := strings.Count(string(b[:codexLimit]), "\n") + 1
			out = append(out, fmt.Sprintf("AGENTS.md is %d KB. Codex stops reading at 32 KB, so it never sees anything after line %d.", len(b)>>10, at))
		}
		claude, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
		if err == nil && !strings.Contains(string(claude), "@AGENTS.md") {
			out = append(out, "Claude Code reads CLAUDE.md, not AGENTS.md. Add a line with @AGENTS.md to CLAUDE.md so it gets the same notes as Codex and Cursor.")
		}
	}
	for _, l := range lines {
		if l.Source == "Claude Code memory" && strings.HasSuffix(l.File, "/MEMORY.md") {
			if b, err := os.ReadFile(l.Path); err == nil {
				if n := strings.Count(string(b), "\n") + 1; n > claudeIndex {
					out = append(out, fmt.Sprintf("Claude Code's MEMORY.md for this repo has %d lines. It loads only the first %d at the start of a session.", n, claudeIndex))
				}
			}
			break
		}
	}
	return out
}

// Text is a plain report, for an agent to read.
func Text(rep Report) string {
	if rep.Head == "" {
		return "This folder isn't a git repo with commits, so there's no code to check memory against."
	}
	if len(rep.Files) == 0 {
		return "No agent memory found for this repo: no CLAUDE.md, AGENTS.md, Cursor rules, Claude Code memory, Codex memories or claude-mem observations."
	}
	var b strings.Builder
	var names []string
	for _, f := range rep.Files {
		names = append(names, f.File)
	}
	fmt.Fprintf(&b, "Checked %d lines in %s against commit %s.\n", rep.Lines, strings.Join(names, ", "), rep.Head)
	if len(rep.Findings) == 0 {
		b.WriteString("Every function, file and command they name is still in the code.\n")
	} else {
		fmt.Fprintf(&b, "%d lines name code that's gone. Don't rely on them; offer to fix them:\n", len(rep.Findings))
		for _, f := range rep.Findings {
			fmt.Fprintf(&b, "- %s: %q. %s", f.Where(), f.Text, f.Reason)
			if f.Commit != "" {
				fmt.Fprintf(&b, " in %s", f.Commit)
			}
			if f.Fix != "" {
				fmt.Fprintf(&b, ". Fix: use %s", f.Fix)
			}
			b.WriteString(".\n")
		}
	}
	for _, n := range rep.Notes {
		b.WriteString("Note: " + n + "\n")
	}
	if rep.Partial {
		b.WriteString("Stopped early to stay fast; some lines weren't checked.\n")
	}
	return b.String()
}

// sinceOf is how far back to look for what a line names: a month before the
// oldest memory file was written. Something removed long before anyone
// wrote the note is history, not a mistake, and skipping it keeps big repos
// fast. Files outside the repo count from a year before they last changed.
func sinceOf(dir string, files []File, lines []Line) string {
	oldest := time.Now()
	paths := map[string]string{}
	for _, l := range lines {
		paths[l.File] = l.Path
	}
	for _, f := range files {
		t := time.Time{}
		p := paths[f.File]
		if rel, err := filepath.Rel(dir, p); err == nil && p != "" && !strings.HasPrefix(rel, "..") {
			if d := gitinfo.FirstAdded(dir, filepath.ToSlash(rel)); d != "" {
				t, _ = time.Parse("2006-01-02", d)
			}
		}
		if t.IsZero() {
			if st, err := os.Stat(p); err == nil {
				t = st.ModTime().AddDate(-1, 0, 0)
			} else {
				t = time.Now().AddDate(-1, 0, 0)
			}
		}
		if t.Before(oldest) {
			oldest = t
		}
	}
	return oldest.AddDate(0, -1, 0).Format("2006-01-02")
}
