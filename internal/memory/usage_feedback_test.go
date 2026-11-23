package memory

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestOldUsageFileStillLoads(t *testing.T) {
	s := newStore(t)
	when := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	old := `{"abc":"` + when.Format(time.RFC3339) + `"}`
	if err := os.WriteFile(filepath.Join(filepath.Dir(s.Dir()), "usage.json"), []byte(old), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := s.LastUsed()["abc"]; !got.Equal(when) {
		t.Fatalf("last used = %v", got)
	}
	if u := s.Usage()["abc"]; u.Uses != 1 {
		t.Fatalf("usage = %+v", u)
	}
}

func TestFeedbackCountsAndPrior(t *testing.T) {
	s := newStore(t)
	res, _ := s.Remember(RememberInput{Content: "Prefers tea"})
	s.Touch(res.Memory.ID)
	if _, err := s.Feedback(res.Memory.ID, true); err != nil {
		t.Fatal(err)
	}
	u := s.Usage()[res.Memory.ID]
	if u.Uses != 1 || u.Up != 1 || u.Prior() <= 0 {
		t.Fatalf("usage = %+v prior %v", u, u.Prior())
	}
	if (Usage{Down: 50}).Prior() < -0.1 || (Usage{Up: 50, Uses: 99}).Prior() > 0.1 {
		t.Fatal("prior must stay within ±0.1")
	}
	if _, err := s.Feedback("nope", true); err == nil {
		t.Fatal("feedback on a missing memory should fail")
	}
}
