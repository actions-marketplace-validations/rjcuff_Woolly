// Package why answers "why is this code the way it is?" from project memory:
// the decisions, conventions and gotchas that point at a file, a line or a
// function, and whether the code still backs each one up.
package why

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rjcuff/woolly/internal/decision"
	"github.com/rjcuff/woolly/internal/ground"
	"github.com/rjcuff/woolly/internal/memory"
)

// Target is the code being asked about: a file, a line in it, or a function,
// type or constant in it. Paths are relative to the repo, with forward slashes.
type Target struct {
	Path   string
	Line   int    // 0 for the whole file or a symbol
	Symbol string // "" unless asked by name
}

// ParseTarget reads "path", "path:line" or "path#symbol". A path relative to
// cwd is made relative to the repo at root.
func ParseTarget(s, root, cwd string) (Target, error) {
	s = strings.TrimSpace(s)
	var t Target
	if i := strings.Index(s, "#"); i > 0 {
		t.Symbol, s = s[i+1:], s[:i]
	} else if i := strings.LastIndex(s, ":"); i > 0 {
		if n, err := strconv.Atoi(s[i+1:]); err == nil && n > 0 {
			t.Line, s = n, s[:i]
		}
	}
	if s == "" {
		return t, fmt.Errorf("say which file, like src/auth.ts or src/auth.ts:42")
	}
	p := filepath.FromSlash(s)
	if !filepath.IsAbs(p) {
		p = filepath.Join(cwd, p)
	}
	rel, err := filepath.Rel(root, p)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return t, fmt.Errorf("%s isn't in this repo", s)
	}
	t.Path = filepath.ToSlash(rel)
	return t, nil
}

func (t Target) String() string {
	switch {
	case t.Symbol != "":
		return t.Path + "#" + t.Symbol
	case t.Line > 0:
		return t.Path + ":" + strconv.Itoa(t.Line)
	}
	return t.Path
}

// Match is one memory that explains the target.
type Match struct {
	Memory memory.Memory `json:"memory"`
	// How it's tied to the target: "points at" when one of its sources is the
	// target, "mentions" when only its text names the file or function.
	How string `json:"how"`
	// Ref is the source that matched, if any.
	Ref string `json:"ref,omitempty"`
	// Still is what the code says now: "true", "moved", "changed" or "stale".
	// Empty when the fact doesn't point at code.
	Still  string `json:"still,omitempty"`
	Reason string `json:"reason,omitempty"`
	// Decision parts, when the memory is a decision.
	Choice   string `json:"choice,omitempty"`
	Why      string `json:"why,omitempty"`
	Rejected string `json:"rejected,omitempty"`
}

// Answer is everything Woolly knows about a target.
type Answer struct {
	Target  string  `json:"target"`
	In      string  `json:"in,omitempty"` // the function or type the line is in
	Head    string  `json:"head,omitempty"`
	Matches []Match `json:"matches"`
}

// Find looks through a project's memories for ones about the target, and
// checks each one that points at code against the repo at dir as it is now.
// It reads but never changes memory.
func Find(mems []memory.Memory, project, dir string, t Target) Answer {
	ans := Answer{Target: t.String()}
	repo := ground.Open(dir)
	ans.Head = repo.Head()

	// The lines being asked about, and the definitions they fall in.
	var defs []ground.Def
	if src, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(t.Path))); err == nil {
		defs = ground.Defs(t.Path, src)
	}
	lo, hi := t.Line, t.Line
	var inside []string
	if t.Symbol != "" {
		if d, ok := ground.Find(defs, t.Symbol); ok {
			lo, hi = d.Start, d.End
		}
		inside = append(inside, t.Symbol)
	} else if t.Line > 0 {
		for _, d := range defs {
			if d.Start <= t.Line && t.Line <= d.End {
				inside = append(inside, d.Name)
			}
		}
		if len(inside) > 0 {
			ans.In = inside[len(inside)-1]
		}
	}
	whole := lo == 0

	var out []Match
	for _, m := range mems {
		if !m.In(project) || m.Project == "" || m.Key == "handoff" {
			continue
		}
		if m.Status != memory.StatusActive && m.Status != memory.StatusStale {
			continue
		}
		match, ok := relate(m, t, defs, lo, hi, whole, inside)
		if !ok {
			continue
		}
		if m.Status == memory.StatusStale {
			match.Still = "stale"
			match.Reason = m.StaleReason
			if m.StaleCommit != "" {
				match.Reason += " in " + m.StaleCommit
			}
		} else if len(m.Sources) > 0 {
			res := repo.Check(m)
			match.Still = string(res.Verdict)
			if res.Verdict == ground.OK {
				match.Still = "true"
			}
			match.Reason = res.Reason
		}
		if decision.Is(m) {
			match.Choice, match.Why, match.Rejected = splitDecision(m.Content)
		}
		out = append(out, match)
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if (a.Choice != "") != (b.Choice != "") {
			return a.Choice != "" // decisions first
		}
		if a.How != b.How {
			return a.How == "points at"
		}
		return a.Memory.CreatedAt.After(b.Memory.CreatedAt)
	})
	ans.Matches = out
	return ans
}

// relate says whether a memory is about the target, and how.
func relate(m memory.Memory, t Target, defs []ground.Def, lo, hi int, whole bool, inside []string) (Match, bool) {
	for _, r := range m.Sources {
		if r.Path != t.Path {
			continue
		}
		hit := false
		switch {
		case whole:
			hit = t.Symbol == "" || r.Symbol == t.Symbol || (r.Symbol == "" && r.Start == 0)
		case r.Symbol != "":
			if d, ok := ground.Find(defs, r.Symbol); ok {
				hit = d.Start <= hi && lo <= d.End
			} else {
				hit = contains(inside, r.Symbol)
			}
		case r.Start > 0:
			end := r.End
			if end == 0 {
				end = r.Start
			}
			hit = r.Start <= hi && lo <= end
		default:
			hit = true // the fact is about the whole file
		}
		if hit {
			return Match{Memory: m, How: "points at", Ref: r.Where()}, true
		}
	}
	if len(m.Sources) > 0 {
		return Match{}, false
	}
	// Facts without sources count when they name the file or the function.
	text := m.Content
	base := path.Base(t.Path)
	if strings.Contains(text, t.Path) || (len(base) > 3 && strings.Contains(text, base)) {
		return Match{Memory: m, How: "mentions"}, true
	}
	for _, name := range inside {
		if len(name) > 3 && wordIn(text, name) {
			return Match{Memory: m, How: "mentions"}, true
		}
	}
	return Match{}, false
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// wordIn reports whether name appears in text as a whole word.
func wordIn(text, name string) bool {
	for i := 0; ; {
		j := strings.Index(text[i:], name)
		if j < 0 {
			return false
		}
		j += i
		before := j == 0 || !isWord(text[j-1])
		after := j+len(name) == len(text) || !isWord(text[j+len(name)])
		if before && after {
			return true
		}
		i = j + 1
	}
}

func isWord(c byte) bool {
	return c == '_' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

// splitDecision reads "Decided: X. Why: Y. Rejected: Z." back into its parts.
func splitDecision(s string) (choice, why, rejected string) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "Decided: ")
	if i := strings.Index(s, " Rejected: "); i >= 0 {
		rejected = strings.TrimSuffix(strings.TrimSpace(s[i+len(" Rejected: "):]), ".")
		s = s[:i]
	}
	if i := strings.Index(s, " Why: "); i >= 0 {
		why = strings.TrimSuffix(strings.TrimSpace(s[i+len(" Why: "):]), ".")
		s = s[:i]
	}
	return strings.TrimSuffix(strings.TrimSpace(s), "."), why, rejected
}

// Text is the answer the way the command line and agents read it.
func Text(ans Answer) string {
	var b strings.Builder
	head := ans.Target
	if ans.In != "" {
		head += ", in " + ans.In
	}
	if len(ans.Matches) == 0 {
		fmt.Fprintf(&b, "Woolly doesn't know anything about %s yet.\n", head)
		b.WriteString("Facts and decisions show up here once they point at this code. Agents add sources when they save them, or run: woolly remember --source-ref " + ans.Target + " \"...\"\n")
		return b.String()
	}
	fmt.Fprintf(&b, "%s\n", head)
	for _, m := range ans.Matches {
		b.WriteString("\n")
		when := m.Memory.CreatedAt.Local().Format("Jan 2, 2006")
		if m.Memory.CreatedAt.Year() == time.Now().Year() {
			when = m.Memory.CreatedAt.Local().Format("Jan 2")
		}
		via := ""
		switch m.Memory.Via {
		case "":
		case "command line":
			via = " from the command line"
		default:
			via = " in " + m.Memory.Via
		}
		if m.Choice != "" {
			fmt.Fprintf(&b, "  Decided %s%s: %s.\n", when, via, m.Choice)
			if m.Why != "" {
				fmt.Fprintf(&b, "  Why: %s.\n", m.Why)
			}
			if m.Rejected != "" {
				fmt.Fprintf(&b, "  Rejected: %s.\n", m.Rejected)
			}
		} else {
			kind := memory.TypeOf(m.Memory)
			if kind == "" {
				kind = "note"
			}
			fmt.Fprintf(&b, "  %s, saved %s%s\n", capitalize(kind), when, via)
			fmt.Fprintf(&b, "  %s\n", strings.TrimSpace(m.Memory.Content))
		}
		switch m.Still {
		case "true":
			at := ""
			if ans.Head != "" {
				at = " at " + ans.Head
			}
			fmt.Fprintf(&b, "  Still true%s. Points at %s.\n", at, m.Ref)
		case "moved", "changed":
			fmt.Fprintf(&b, "  Still true, but the code %s: %s.\n", m.Still, m.Reason)
		case "stale":
			fmt.Fprintf(&b, "  Stale: %s. Run woolly review.\n", m.Reason)
		default:
			if m.How == "mentions" {
				b.WriteString("  Mentions this code but doesn't point at it, so Woolly can't check it.\n")
			}
		}
	}
	return b.String()
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
