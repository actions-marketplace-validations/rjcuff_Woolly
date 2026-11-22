package memory

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

// How each memory has been used is kept in usage.json next to the memories,
// not in the Markdown files. Writing it into the files would rewrite them on
// every search, fill woolly.md diffs with noise and wake device sync for
// nothing. It stays on this device and is never synced.

// Usage is how one memory has been used: when it was last recalled or put in
// a session brief, how often, and how often an app said it helped or didn't.
type Usage struct {
	Last time.Time `json:"last"`
	Uses int       `json:"uses,omitempty"`
	Up   int       `json:"up,omitempty"`
	Down int       `json:"down,omitempty"`
}

// Prior is a small ranking nudge from how a memory has been used, from -0.1
// to +0.1. Being used a lot helps a little; feedback helps or hurts more.
// It never hides or removes a memory.
func (u Usage) Prior() float64 {
	p := 0.0
	if u.Uses > 0 {
		p += 0.02 * min(float64(u.Uses), 10) / 10
	}
	if votes := u.Up + u.Down; votes > 0 {
		p += 0.08 * float64(u.Up-u.Down) / float64(votes)
	}
	return max(-0.1, min(0.1, p))
}

func (s *Store) usagePath() string { return filepath.Join(filepath.Dir(s.dir), "usage.json") }

// readUsage reads usage.json. Older versions stored only a time per memory,
// which still loads.
func (s *Store) readUsage() map[string]Usage {
	out := map[string]Usage{}
	b, err := os.ReadFile(s.usagePath())
	if err != nil {
		return out
	}
	if json.Unmarshal(b, &out) == nil {
		return out
	}
	old := map[string]time.Time{}
	if json.Unmarshal(b, &old) == nil {
		for id, t := range old {
			out[id] = Usage{Last: t, Uses: 1}
		}
	}
	return out
}

func (s *Store) writeUsage(u map[string]Usage) {
	// Forget memories that no longer exist, so the file doesn't grow forever.
	if len(u) > 256 {
		s.mu.Lock()
		_ = s.refresh()
		for id := range u {
			if _, ok := s.mems[id]; !ok {
				delete(u, id)
			}
		}
		s.mu.Unlock()
	}
	if b, err := json.Marshal(u); err == nil {
		_ = writeAtomic(s.usagePath(), b)
	}
}

// Touch records that memories were just used.
func (s *Store) Touch(ids ...string) {
	if len(ids) == 0 {
		return
	}
	s.usageMu.Lock()
	defer s.usageMu.Unlock()
	u := s.readUsage()
	now := time.Now().UTC().Truncate(time.Second)
	for _, id := range ids {
		if id == "" {
			continue
		}
		e := u[id]
		e.Last = now
		e.Uses++
		u[id] = e
	}
	s.writeUsage(u)
}

// Feedback records whether a memory helped. It only nudges ranking; nothing
// is ever hidden or deleted because of it.
func (s *Store) Feedback(id string, helpful bool) (Memory, error) {
	m, err := s.Get(id)
	if err != nil {
		return Memory{}, err
	}
	if m.ID == "" {
		return Memory{}, errors.New("memory not found")
	}
	s.usageMu.Lock()
	defer s.usageMu.Unlock()
	u := s.readUsage()
	e := u[m.ID]
	if helpful {
		e.Up++
	} else {
		e.Down++
	}
	u[m.ID] = e
	s.writeUsage(u)
	return m, nil
}

// Usage returns how each memory has been used. Memories never used are
// missing.
func (s *Store) Usage() map[string]Usage {
	s.usageMu.Lock()
	defer s.usageMu.Unlock()
	return s.readUsage()
}

// LastUsed returns when each memory was last used. Memories never used are
// missing.
func (s *Store) LastUsed() map[string]time.Time {
	out := map[string]time.Time{}
	for id, u := range s.Usage() {
		if !u.Last.IsZero() {
			out[id] = u.Last
		}
	}
	return out
}

// writeAtomic writes data through a temp file of its own, so two processes
// writing at once, like the MCP server and woolly serve, never mix their
// writes in one file.
func writeAtomic(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+"-*.tmp")
	if err != nil {
		return err
	}
	_, werr := tmp.Write(data)
	if cerr := tmp.Close(); werr != nil || cerr != nil {
		os.Remove(tmp.Name())
		if werr != nil {
			return werr
		}
		return cerr
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return nil
}
