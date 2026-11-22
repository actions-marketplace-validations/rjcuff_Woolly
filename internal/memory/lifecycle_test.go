package memory

import "testing"

// Each case is a sequence of steps on one key, and the status every memory
// should end up in. Steps: user/agent/import remember text; "forget N",
// "restore N" and "stale N" act on the Nth memory saved.
func TestLifecycleTable(t *testing.T) {
	type step struct {
		op   string // user, agent, import, forget, restore, stale
		text string
		n    int
	}
	cases := []struct {
		name  string
		steps []step
		want  []Status
	}{
		{"newer user fact replaces older", []step{{"user", "A", 0}, {"user", "B", 0}}, []Status{StatusSuperseded, StatusActive}},
		{"agent guess conflicts with user", []step{{"user", "A", 0}, {"agent", "B", 0}}, []Status{StatusActive, StatusConflict}},
		{"user replaces agent guess", []step{{"agent", "A", 0}, {"user", "B", 0}}, []Status{StatusSuperseded, StatusActive}},
		{"import beats agent", []step{{"agent", "A", 0}, {"import", "B", 0}}, []Status{StatusSuperseded, StatusActive}},
		{"agent can't replace stale user fact", []step{{"user", "A", 0}, {"stale", "", 0}, {"agent", "B", 0}}, []Status{StatusStale, StatusConflict}},
		{"forgetting the user fact frees the guess", []step{{"user", "A", 0}, {"agent", "B", 0}, {"forget", "", 0}}, []Status{"", StatusActive}},
		{"restore after a newer fact conflicts", []step{{"user", "A", 0}, {"forget", "", 0}, {"user", "B", 0}, {"restore", "", 0}}, []Status{StatusConflict, StatusActive}},
		{"repeat refreshes, no new memory", []step{{"user", "A", 0}, {"user", "A", 0}}, []Status{StatusActive}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := newStore(t)
			var ids []string
			for _, st := range c.steps {
				switch st.op {
				case "user", "agent", "import":
					res, err := s.Remember(RememberInput{Content: st.text, Key: "k", Source: Source(st.op)})
					if err != nil {
						t.Fatal(err)
					}
					if !res.Duplicate {
						ids = append(ids, res.Memory.ID)
					}
				case "forget":
					if _, err := s.Forget(ids[st.n]); err != nil {
						t.Fatal(err)
					}
				case "restore":
					if _, err := s.Restore(ids[st.n]); err != nil {
						t.Fatal(err)
					}
				case "stale":
					if _, err := s.MarkStale(ids[st.n], StaleInput{Reason: "gone"}); err != nil {
						t.Fatal(err)
					}
				}
			}
			if len(ids) != len(c.want) {
				t.Fatalf("%d memories, want %d", len(ids), len(c.want))
			}
			for i, want := range c.want {
				got, err := s.Get(ids[i])
				if want == "" {
					if err == nil {
						t.Fatalf("memory %d should be in the trash, is %s", i, got.Status)
					}
					continue
				}
				if err != nil || got.Status != want {
					t.Fatalf("memory %d: %s (%v), want %s", i, got.Status, err, want)
				}
			}
		})
	}
}
