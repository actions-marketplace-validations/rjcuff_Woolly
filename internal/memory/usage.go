package memory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// When each memory was last used (recalled or put in a session brief) is
// kept in usage.json next to the memories, not in the Markdown files.
// Writing it into the files would rewrite them on every search, fill
// woolly.md diffs with noise and wake device sync for nothing.

func (s *Store) usagePath() string { return filepath.Join(filepath.Dir(s.dir), "usage.json") }

func (s *Store) readUsage() map[string]time.Time {
	out := map[string]time.Time{}
	if b, err := os.ReadFile(s.usagePath()); err == nil {
		_ = json.Unmarshal(b, &out)
	}
	return out
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
		if id != "" {
			u[id] = now
		}
	}
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
	b, err := json.Marshal(u)
	if err != nil {
		return
	}
	writeAtomic(s.usagePath(), b)
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

// LastUsed returns when each memory was last used. Memories never used are
// missing.
func (s *Store) LastUsed() map[string]time.Time {
	s.usageMu.Lock()
	defer s.usageMu.Unlock()
	return s.readUsage()
}
