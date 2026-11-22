package memory

import (
	"errors"
	"path"
	"strconv"
	"strings"
	"time"
)

// A grounded fact points at where it came from in the code, so Woolly can
// tell when the code changes underneath it.

// Memory types. A memory has at most one. The first five are kinds of
// project fact; the last three are kinds of memory agents keep about a user
// or a task.
const (
	TypeDecision   = "decision"
	TypeConvention = "convention"
	TypeGotcha     = "gotcha"
	TypeBug        = "bug"
	TypeHardware   = "hardware"
	// TypePreference is how someone likes things done.
	TypePreference = "preference"
	// TypeProcedure is how to do something, step by step.
	TypeProcedure = "procedure"
	// TypeEpisode is something that happened, like a session or an incident.
	TypeEpisode = "episode"
)

// Types lists every memory type.
var Types = []string{TypeDecision, TypeConvention, TypeGotcha, TypeBug, TypeHardware, TypePreference, TypeProcedure, TypeEpisode}

// ProjectType reports whether t is a kind of project fact, which belongs
// under a heading in woolly.md.
func ProjectType(t string) bool {
	k, _ := NormalizeType(t)
	switch k {
	case TypeDecision, TypeConvention, TypeGotcha, TypeBug, TypeHardware:
		return true
	}
	return false
}

// NormalizeTypes keeps the known types in a list, normalized.
func NormalizeTypes(ts []string) []string {
	var out []string
	for _, t := range ts {
		if k, ok := NormalizeType(t); ok {
			out = append(out, k)
		}
	}
	return out
}

// NormalizeType returns a known fact type, accepting plurals and any case.
// It returns "" and false for anything else.
func NormalizeType(t string) (string, bool) {
	t = strings.ToLower(strings.TrimSpace(t))
	if t == "" {
		return "", false
	}
	for _, k := range Types {
		if t == k || t == k+"s" {
			return k, true
		}
	}
	return "", false
}

// TypeOf is a memory's type, or the type its tags imply.
func TypeOf(m Memory) string {
	if m.Type != "" {
		return m.Type
	}
	for _, t := range m.Tags {
		if k, ok := NormalizeType(t); ok {
			return k
		}
	}
	return ""
}

// Ref points at code a fact is about: a file, a symbol in a file, or a range
// of lines. Hash is a short fingerprint of that code when the fact was
// saved, so a later check can tell whether it changed.
//
//	main/imu.c                   the file
//	main/imu.c#imu_read_accel    a function, type, macro or constant in it
//	main/pins.h:12-18            lines 12 to 18
//	main/pins.h:12-18@77aa01b3   with the fingerprint
type Ref struct {
	Path   string
	Symbol string
	Start  int
	End    int
	Hash   string
}

// String renders a ref as it's written in a memory file.
func (r Ref) String() string {
	s := r.Where()
	if r.Hash != "" {
		s += "@" + r.Hash
	}
	return s
}

// Where is the ref without its fingerprint.
func (r Ref) Where() string {
	s := r.Path
	switch {
	case r.Symbol != "":
		s += "#" + r.Symbol
	case r.Start > 0 && r.End > r.Start:
		s += ":" + strconv.Itoa(r.Start) + "-" + strconv.Itoa(r.End)
	case r.Start > 0:
		s += ":" + strconv.Itoa(r.Start)
	}
	return s
}

// MarshalText writes a ref as its string, so it's a plain string in JSON.
func (r Ref) MarshalText() ([]byte, error) { return []byte(r.String()), nil }

// UnmarshalText parses a ref from JSON.
func (r *Ref) UnmarshalText(b []byte) error {
	ref, ok := ParseRef(string(b))
	if !ok {
		return errors.New("bad source: want path, path#symbol or path:start-end")
	}
	*r = ref
	return nil
}

// ParseRef reads a ref. The path must be relative to the repo and stay
// inside it.
func ParseRef(s string) (Ref, bool) {
	s = strings.TrimSpace(s)
	var r Ref
	if i := strings.LastIndex(s, "@"); i > 0 && isHex(s[i+1:]) {
		r.Hash, s = s[i+1:], s[:i]
	}
	if i := strings.Index(s, "#"); i >= 0 {
		r.Symbol, s = s[i+1:], s[:i]
		if !validSymbol(r.Symbol) {
			return Ref{}, false
		}
	} else if i := strings.LastIndex(s, ":"); i > 0 {
		lo, hi, ok := strings.Cut(s[i+1:], "-")
		a, err := strconv.Atoi(lo)
		if err == nil && a > 0 {
			r.Start = a
			if ok {
				b, err := strconv.Atoi(hi)
				if err != nil || b < a {
					return Ref{}, false
				}
				r.End = b
			}
			s = s[:i]
		}
	}
	p, ok := CleanPath(s)
	if !ok {
		return Ref{}, false
	}
	r.Path = p
	if r.End == r.Start {
		r.End = 0
	}
	return r, true
}

// ParseRefs reads a comma-separated list of refs, skipping bad ones.
func ParseRefs(s string) []Ref {
	var out []Ref
	for _, part := range strings.Split(s, ",") {
		if r, ok := ParseRef(part); ok {
			out = append(out, r)
		}
	}
	return out
}

// CleanPath tidies a repo-relative path: forward slashes, no "./", and
// never absolute or outside the repo.
func CleanPath(p string) (string, bool) {
	p = strings.TrimSpace(strings.ReplaceAll(p, `\`, "/"))
	if p == "" || strings.ContainsAny(p, ", \t\n<>\"") || strings.HasPrefix(p, "/") || (len(p) > 1 && p[1] == ':') {
		return "", false
	}
	p = path.Clean(p)
	if p == "." || p == ".." || strings.HasPrefix(p, "../") || len(p) > 300 {
		return "", false
	}
	return p, true
}

func validSymbol(s string) bool {
	if s == "" || len(s) > 120 {
		return false
	}
	for _, r := range s {
		if !(r == '_' || r == '.' || r == ':' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}

func isHex(s string) bool {
	if len(s) < 4 || len(s) > 64 {
		return false
	}
	for _, r := range s {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') {
			return false
		}
	}
	return true
}

func formatRefs(refs []Ref) string {
	var parts []string
	for _, r := range refs {
		parts = append(parts, r.String())
	}
	return strings.Join(parts, ", ")
}

// cleanRefs drops bad refs and repeats, keeping the order.
func cleanRefs(refs []Ref) []Ref {
	var out []Ref
	seen := map[string]bool{}
	for _, r := range refs {
		p, ok := CleanPath(r.Path)
		if !ok || (r.Symbol != "" && !validSymbol(r.Symbol)) {
			continue
		}
		r.Path = p
		if seen[r.Where()] {
			continue
		}
		seen[r.Where()] = true
		out = append(out, r)
	}
	return out
}

// ClampConfidence keeps a confidence between 0 and 1, rounded to hundredths.
func ClampConfidence(c float64) float64 {
	if c < 0 || c != c {
		return 0
	}
	if c > 1 {
		return 1
	}
	return float64(int(c*100+0.5)) / 100
}

// Confidence is how sure Woolly is that a fact is right: the saved value,
// or a default by source when none was saved.
func (m *Memory) Confidence() float64 {
	if m.Conf > 0 {
		return m.Conf
	}
	switch m.Source {
	case SourceUser:
		return 1
	case SourceImport:
		return 0.8
	}
	return 0.6
}

// Grounded reports whether the fact points at code.
func (m *Memory) Grounded() bool { return len(m.Sources) > 0 }

// StaleInput says why a fact went stale.
type StaleInput struct {
	Reason string // like "imu_read_accel was removed"
	Commit string // the commit that changed it, if known
}

// MarkStale flags a current fact as stale: the code it points at changed or
// is gone. Nothing is deleted; Confirm or Expire settles it.
func (s *Store) MarkStale(id string, in StaleInput) (Memory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(); err != nil {
		return Memory{}, err
	}
	m, err := s.find(id)
	if err != nil {
		return Memory{}, err
	}
	if m.Status != StatusActive && m.Status != StatusStale {
		return Memory{}, errors.New("only a current fact can go stale")
	}
	reason := oneLineClip(in.Reason, 200)
	if m.Status == StatusStale && m.StaleReason == reason && m.StaleCommit == in.Commit {
		return m.clone(), nil
	}
	now := time.Now().UTC()
	if m.Status != StatusStale {
		m.StaleAt = &now
	}
	m.Status = StatusStale
	m.StaleReason = reason
	m.StaleCommit = strings.TrimSpace(in.Commit)
	m.UpdatedAt = now
	if err := s.write(m); err != nil {
		return Memory{}, err
	}
	return m.clone(), nil
}

// Confirm says a stale fact is still true. It becomes current again, and
// its sources are replaced by refs when refs isn't nil (with fresh
// fingerprints, from the caller). commit, when set, is the commit the code
// was checked at, so later checks compare against it.
func (s *Store) Confirm(id string, refs []Ref, commit string) (Memory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(); err != nil {
		return Memory{}, err
	}
	m, err := s.find(id)
	if err != nil {
		return Memory{}, err
	}
	if !m.Live() {
		return Memory{}, errors.New("this fact is no longer current")
	}
	if refs != nil {
		m.Sources = cleanRefs(refs)
	}
	if commit != "" {
		m.Commit = commit
	}
	m.Status = StatusActive
	m.StaleReason, m.StaleCommit, m.StaleAt = "", "", nil
	m.UpdatedAt = time.Now().UTC()
	if m.Key != "" {
		if _, err := s.resolveLocked(m, m.UpdatedAt); err != nil {
			return Memory{}, err
		}
	} else if err := s.write(m); err != nil {
		return Memory{}, err
	}
	return m.clone(), nil
}

// SetSources replaces a fact's sources without changing its status, as when
// a check finds a symbol that moved to another file or records a missing
// fingerprint. commit is saved too when the fact has none, so later checks
// have code to compare against.
func (s *Store) SetSources(id string, refs []Ref, commit string) (Memory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(); err != nil {
		return Memory{}, err
	}
	m, err := s.find(id)
	if err != nil {
		return Memory{}, err
	}
	m.Sources = cleanRefs(refs)
	if m.Commit == "" {
		m.Commit = commit
	}
	// A newer time lets sync carry the new sources to other devices.
	m.UpdatedAt = time.Now().UTC()
	if err := s.write(m); err != nil {
		return Memory{}, err
	}
	return m.clone(), nil
}

// Stale lists stale facts, newest first. An empty project means every one.
func (s *Store) Stale(project string) ([]Memory, error) {
	project = NormalizeProject(project)
	return s.filter(func(m *Memory) bool {
		return m.Status == StatusStale && (project == "" || m.Project == project)
	})
}

func oneLineClip(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > n {
		s = string(r[:n])
	}
	return s
}
