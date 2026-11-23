package memory

import (
	"testing"
	"time"
)

func TestReviewQueue(t *testing.T) {
	s := newStore(t)
	guess, _ := s.Remember(RememberInput{Content: "Uses Postgres", Source: SourceAgent})
	sure, _ := s.Remember(RememberInput{Content: "Uses Redis", Source: SourceAgent, Confidence: 0.9})
	mine, _ := s.Remember(RememberInput{Content: "Lives in Denver"})
	app, _ := s.Remember(RememberInput{Content: "Alice uses MySQL", Source: SourceAgent, User: "alice"})
	later := time.Now().Add(ReviewAge + time.Hour)

	q, err := s.ReviewQueue(later)
	if err != nil {
		t.Fatal(err)
	}
	if len(q) != 1 || q[0].ID != guess.Memory.ID {
		t.Fatalf("queue = %+v (sure %s, mine %s, app %s)", q, sure.Memory.ID, mine.Memory.ID, app.Memory.ID)
	}
	if q, _ := s.ReviewQueue(time.Now()); len(q) != 0 {
		t.Fatalf("nothing is old yet: %+v", q)
	}
	// Using it resets the clock; nothing ever expires on its own.
	s.Touch(guess.Memory.ID)
	if q, _ := s.ReviewQueue(time.Now().Add(ReviewAge - time.Hour)); len(q) != 0 {
		t.Fatalf("used memory still queued: %+v", q)
	}
	if got, _ := s.Get(guess.Memory.ID); got.Status != StatusActive {
		t.Fatalf("status = %s", got.Status)
	}
}
