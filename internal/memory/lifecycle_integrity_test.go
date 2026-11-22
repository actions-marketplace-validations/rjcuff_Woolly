package memory

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSwappedFileWithOldTimeReloads(t *testing.T) {
	s := newStore(t)
	a, _ := s.Remember(RememberInput{Content: "Prefers tea"})
	time.Sleep(20 * time.Millisecond)
	s.Remember(RememberInput{Content: "Lives in Denver"})
	s.List(false) // load and fingerprint
	// A sync tool swaps the older file for another one with an older time:
	// the count and the newest time stay the same.
	oldPath := filepath.Join(s.Dir(), a.Memory.ID+".md")
	data, _ := os.ReadFile(oldPath)
	os.Remove(oldPath)
	id := "20200101-000000-aaaaaaaa"
	text := strings.Replace(strings.Replace(string(data), a.Memory.ID, id, 1), "Prefers tea", "Prefers coffee", 1)
	path := filepath.Join(s.Dir(), id+".md")
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
	past := time.Now().Add(-48 * time.Hour)
	os.Chtimes(path, past, past)
	if _, err := s.Get(id); err != nil {
		t.Fatalf("the swapped-in file wasn't loaded: %v", err)
	}
}

func TestForgetSettlesConflictsWithIt(t *testing.T) {
	s := newStore(t)
	user, _ := s.Remember(RememberInput{Content: "Lives in Austin", Key: "user.location"})
	guess, _ := s.Remember(RememberInput{Content: "Lives in Denver", Key: "user.location", Source: SourceAgent})
	if _, err := s.Forget(user.Memory.ID); err != nil {
		t.Fatal(err)
	}
	got, _ := s.Get(guess.Memory.ID)
	if got.Status != StatusActive || got.ConflictsWith != "" {
		t.Fatalf("after forgetting its partner: %+v", got)
	}
}

func TestRestoreNeverMakesTwoCurrentFacts(t *testing.T) {
	s := newStore(t)
	old, _ := s.Remember(RememberInput{Content: "Lives in Austin", Key: "user.location"})
	if _, err := s.Forget(old.Memory.ID); err != nil {
		t.Fatal(err)
	}
	now, _ := s.Remember(RememberInput{Content: "Lives in Denver", Key: "user.location"})
	back, err := s.Restore(old.Memory.ID)
	if err != nil {
		t.Fatal(err)
	}
	if back.Status != StatusConflict || back.ConflictsWith != now.Memory.ID {
		t.Fatalf("restored = %+v", back)
	}
	active := 0
	mems, _ := s.List(false)
	for _, m := range mems {
		if m.Key == "user.location" {
			active++
		}
	}
	if active != 1 {
		t.Fatalf("%d current facts for the key", active)
	}
}

func TestSupersedeAndConflictGuards(t *testing.T) {
	s := newStore(t)
	a, _ := s.Remember(RememberInput{Content: "A"})
	b, _ := s.Remember(RememberInput{Content: "B"})
	if _, err := s.MarkConflict(a.Memory.ID, a.Memory.ID); err == nil {
		t.Fatal("a memory conflicted with itself")
	}
	if _, err := s.Supersede(a.Memory.ID, b.Memory.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Supersede(b.Memory.ID, a.Memory.ID); err == nil {
		t.Fatal("superseding by a replaced memory should fail")
	}
}

func TestProblemsListsBadFiles(t *testing.T) {
	s := newStore(t)
	res, _ := s.Remember(RememberInput{Content: "Prefers tea"})
	path := filepath.Join(s.Dir(), res.Memory.ID+".md")
	data, _ := os.ReadFile(path)
	os.WriteFile(path, []byte(strings.Replace(string(data), "status: active", "status: actve", 1)), 0o600)
	os.WriteFile(filepath.Join(s.Dir(), "broken.md"), []byte("no header here"), 0o600)
	problems := strings.Join(s.Problems(), "\n")
	if !strings.Contains(problems, "actve") || !strings.Contains(problems, "broken.md") {
		t.Fatalf("problems = %q", problems)
	}
	if got, _ := s.Get(res.Memory.ID); got.Status != StatusActive {
		t.Fatalf("status = %s", got.Status)
	}
}

func TestTouchFromManyGoroutines(t *testing.T) {
	s := newStore(t)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			s.Touch(string(rune('a' + i)))
		}(i)
	}
	wg.Wait()
	if len(s.LastUsed()) == 0 {
		t.Fatal("usage.json is empty or corrupt")
	}
	leftovers, _ := filepath.Glob(filepath.Join(filepath.Dir(s.Dir()), "*.tmp"))
	if len(leftovers) != 0 {
		t.Fatalf("temp files left: %v", leftovers)
	}
}
