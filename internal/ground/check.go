package ground

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/rjcuff/woolly/internal/gitinfo"
	"github.com/rjcuff/woolly/internal/memory"
)

// Verdict is what a check found for one fact.
type Verdict string

const (
	OK        Verdict = "ok"        // the code still backs it up
	Changed   Verdict = "changed"   // the code changed, but the fact's values are still there
	Moved     Verdict = "moved"     // the code moved; its sources were updated
	Stale     Verdict = "stale"     // the code it points at is gone or contradicts it
	Unchecked Verdict = "unchecked" // it doesn't point at any code
)

// Result is one fact's check.
type Result struct {
	ID      string       `json:"id"`
	Content string       `json:"content"`
	Verdict Verdict      `json:"verdict"`
	Reason  string       `json:"reason,omitempty"`
	Commit  string       `json:"commit,omitempty"` // the commit that made it stale, when git knows
	Refs    []memory.Ref `json:"sources,omitempty"`
	// Update is set when the fact's sources should be saved again: a
	// source moved, or a fingerprint was missing.
	Update bool `json:"-"`
}

// Repo checks facts against one repo's working tree, caching what it reads.
type Repo struct {
	Dir   string
	files map[string]*file
	old   map[string]*file
	index map[string][]string // definition name to the files that define it
}

type file struct {
	text string
	ok   bool
	defs []Def
	done bool
}

// Open starts checking against the repo at dir.
func Open(dir string) *Repo {
	return &Repo{Dir: dir, files: map[string]*file{}, old: map[string]*file{}}
}

const maxFile = 2 << 20

func (r *Repo) read(p string) *file {
	if f, ok := r.files[p]; ok {
		return f
	}
	f := &file{}
	full := filepath.Join(r.Dir, filepath.FromSlash(p))
	if st, err := os.Stat(full); err == nil && st.Mode().IsRegular() && st.Size() <= maxFile {
		if b, err := os.ReadFile(full); err == nil {
			f.text, f.ok = strings.ReplaceAll(string(b), "\r\n", "\n"), true
		}
	}
	r.files[p] = f
	return f
}

func (r *Repo) atCommit(commit, p string) *file {
	key := commit + ":" + p
	if f, ok := r.old[key]; ok {
		return f
	}
	f := &file{}
	if b, ok := gitinfo.Show(r.Dir, commit, p); ok {
		f.text, f.ok = strings.ReplaceAll(string(b), "\r\n", "\n"), true
	}
	r.old[key] = f
	return f
}

func defsOf(p string, f *file) []Def {
	if !f.done {
		f.defs, f.done = Defs(p, []byte(f.text)), true
	}
	return f.defs
}

// Head is the commit the repo is at, shortened.
func (r *Repo) Head() string {
	h := gitinfo.Head(r.Dir)
	if len(h) > 7 {
		h = h[:7]
	}
	return h
}

// definedIn lists the files that define name, searching the whole repo once.
func (r *Repo) definedIn(name string) []string {
	if r.index == nil {
		r.index = map[string][]string{}
		files := gitinfo.Files(r.Dir)
		if len(files) > 5000 {
			files = files[:5000]
		}
		for _, p := range files {
			if langOf(p) == langOther {
				continue
			}
			f := r.read(p)
			if !f.ok {
				continue
			}
			seen := map[string]bool{}
			for _, d := range defsOf(p, f) {
				for _, n := range []string{d.Name, short(d.Name)} {
					if !seen[n] {
						seen[n] = true
						r.index[n] = append(r.index[n], p)
					}
				}
			}
		}
	}
	return r.index[name]
}

func short(name string) string {
	if i := strings.LastIndexAny(name, ".:"); i >= 0 {
		return name[i+1:]
	}
	return name
}

// Resolve keeps the refs that point at code that exists now, each with a
// fresh fingerprint. Facts saved with sources go through this, so a model
// can't cite a file that isn't there.
func (r *Repo) Resolve(refs []memory.Ref) []memory.Ref {
	var out []memory.Ref
	for _, ref := range refs {
		f := r.read(ref.Path)
		if !f.ok {
			continue
		}
		region, ok := r.region(ref, f)
		if !ok {
			continue
		}
		ref.Hash = Fingerprint(ref.Path, region)
		out = append(out, ref)
	}
	return out
}

// region is the code a ref points at in f.
func (r *Repo) region(ref memory.Ref, f *file) (string, bool) {
	switch {
	case ref.Symbol != "":
		if d, ok := Find(defsOf(ref.Path, f), ref.Symbol); ok {
			return Lines(f.text, d.Start, d.End), true
		}
		if d, ok := mention(f.text, short(ref.Symbol)); ok {
			return Lines(f.text, d.Start, d.End), true
		}
		return "", false
	case ref.Start > 0:
		end := ref.End
		if end == 0 {
			end = ref.Start
		}
		text := Lines(f.text, ref.Start, end)
		return text, text != ""
	}
	return f.text, true
}

// Suggest finds sources for a fact that has none, from the files and code
// names it mentions. It returns at most four.
func (r *Repo) Suggest(content string) []memory.Ref {
	var out []memory.Ref
	names := codeNames(content)
	used := map[string]bool{}
	for _, p := range gitinfo.Paths(r.Dir, content) {
		f := r.read(p)
		if !f.ok {
			continue // a folder, or too big
		}
		found := false
		for _, n := range names {
			if d, ok := Find(defsOf(p, f), n); ok {
				out = append(out, memory.Ref{Path: p, Symbol: d.Name})
				used[n], found = true, true
			}
		}
		if !found {
			out = append(out, memory.Ref{Path: p})
		}
	}
	for _, n := range names {
		if used[n] {
			continue
		}
		if in := r.definedIn(n); len(in) == 1 {
			out = append(out, memory.Ref{Path: in[0], Symbol: n})
		}
	}
	if len(out) > 4 {
		out = out[:4]
	}
	return r.Resolve(out)
}

// Check compares one fact with the code it points at.
func (r *Repo) Check(m memory.Memory) Result {
	res := Result{ID: m.ID, Content: m.Content, Verdict: OK}
	if len(m.Sources) == 0 {
		res.Verdict = Unchecked
		return res
	}
	refs := append([]memory.Ref(nil), m.Sources...)
	var moved, changed []string
	for i, ref := range refs {
		out := r.checkRef(m, ref)
		if out.ref != ref {
			refs[i] = out.ref
			res.Update = true
		}
		switch out.verdict {
		case Stale:
			if res.Verdict != Stale {
				res.Verdict, res.Reason, res.Commit = Stale, out.reason, out.commit
			}
		case Moved:
			moved = append(moved, out.reason)
		case Changed:
			changed = append(changed, out.reason)
		}
	}
	res.Refs = refs
	if res.Verdict == Stale {
		return res
	}
	switch {
	case len(moved) > 0:
		res.Verdict, res.Reason = Moved, strings.Join(moved, "; ")
	case len(changed) > 0:
		res.Verdict, res.Reason = Changed, strings.Join(changed, "; ")
	}
	return res
}

type refOut struct {
	verdict Verdict
	ref     memory.Ref
	reason  string
	commit  string
}

// brokeBy finds the first commit after base where broken says the file at
// path stopped backing up the fact. It returns "" when the change isn't
// committed yet or git can't tell.
func (r *Repo) brokeBy(p, base string, broken func(text string, exists bool) bool) string {
	if gitinfo.Dirty(r.Dir, p) {
		return ""
	}
	for _, c := range gitinfo.Since(r.Dir, base, p) {
		f := r.atCommit(c, p)
		if broken(f.text, f.ok) {
			return c
		}
	}
	return gitinfo.LastTouch(r.Dir, p)
}

func gone(text string, exists bool) bool { return !exists }

func lacksWord(word string) func(string, bool) bool {
	return func(text string, exists bool) bool { return !exists || !hasWord(text, word) }
}

func lacksValue(v string) func(string, bool) bool {
	return func(text string, exists bool) bool { return !exists || !codeTokens(text)[strings.ToLower(v)] }
}

func (r *Repo) checkRef(m memory.Memory, ref memory.Ref) refOut {
	orig := ref
	out := refOut{verdict: OK}
	f := r.read(ref.Path)
	if !f.ok {
		to := gitinfo.RenamedTo(r.Dir, m.Commit, ref.Path)
		if to == "" || !r.read(to).ok {
			return refOut{verdict: Stale, ref: ref, reason: ref.Path + " was deleted", commit: r.brokeBy(ref.Path, m.Commit, gone)}
		}
		ref.Path = to
		f = r.read(to)
		out = refOut{verdict: Moved, reason: orig.Path + " moved to " + to}
	}
	var old *file
	if m.Commit != "" {
		if o := r.atCommit(m.Commit, orig.Path); o.ok {
			old = o
		}
	}

	var region, oldRegion string
	switch {
	case ref.Symbol != "":
		d, found := Find(defsOf(ref.Path, f), ref.Symbol)
		var od Def
		oldFound := false
		if old != nil {
			od, oldFound = Find(defsOf(orig.Path, old), ref.Symbol)
		}
		if !found {
			// Renamed in the same file: a definition with the same body.
			if oldFound {
				was := shape(ref.Path, Lines(old.text, od.Start, od.End), od.Name)
				for _, nd := range defsOf(ref.Path, f) {
					if nd.Name != od.Name && shape(ref.Path, Lines(f.text, nd.Start, nd.End), nd.Name) == was {
						ref.Symbol = nd.Name
						ref.Hash = Fingerprint(ref.Path, Lines(f.text, nd.Start, nd.End))
						what := fmt.Sprintf("%s: %s was renamed to %s", ref.Path, short(orig.Symbol), short(nd.Name))
						if hasWord(m.Content, short(orig.Symbol)) {
							return refOut{verdict: Stale, ref: ref, reason: what, commit: r.brokeBy(ref.Path, m.Commit, lacksWord(short(orig.Symbol)))}
						}
						return refOut{verdict: Moved, ref: ref, reason: what}
					}
				}
			}
			// Moved to another file.
			if in := r.definedIn(ref.Symbol); len(in) == 1 && in[0] != ref.Path {
				to := in[0]
				nf := r.read(to)
				if nd, ok := Find(defsOf(to, nf), ref.Symbol); ok {
					ref.Path = to
					ref.Hash = Fingerprint(to, Lines(nf.text, nd.Start, nd.End))
					return refOut{verdict: Moved, ref: ref, reason: fmt.Sprintf("%s moved from %s to %s", short(ref.Symbol), orig.Path, to)}
				}
			}
			// Still mentioned, just not defined here.
			if md, ok := mention(f.text, short(ref.Symbol)); ok {
				d, found = md, true
			} else {
				return refOut{verdict: Stale, ref: ref, reason: fmt.Sprintf("%s: %s is gone", ref.Path, short(ref.Symbol)),
					commit: r.brokeBy(ref.Path, m.Commit, lacksWord(short(ref.Symbol)))}
			}
		}
		region = Lines(f.text, d.Start, d.End)
		if oldFound {
			oldRegion = Lines(old.text, od.Start, od.End)
		}
	case ref.Start > 0:
		end := ref.End
		if end == 0 {
			end = ref.Start
		}
		region = Lines(f.text, ref.Start, end)
		if old != nil {
			oldRegion = Lines(old.text, ref.Start, end)
			// Lines added above shift the range; find where it went.
			if oldRegion != "" && normalize(region, langOf(ref.Path)) != normalize(oldRegion, langOf(ref.Path)) {
				if at, ok := locate(f.text, oldRegion, ref.Path); ok && at != ref.Start {
					ref.Start = at
					if ref.End > 0 {
						ref.End = at + (end - orig.Start)
					}
					region = oldRegion
					out = refOut{verdict: Moved, reason: fmt.Sprintf("%s lines moved to %d", ref.Path, at)}
				}
			}
		}
		if region == "" {
			return refOut{verdict: Stale, ref: ref, reason: fmt.Sprintf("%s is shorter than line %d now", ref.Path, ref.Start), commit: r.brokeBy(ref.Path, m.Commit, gone)}
		}
	default:
		region = f.text
		if old != nil {
			oldRegion = old.text
		}
	}

	hash := Fingerprint(ref.Path, region)
	if ref.Hash == "" {
		ref.Hash = hash
	} else if ref.Hash != hash && out.verdict == OK {
		out.verdict = Changed
		out.reason = ref.Where() + " changed"
	}
	out.ref = ref

	// The values the fact states, that were in the code when it was saved,
	// must still be there.
	if oldRegion != "" {
		if lost := lostValues(m.Content, oldRegion, region); len(lost) > 0 {
			what := quoteList(lost)
			verb := "is"
			if len(lost) > 1 {
				verb = "are"
			}
			return refOut{verdict: Stale, ref: ref, reason: fmt.Sprintf("%s: %s %s no longer there", ref.Where(), what, verb), commit: r.brokeBy(ref.Path, m.Commit, lacksValue(lost[0]))}
		}
	}
	return out
}

// locate finds where a block of lines sits in text now, ignoring comments
// and spacing, and returns its first line.
func locate(text, block, p string) (int, bool) {
	lg := langOf(p)
	want := strings.Split(block, "\n")
	lines := strings.Split(text, "\n")
	norm := func(s string) string { return normalize(s, lg) }
	first := norm(want[0])
	for i := 0; i+len(want) <= len(lines); i++ {
		if norm(lines[i]) != first {
			continue
		}
		if norm(strings.Join(lines[i:i+len(want)], "\n")) == norm(block) {
			return i + 1, true
		}
	}
	return 0, false
}

var (
	tokenRe   = regexp.MustCompile("0[xX][0-9a-fA-F]+|\\d+(?:\\.\\d+)?|[A-Za-z_][A-Za-z0-9_]*")
	numberRe  = regexp.MustCompile(`^(?:0[xX][0-9a-fA-F]+|\d+(?:\.\d+)?)$`)
	tailNumRe = regexp.MustCompile(`(\d+)$`)
	backtick  = regexp.MustCompile("`([^`]+)`")
)

// codeTokens is every number and name in code. Names also give up their
// parts and trailing numbers, so GPIO_NUM_5 holds 5.
func codeTokens(code string) map[string]bool {
	out := map[string]bool{}
	for _, t := range tokenRe.FindAllString(code, -1) {
		out[strings.ToLower(t)] = true
		if numberRe.MatchString(t) {
			continue
		}
		for _, part := range strings.Split(t, "_") {
			if m := tailNumRe.FindStringSubmatch(part); m != nil {
				out[m[1]] = true
			}
		}
	}
	return out
}

// factValues are the numbers and code names a fact states.
func factValues(fact string) []string {
	var out []string
	seen := map[string]bool{}
	add := func(t string) {
		k := strings.ToLower(t)
		if !seen[k] {
			seen[k] = true
			out = append(out, t)
		}
	}
	for _, t := range tokenRe.FindAllString(fact, -1) {
		if numberRe.MatchString(t) || looksLikeCode(t) {
			add(t)
		}
	}
	for _, m := range backtick.FindAllStringSubmatch(fact, -1) {
		for _, t := range tokenRe.FindAllString(m[1], -1) {
			if len(t) >= 3 || numberRe.MatchString(t) {
				add(t)
			}
		}
	}
	return out
}

// looksLikeCode reports whether a word is a code name rather than English:
// it has an underscore, a digit, or a capital after the first letter.
func looksLikeCode(t string) bool {
	if len(t) < 3 {
		return false
	}
	if strings.ContainsAny(t, "_0123456789") {
		return true
	}
	for _, r := range t[1:] {
		if r >= 'A' && r <= 'Z' {
			// Words in all capitals, like GPIO or SPI, are names too, but
			// they're also how people write acronyms in prose.
			return strings.ToUpper(t) != t
		}
	}
	return false
}

func codeNames(fact string) []string {
	var out []string
	for _, v := range factValues(fact) {
		if !numberRe.MatchString(v) && len(v) >= 3 {
			out = append(out, v)
		}
	}
	return out
}

// lostValues are the fact's values that were in the old code and aren't in
// the new code.
func lostValues(fact, oldCode, newCode string) []string {
	before, after := codeTokens(oldCode), codeTokens(newCode)
	var lost []string
	for _, v := range factValues(fact) {
		k := strings.ToLower(v)
		if before[k] && !after[k] {
			lost = append(lost, v)
		}
	}
	return lost
}

func quoteList(vals []string) string {
	if len(vals) > 3 {
		vals = vals[:3]
	}
	switch len(vals) {
	case 1:
		return vals[0]
	case 2:
		return vals[0] + " and " + vals[1]
	}
	return strings.Join(vals[:len(vals)-1], ", ") + " and " + vals[len(vals)-1]
}

func hasWord(text, word string) bool {
	re, err := regexp.Compile(`(^|[^\w])` + regexp.QuoteMeta(word) + `($|[^\w])`)
	return err == nil && re.MatchString(text)
}

// Summary counts results by verdict, for a one-line report.
func Summary(results []Result) map[Verdict]int {
	out := map[Verdict]int{}
	for _, r := range results {
		out[r.Verdict]++
	}
	return out
}

// Sort puts stale results first, then moved, changed and the rest, keeping
// the original order within each.
func Sort(results []Result) {
	rank := map[Verdict]int{Stale: 0, Moved: 1, Changed: 2, OK: 3, Unchecked: 4}
	sort.SliceStable(results, func(i, j int) bool { return rank[results[i].Verdict] < rank[results[j].Verdict] })
}
