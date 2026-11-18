package share

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/rjcuff/woolly/internal/secrets"
)

// The notes people already keep for their agents, read once when a repo
// starts using Woolly so project memory doesn't start empty.

// AgentFiles are the files Woolly reads, relative to the repo. Cursor's
// rules folder is read too.
var AgentFiles = []string{"CLAUDE.md", "AGENTS.md", "GEMINI.md", ".cursorrules", ".windsurfrules", ".github/copilot-instructions.md"}

// Found is one fact read from an agent file.
type Found struct {
	File    string // where it came from, relative to the repo
	Text    string
	Section string // the woolly.md heading it belongs under
}

// maxFound caps how much one first run brings in.
const maxFound = 200

var (
	bulletRe  = regexp.MustCompile(`^\s*(?:[-*+]|\d+[.)])\s+(.+)$`)
	headingRe = regexp.MustCompile(`^\s*#{1,6}\s+(.+)$`)
	boldRe    = regexp.MustCompile(`\*\*([^*]+)\*\*|__([^_]+)__`)
	linkRe    = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
)

// sectionFor picks a woolly.md heading from the heading a line sits under,
// or from the line itself.
func sectionFor(heading, text string) string {
	h := strings.ToLower(heading)
	switch {
	case strings.Contains(h, "command") || strings.Contains(h, "script") || strings.Contains(h, "build") || strings.Contains(h, "test"):
		return "Commands"
	case strings.Contains(h, "gotcha") || strings.Contains(h, "pitfall") || strings.Contains(h, "caveat") || strings.Contains(h, "warning") || strings.Contains(h, "careful"):
		return "Gotchas"
	case strings.Contains(h, "decision"):
		return "Decisions"
	case strings.Contains(h, "stack") || strings.Contains(h, "tech") || strings.Contains(h, "architecture"):
		return "Stack"
	case strings.Contains(h, "convention") || strings.Contains(h, "style") || strings.Contains(h, "rule") || strings.Contains(h, "guideline"):
		return "Conventions"
	case strings.Contains(h, "overview") || strings.Contains(h, "about") || strings.Contains(h, "project"):
		return "Overview"
	}
	return Guess(text)
}

// cleanFact strips markdown emphasis and links so the fact reads as plain text.
func cleanFact(s string) string {
	s = boldRe.ReplaceAllString(s, "$1$2")
	s = linkRe.ReplaceAllString(s, "$1")
	return strings.TrimSpace(strings.TrimRight(strings.TrimSpace(s), ":"))
}

// keep says whether a line is worth bringing in: long enough to say
// something, short enough to be one fact, and nothing personal, secret or
// written to steer an AI.
func keep(s string) bool {
	n := len([]rune(s))
	return n >= 12 && n <= 300 && !LooksPersonal(s) && !secrets.LooksLikeInstructions(s) && len(secrets.Scan(s)) == 0
}

// readAgentFile pulls facts out of one file: list items in Markdown, and
// every plain line in rules files that don't use lists.
func readAgentFile(name, text string) []Found {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	// Woolly's own section is project memory already.
	for {
		i := strings.Index(text, markStart)
		if i < 0 {
			break
		}
		j := strings.Index(text[i:], markEnd)
		if j < 0 {
			text = text[:i]
			break
		}
		text = text[:i] + text[i+j+len(markEnd):]
	}
	// Cursor's .mdc rules start with frontmatter.
	if strings.HasPrefix(text, "---\n") {
		if end := strings.Index(text[4:], "\n---"); end >= 0 {
			text = text[4+end+4:]
		}
	}
	plain := !strings.HasSuffix(name, ".md") && !strings.HasSuffix(name, ".mdc")
	var out []Found
	heading, fenced := "", false
	for _, ln := range strings.Split(text, "\n") {
		t := strings.TrimSpace(ln)
		if strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~") {
			fenced = !fenced
			continue
		}
		if fenced || t == "" {
			continue
		}
		if m := headingRe.FindStringSubmatch(t); m != nil {
			heading = m[1]
			continue
		}
		var fact string
		if m := bulletRe.FindStringSubmatch(ln); m != nil {
			fact = m[1]
		} else if plain && !strings.HasPrefix(t, ">") && !strings.HasPrefix(t, "|") {
			fact = t
		}
		fact = cleanFact(fact)
		if fact == "" || !keep(fact) {
			continue
		}
		out = append(out, Found{File: name, Text: fact, Section: sectionFor(heading, fact)})
	}
	return out
}

// ReadAgentFiles reads the notes a repo already keeps for its agents, in a
// fixed order, without repeats.
func ReadAgentFiles(dir string) []Found {
	names := append([]string{}, AgentFiles...)
	if rules, err := filepath.Glob(filepath.Join(dir, ".cursor", "rules", "*")); err == nil {
		sort.Strings(rules)
		for _, p := range rules {
			if strings.HasSuffix(p, ".md") || strings.HasSuffix(p, ".mdc") {
				if rel, err := filepath.Rel(dir, p); err == nil {
					names = append(names, filepath.ToSlash(rel))
				}
			}
		}
	}
	seen := map[string]bool{}
	var out []Found
	for _, name := range names {
		b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(name)))
		if err != nil || len(b) > 512<<10 {
			continue
		}
		for _, f := range readAgentFile(name, string(b)) {
			key := strings.ToLower(f.Text)
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, f)
			if len(out) == maxFound {
				return out
			}
		}
	}
	return out
}
