package memory

import (
	"sort"
	"time"
)

// ReviewAge is how long an AI's guess can go unused before Woolly asks
// whether it's still true.
const ReviewAge = 90 * 24 * time.Hour

// ReviewQueue lists current memories worth a "still true?" check: ones an AI
// guessed, that Woolly isn't sure of, and that nothing has used or changed
// for ReviewAge. Oldest first. Nothing in it expires on its own; the user
// keeps, edits or forgets each one.
func (s *Store) ReviewQueue(now time.Time) ([]Memory, error) {
	usage := s.Usage()
	mems, err := s.filter(func(m *Memory) bool {
		if m.Status != StatusActive || m.Source != SourceAgent || m.Confidence() >= 0.7 || !m.Unscoped() {
			return false
		}
		last := m.UpdatedAt
		if u := usage[m.ID].Last; u.After(last) {
			last = u
		}
		return now.Sub(last) >= ReviewAge
	})
	if err != nil {
		return nil, err
	}
	sort.SliceStable(mems, func(i, j int) bool { return mems[i].UpdatedAt.Before(mems[j].UpdatedAt) })
	return mems, nil
}
