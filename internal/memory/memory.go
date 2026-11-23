// Package memory stores memories as plain Markdown files with a small header.
package memory

import (
	"crypto/rand"
	"encoding/hex"
	"sort"
	"strings"
	"time"
)

// Source records where a memory came from. It decides which memory wins when
// two memories about the same key disagree.
type Source string

const (
	SourceUser   Source = "user"   // the user stated it directly
	SourceImport Source = "import" // imported from another tool's export
	SourceAgent  Source = "agent"  // an AI inferred it
)

// Rank orders sources by trust: user > import > agent. Unknown sources rank 0.
func (s Source) Rank() int {
	switch s {
	case SourceUser:
		return 3
	case SourceImport:
		return 2
	case SourceAgent:
		return 1
	}
	return 0
}

// ParseSource parses a source name. The empty string is not a valid source.
func ParseSource(s string) (Source, bool) {
	src := Source(strings.ToLower(strings.TrimSpace(s)))
	if src.Rank() == 0 {
		return "", false
	}
	return src, true
}

// Status is the lifecycle state of a memory.
type Status string

const (
	StatusActive     Status = "active"     // current and used for recall
	StatusSuperseded Status = "superseded" // replaced by a newer memory with the same key
	StatusConflict   Status = "conflict"   // disagrees with a more trusted memory; needs review
	StatusExpired    Status = "expired"    // no longer true, with nothing to replace it; kept as history
	StatusStale      Status = "stale"      // the code it points at changed or is gone; needs review
)

// Memory is a single remembered fact.
type Memory struct {
	ID            string    `json:"id"`
	Content       string    `json:"content"`
	Key           string    `json:"key,omitempty"`
	Tags          []string  `json:"tags,omitempty"`
	Source        Source    `json:"source"`
	Status        Status    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	SupersededBy  string    `json:"superseded_by,omitempty"`
	ConflictsWith string    `json:"conflicts_with,omitempty"`
	// Via is the app that saved the memory, like "Cursor" or "command line".
	Via string `json:"via,omitempty"`
	// Project is set for facts that only matter in one project, like a
	// repo's database. Facts about the user have no project.
	Project string `json:"project,omitempty"`
	// Commit is the git commit a project note was saved at, so Woolly can
	// tell when the code it describes has changed since.
	Commit string `json:"commit,omitempty"`
	// User, Agent and Session scope a memory for apps built on Woolly, the
	// way mem0's user_id, agent_id and run_id do. Personal memories leave
	// them empty.
	User    string `json:"user_id,omitempty"`
	Agent   string `json:"agent_id,omitempty"`
	Session string `json:"run_id,omitempty"`
	// InvalidAt is when the memory stopped being true: when a newer fact
	// replaced it or it was marked as no longer true. Nil while current.
	InvalidAt *time.Time `json:"invalid_at,omitempty"`
	// Metadata is free-form data an app attaches to a memory.
	Metadata map[string]any `json:"metadata,omitempty"`
	// Type is what kind of project fact this is: decision, convention,
	// gotcha, bug or hardware. Empty for anything else.
	Type string `json:"type,omitempty"`
	// Sources are the files, symbols or line ranges the fact is about.
	// Woolly checks them against the code to tell when the fact goes stale.
	Sources []Ref `json:"sources,omitempty"`
	// Conf is how sure Woolly is, from 0 to 1. Zero means not set; use
	// Confidence() for the value with defaults.
	Conf float64 `json:"confidence,omitempty"`
	// StaleReason and StaleCommit say why a stale fact went stale and the
	// commit that did it. StaleAt is when it was flagged.
	StaleReason string     `json:"stale_reason,omitempty"`
	StaleCommit string     `json:"stale_commit,omitempty"`
	StaleAt     *time.Time `json:"stale_at,omitempty"`
	// Links are IDs of related memories. Recalling one brings the others
	// along, so an agent gets the whole picture.
	Links []string `json:"links,omitempty"`

	file      string // path on disk; empty until saved
	badStatus string // a status in the file Woolly didn't know, if any
}

func (m *Memory) clone() Memory {
	c := *m
	c.Tags = append([]string(nil), m.Tags...)
	if m.InvalidAt != nil {
		t := *m.InvalidAt
		c.InvalidAt = &t
	}
	c.Sources = append([]Ref(nil), m.Sources...)
	c.Links = append([]string(nil), m.Links...)
	if m.StaleAt != nil {
		t := *m.StaleAt
		c.StaleAt = &t
	}
	if m.Metadata != nil {
		c.Metadata = make(map[string]any, len(m.Metadata))
		for k, v := range m.Metadata {
			c.Metadata[k] = v
		}
	}
	return c
}

// Live reports whether the memory is still current: not replaced by a newer
// one and not marked as no longer true. Stale and conflicting memories are
// live: they wait for review.
func (m *Memory) Live() bool {
	return m.Status != StatusSuperseded && m.Status != StatusExpired
}

// Scope picks out the memories of one user, agent, session or project.
// Empty fields match anything.
type Scope struct {
	User    string
	Agent   string
	Session string
	Project string
	// Personal leaves out memories that apps built on Woolly saved for one
	// of their users, agents or sessions. The user's own AI apps set it, so
	// what an app keeps about "alice" never shows up in their briefs.
	Personal bool
	// Types keeps only memories of these types, like "preference". Empty
	// means any type.
	Types []string
}

// filterTypes normalizes the types a scope asks for. An unknown type is kept
// as given, so it matches nothing instead of turning the filter off.
func filterTypes(ts []string) []string {
	var out []string
	for _, t := range ts {
		if k, ok := NormalizeType(t); ok {
			out = append(out, k)
		} else if t = strings.TrimSpace(t); t != "" {
			out = append(out, "unknown:"+t)
		}
	}
	return out
}

// Unscoped reports whether the memory belongs to the user themselves rather
// than to a user, agent or session of an app built on Woolly.
func (m *Memory) Unscoped() bool { return m.User == "" && m.Agent == "" && m.Session == "" }

// Personal returns the memories that belong to the user themselves.
func Personal(mems []Memory) []Memory {
	var out []Memory
	for i := range mems {
		if mems[i].Unscoped() {
			out = append(out, mems[i])
		}
	}
	return out
}

// CleanScopeID tidies a user, agent or session ID: one line, at most 128 characters.
func CleanScopeID(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > 128 {
		s = string(r[:128])
	}
	return s
}

// Normalized returns the scope with its IDs tidied.
func (sc Scope) Normalized() Scope {
	return Scope{
		User:     CleanScopeID(sc.User),
		Agent:    CleanScopeID(sc.Agent),
		Session:  CleanScopeID(sc.Session),
		Project:  NormalizeProject(sc.Project),
		Personal: sc.Personal,
		Types:    filterTypes(sc.Types),
	}
}

// Empty reports whether the scope has no user, agent or session. It ignores
// Personal: a personal scope still filters.
func (sc Scope) Empty() bool { return sc.User == "" && sc.Agent == "" && sc.Session == "" }

// Matches reports whether m is in the scope. A user, agent or session must
// match exactly when set; the project follows In.
func (sc Scope) Matches(m *Memory) bool {
	if sc.Personal && !m.Unscoped() {
		return false
	}
	if len(sc.Types) > 0 {
		t, ok := TypeOf(*m), false
		for _, want := range sc.Types {
			ok = ok || t == want
		}
		if !ok {
			return false
		}
	}
	return (sc.User == "" || m.User == sc.User) &&
		(sc.Agent == "" || m.Agent == sc.Agent) &&
		(sc.Session == "" || m.Session == sc.Session) &&
		m.In(sc.Project)
}

// sameSlot reports whether two memories are in exactly the same scope, so
// one can repeat or replace the other.
func sameSlot(a, b *Memory) bool {
	return a.Project == b.Project && a.User == b.User && a.Agent == b.Agent && a.Session == b.Session
}

// NormalizeKey lowercases a key and joins whitespace with underscores, so
// "User Location" and "user_location" are the same key.
func NormalizeKey(k string) string {
	return strings.Join(strings.Fields(strings.ToLower(k)), "_")
}

// NormalizeTags trims, lowercases, de-duplicates and sorts tags.
func NormalizeTags(tags []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, t := range tags {
		for _, part := range strings.Split(t, ",") {
			p := strings.Join(strings.Fields(strings.ToLower(part)), "-")
			if p == "" || seen[p] {
				continue
			}
			seen[p] = true
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out
}

func mergeTags(a, b []string) []string {
	return NormalizeTags(append(append([]string(nil), a...), b...))
}

func sameText(a, b string) bool {
	return strings.EqualFold(strings.Join(strings.Fields(a), " "), strings.Join(strings.Fields(b), " "))
}

// newID returns a time-sortable, filename-safe ID such as 20261110-153000-1a2b3c4d.
func newID(t time.Time) string {
	var b [4]byte
	_, _ = rand.Read(b[:])
	return t.UTC().Format("20060102-150405") + "-" + hex.EncodeToString(b[:])
}

// NormalizeProject lowercases a project name and keeps it filename and
// URL safe: "My App" becomes "my-app".
func NormalizeProject(p string) string {
	p = strings.Join(strings.Fields(strings.ToLower(p)), "-")
	var b strings.Builder
	for _, r := range p {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
		}
	}
	out := strings.Trim(b.String(), "-.")
	if len(out) > 60 {
		out = out[:60]
	}
	return out
}

// In reports whether the memory belongs in project: global memories are in
// every project, and project memories only in their own. An empty project
// means "everywhere", which includes every memory.
func (m *Memory) In(project string) bool {
	return project == "" || m.Project == "" || m.Project == project
}

// CleanApp tidies an app name from a client: one line, at most 40 characters.
func CleanApp(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > 40 {
		s = string(r[:40])
	}
	return s
}
