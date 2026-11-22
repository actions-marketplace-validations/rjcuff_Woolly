package memory

// QA tests for F4 (lifecycle integrity), G1 (types), G2 (usage and feedback),
// S1 (links) and S2 (review queue), written against
// plans/memory-growth-and-fix.md.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func usageFile(s *Store) string { return filepath.Join(filepath.Dir(s.Dir()), "usage.json") }

// rewriteField changes one front-matter field of a memory file by hand, the
// way a user or a sync tool might.
func rewriteField(t *testing.T, s *Store, id, field, value string) {
	t.Helper()
	path := filepath.Join(s.Dir(), id+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	found := false
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, field+":") {
			line = field + ": " + value
			found = true
		}
		out = append(out, line)
	}
	text := strings.Join(out, "\n")
	if !found {
		text = strings.Replace(text, "---\n", "---\n"+field+": "+value+"\n", 1)
	}
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
}

func noPanic(t *testing.T, what string, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("%s panicked: %v", what, r)
		}
	}()
	f()
}

// --- G2: usage.json ---

func TestQAUsageFileCorruptOrOdd(t *testing.T) {
	for name, content := range map[string]string{
		"empty":   "",
		"garbage": "{not json",
		"array":   "[1,2,3]",
		"null":    "null",
		"number":  "42",
		"wrong":   `{"x": 5}`,
	} {
		t.Run(name, func(t *testing.T) {
			s := newStore(t)
			res, _ := s.Remember(RememberInput{Content: "Prefers tea"})
			if err := os.WriteFile(usageFile(s), []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
			noPanic(t, "Usage", func() { s.Usage() })
			noPanic(t, "Touch", func() { s.Touch(res.Memory.ID) })
			noPanic(t, "Feedback", func() {
				if _, err := s.Feedback(res.Memory.ID, true); err != nil {
					t.Fatal(err)
				}
			})
			u := s.Usage()[res.Memory.ID]
			if u.Uses != 1 || u.Up != 1 {
				t.Fatalf("after a corrupt %s usage.json: %+v", name, u)
			}
		})
	}
}

func TestQAPriorBounds(t *testing.T) {
	cases := []struct {
		u        Usage
		min, max float64
	}{
		{Usage{}, 0, 0},
		{Usage{Uses: 1}, 0.001, 0.003},
		{Usage{Uses: 1000}, 0.02, 0.02},
		{Usage{Up: 1}, 0.08, 0.08},
		{Usage{Down: 1}, -0.08, -0.08},
		{Usage{Up: 1000, Uses: 1000}, 0.1, 0.1},
		{Usage{Down: 1000, Uses: 1000}, -0.06, -0.06},
		{Usage{Up: 5, Down: 5}, 0, 0},
		{Usage{Uses: -5, Up: -3, Down: 0}, -0.1, 0.1}, // hand-edited nonsense stays in bounds
	}
	for _, c := range cases {
		p := c.u.Prior()
		if p < c.min-1e-9 || p > c.max+1e-9 || p < -0.1 || p > 0.1 {
			t.Errorf("Prior(%+v) = %v, want in [%v, %v]", c.u, p, c.min, c.max)
		}
	}
}

func TestQAFeedbackByPrefixAndOnHistory(t *testing.T) {
	s := newStore(t)
	old, _ := s.Remember(RememberInput{Content: "Lives in Austin", Key: "loc"})
	s.Remember(RememberInput{Content: "Lives in Denver", Key: "loc"})
	// Rating a replaced version works (it's still a memory), and never
	// changes its status.
	m, err := s.Feedback(old.Memory.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if m.Status != StatusSuperseded {
		t.Fatalf("status = %s", m.Status)
	}
	got, _ := s.Get(old.Memory.ID)
	if got.Status != StatusSuperseded {
		t.Fatalf("feedback changed status to %s", got.Status)
	}
	// A prefix is recorded under the full ID.
	prefix := old.Memory.ID[:len(old.Memory.ID)-2]
	if _, err := s.Feedback(prefix, true); err != nil && !strings.Contains(err.Error(), "more than one") {
		t.Fatal(err)
	} else if err == nil {
		if _, ok := s.Usage()[prefix]; ok {
			t.Fatal("feedback stored under the prefix, not the ID")
		}
	}
	if _, err := s.Feedback("", true); err == nil {
		t.Fatal("empty ID accepted")
	}
}

func TestQATouchIgnoresEmptyAndCountsRepeats(t *testing.T) {
	s := newStore(t)
	s.Touch()
	s.Touch("")
	if _, err := os.Stat(usageFile(s)); err == nil {
		if len(s.Usage()) != 0 {
			t.Fatalf("empty touches recorded: %v", s.Usage())
		}
	}
	s.Touch("a", "a")
	if u := s.Usage()["a"]; u.Uses != 2 {
		t.Fatalf("uses = %d", u.Uses)
	}
}

// --- F4: restore and forget ---

func TestQARestoreWhenFileAlreadyExists(t *testing.T) {
	s := newStore(t)
	res, _ := s.Remember(RememberInput{Content: "Prefers tea"})
	path := filepath.Join(s.Dir(), res.Memory.ID+".md")
	data, _ := os.ReadFile(path)
	// A sync tool brought the file back while a copy is in the trash.
	os.MkdirAll(s.trashDir(), 0o700)
	if err := os.WriteFile(filepath.Join(s.trashDir(), res.Memory.ID+".md"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Restore(res.Memory.ID); err == nil {
		t.Fatal("restore overwrote a live memory")
	}
	if _, err := os.Stat(filepath.Join(s.trashDir(), res.Memory.ID+".md")); err != nil {
		t.Fatalf("a failed restore lost the trash copy: %v", err)
	}
	mems, _ := s.List(true)
	if len(mems) != 1 {
		t.Fatalf("%d memories", len(mems))
	}
}

func TestQARestorePrefixAmbiguity(t *testing.T) {
	s := newStore(t)
	a, _ := s.Remember(RememberInput{Content: "A fact"})
	b, _ := s.Remember(RememberInput{Content: "B fact"})
	s.Forget(a.Memory.ID)
	s.Forget(b.Memory.ID)
	// IDs made in the same second share "YYYYMMDD-HHMMSS-".
	common := 0
	for common < len(a.Memory.ID) && a.Memory.ID[common] == b.Memory.ID[common] {
		common++
	}
	if _, err := s.Restore(a.Memory.ID[:common]); err == nil {
		t.Fatal("an ambiguous prefix restored something")
	}
	if _, err := s.Restore(""); err == nil {
		t.Fatal("an empty ID restored something")
	}
	if _, err := s.Restore(a.Memory.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Restore(b.Memory.ID[:common+3]); err != nil {
		t.Fatalf("a now-unique prefix: %v", err)
	}
}

// An exact ID must win over a longer name that starts with it, the way
// find() treats live memories.
func TestQARestoreExactIDBeatsLongerName(t *testing.T) {
	s := newStore(t)
	a, _ := s.Remember(RememberInput{Content: "A fact"})
	s.Forget(a.Memory.ID)
	data, _ := os.ReadFile(filepath.Join(s.trashDir(), a.Memory.ID+".md"))
	other := a.Memory.ID + "b"
	os.WriteFile(filepath.Join(s.trashDir(), other+".md"), []byte(strings.Replace(string(data), a.Memory.ID, other, 1)), 0o600)
	if _, err := s.Restore(a.Memory.ID); err != nil {
		t.Fatalf("restore by exact ID: %v", err)
	}
}

func TestQAForgetLinkedAndConflicted(t *testing.T) {
	s := newStore(t)
	user, _ := s.Remember(RememberInput{Content: "Lives in Austin", Key: "user.location"})
	guess, _ := s.Remember(RememberInput{Content: "Lives in Denver", Key: "user.location", Source: SourceAgent})
	third, _ := s.Remember(RememberInput{Content: "Works remotely"})
	if guess.Memory.Status != StatusConflict {
		t.Fatalf("guess = %s", guess.Memory.Status)
	}
	if _, _, err := s.Link(user.Memory.ID, guess.Memory.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.Link(user.Memory.ID, third.Memory.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Forget(user.Memory.ID); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{guess.Memory.ID, third.Memory.ID} {
		got, _ := s.Get(id)
		if got.Status != StatusActive || got.ConflictsWith != "" || len(got.Links) != 0 {
			t.Fatalf("after forgetting its partner: %+v", got)
		}
	}
	// The trash copy keeps its own links, and a reload from disk agrees.
	s2, _ := Open(s.Dir())
	got, _ := s2.Get(guess.Memory.ID)
	if len(got.Links) != 0 || got.Status != StatusActive {
		t.Fatalf("on disk: %+v", got)
	}
}

// After forget then restore, links must be symmetric again or gone on both
// sides; a one-way link is a half-broken state.
func TestQARestoreLeavesNoOneWayLinks(t *testing.T) {
	s := newStore(t)
	a, _ := s.Remember(RememberInput{Content: "The March outage came from a bad config push"})
	b, _ := s.Remember(RememberInput{Content: "Every change needs a canary"})
	s.Link(a.Memory.ID, b.Memory.ID)
	s.Forget(a.Memory.ID)
	back, err := s.Restore(a.Memory.ID)
	if err != nil {
		t.Fatal(err)
	}
	other, _ := s.Get(b.Memory.ID)
	if hasID(back.Links, b.Memory.ID) != hasID(other.Links, a.Memory.ID) {
		t.Fatalf("one-way link after restore: a.links=%v b.links=%v", back.Links, other.Links)
	}
}

// What the restore-as-conflict rule does when the user restores their own
// fact over an AI guess that took the key while it was in the trash.
func TestQARestoreUserFactOverAGuess(t *testing.T) {
	s := newStore(t)
	mine, _ := s.Remember(RememberInput{Content: "Lives in Austin", Key: "user.location"})
	s.Forget(mine.Memory.ID)
	guess, _ := s.Remember(RememberInput{Content: "Lives in Denver", Key: "user.location", Source: SourceAgent})
	back, err := s.Restore(mine.Memory.ID)
	if err != nil {
		t.Fatal(err)
	}
	g, _ := s.Get(guess.Memory.ID)
	t.Logf("restored user fact: %s; agent guess: %s", back.Status, g.Status)
	// Never two current facts for one key.
	if back.Status == StatusActive && g.Status == StatusActive {
		t.Fatal("two current facts for one key")
	}
	// An agent can't settle it in its own favor.
	if _, _, err := s.ResolveAs(guess.Memory.ID, SourceAgent); err == nil {
		t.Fatal("an agent resolved a conflict against the user's fact")
	}
}

func TestQARestoreStaleKeyedFact(t *testing.T) {
	s := newStore(t)
	old, _ := s.Remember(RememberInput{Content: "Lives in Austin", Key: "loc"})
	s.Forget(old.Memory.ID)
	s.Remember(RememberInput{Content: "Lives in Denver", Key: "loc"})
	// The trash copy was stale when forgotten.
	path := filepath.Join(s.trashDir(), old.Memory.ID+".md")
	data, _ := os.ReadFile(path)
	os.WriteFile(path, []byte(strings.Replace(string(data), "status: active", "status: stale", 1)), 0o600)
	if _, err := s.Restore(old.Memory.ID); err != nil {
		t.Fatal(err)
	}
	live := 0
	mems, _ := s.List(true)
	for _, m := range mems {
		if m.Key == "loc" && m.Status != StatusConflict && m.Live() {
			live++
		}
	}
	if live > 1 {
		t.Fatalf("%d live, unsettled facts for one key after restoring a stale one", live)
	}
}

func TestQASupersedeGuards(t *testing.T) {
	s := newStore(t)
	a, _ := s.Remember(RememberInput{Content: "A"})
	b, _ := s.Remember(RememberInput{Content: "B"})
	c, _ := s.Remember(RememberInput{Content: "C"})
	if _, err := s.Supersede(a.Memory.ID, a.Memory.ID); err == nil {
		t.Fatal("self supersede")
	}
	s.Supersede(a.Memory.ID, b.Memory.ID)
	s.Supersede(b.Memory.ID, c.Memory.ID)
	// A three-step cycle: c by a, where a is already replaced.
	if _, err := s.Supersede(c.Memory.ID, a.Memory.ID); err == nil {
		t.Fatal("a cycle a->b->c->a was allowed")
	}
	h, err := s.History(a.Memory.ID)
	if err != nil || len(h) != 3 {
		t.Fatalf("history = %d %v", len(h), err)
	}
	if _, err := s.Supersede(a.Memory.ID, "nope"); err == nil {
		t.Fatal("missing target")
	}
}

func TestQAProblemsClearAfterFix(t *testing.T) {
	s := newStore(t)
	bad := filepath.Join(s.Dir(), "broken.md")
	os.WriteFile(bad, []byte("---\nid: x\nno end"), 0o600)
	if len(s.Problems()) != 1 {
		t.Fatalf("problems = %v", s.Problems())
	}
	os.Remove(bad)
	if p := s.Problems(); len(p) != 0 {
		t.Fatalf("problems after fix = %v", p)
	}
	// Non-.md files are never problems.
	os.WriteFile(filepath.Join(s.Dir(), "notes.txt"), []byte("hi"), 0o600)
	if p := s.Problems(); len(p) != 0 {
		t.Fatalf("problems = %v", p)
	}
}

// --- G1: types ---

func TestQATypeFilter(t *testing.T) {
	pref := &Memory{Type: TypePreference, Status: StatusActive}
	tagged := &Memory{Tags: []string{"procedure"}, Status: StatusActive}
	none := &Memory{Status: StatusActive}
	if !(Scope{Types: []string{"Preferences"}}.Normalized()).Matches(pref) {
		t.Fatal("plural/case not accepted")
	}
	if !(Scope{Types: []string{"procedure"}}.Normalized()).Matches(tagged) {
		t.Fatal("a type from tags didn't match")
	}
	if (Scope{Types: []string{"procedure"}}.Normalized()).Matches(none) {
		t.Fatal("an untyped memory matched a type filter")
	}
	if !(Scope{Types: []string{"bogus", "preference"}}.Normalized()).Matches(pref) {
		t.Fatal("one unknown type hid a known one")
	}
}

// A filter naming only types Woolly doesn't know must not quietly widen to
// "every type". An agent asking for ["facts"] would otherwise get everything.
func TestQAUnknownTypeFilterMatchesNothing(t *testing.T) {
	m := &Memory{Type: TypePreference, Status: StatusActive}
	if (Scope{Types: []string{"bogus"}}.Normalized()).Matches(m) {
		t.Fatal(`Scope{Types: ["bogus"]} matches a preference: an unknown type filter widens to every type`)
	}
}

// --- S1: links ---

func TestQALinkGuards(t *testing.T) {
	s := newStore(t)
	a, _ := s.Remember(RememberInput{Content: "A"})
	b, _ := s.Remember(RememberInput{Content: "B"})
	if _, _, err := s.Link(a.Memory.ID, a.Memory.ID); err == nil {
		t.Fatal("linked to itself")
	}
	if _, _, err := s.Link(a.Memory.ID, "nope"); err == nil {
		t.Fatal("linked to nothing")
	}
	s.Link(a.Memory.ID, b.Memory.ID)
	s.Link(b.Memory.ID, a.Memory.ID)
	ga, _ := s.Get(a.Memory.ID)
	gb, _ := s.Get(b.Memory.ID)
	if len(ga.Links) != 1 || len(gb.Links) != 1 {
		t.Fatalf("links doubled: %v %v", ga.Links, gb.Links)
	}
	// Agent-scoped and session-scoped memories don't mix with the user's.
	ag, _ := s.Remember(RememberInput{Content: "C", Agent: "bot"})
	if _, _, err := s.Link(a.Memory.ID, ag.Memory.ID); err == nil {
		t.Fatal("linked across agent scope")
	}
}

func TestQALinksHeaderGarbage(t *testing.T) {
	s := newStore(t)
	a, _ := s.Remember(RememberInput{Content: "A"})
	b, _ := s.Remember(RememberInput{Content: "B"})
	rewriteField(t, s, a.Memory.ID, "links", " , ../../etc/passwd, "+b.Memory.ID+",,  not an id ; drop table, "+strings.Repeat("x", 300))
	got, err := s.Get(a.Memory.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Links) != 1 || got.Links[0] != b.Memory.ID {
		t.Fatalf("links = %q", got.Links)
	}
	if len(s.Problems()) != 0 {
		t.Fatalf("problems = %v", s.Problems())
	}
}

// Linking two of the user's memories in different projects is allowed by
// Link today. Recording the behavior so a change is deliberate.
func TestQALinkAcrossProjects(t *testing.T) {
	s := newStore(t)
	a, _ := s.Remember(RememberInput{Content: "A", Project: "alpha"})
	b, _ := s.Remember(RememberInput{Content: "B", Project: "beta"})
	_, _, err := s.Link(a.Memory.ID, b.Memory.ID)
	t.Logf("link across projects: err=%v", err)
}

// Replace keeps sources, type, tags and commit so an edited fact stays
// grounded (F3). Links are the same kind of tie: an edited fact shouldn't
// silently lose what it was linked to.
func TestQALinksSurviveReplace(t *testing.T) {
	s := newStore(t)
	a, _ := s.Remember(RememberInput{Content: "Deploys need a canary"})
	b, _ := s.Remember(RememberInput{Content: "The March outage came from a bad push"})
	s.Link(a.Memory.ID, b.Memory.ID)
	res, err := s.Replace(a.Memory.ID, RememberInput{Content: "Deploys need a canary and a rollback plan"})
	if err != nil {
		t.Fatal(err)
	}
	gb, _ := s.Get(b.Memory.ID)
	if !hasID(res.Memory.Links, b.Memory.ID) || !hasID(gb.Links, res.Memory.ID) {
		t.Fatalf("links lost on Replace: new.links=%v other.links=%v", res.Memory.Links, gb.Links)
	}
}

func TestQALinksSurviveMerge(t *testing.T) {
	s := newStore(t)
	a, _ := s.Remember(RememberInput{Content: "Uses pnpm"})
	b, _ := s.Remember(RememberInput{Content: "Uses pnpm workspaces"})
	c, _ := s.Remember(RememberInput{Content: "CI caches the pnpm store"})
	s.Link(b.Memory.ID, c.Memory.ID)
	res, err := s.Merge([]string{a.Memory.ID, b.Memory.ID}, RememberInput{Content: "Uses pnpm with workspaces", Source: SourceUser})
	if err != nil {
		t.Fatal(err)
	}
	if !hasID(res.Memory.Links, c.Memory.ID) {
		t.Fatalf("links lost on Merge: %v", res.Memory.Links)
	}
}

// --- S2: review queue ---

func TestQAReviewQueueFilters(t *testing.T) {
	s := newStore(t)
	guess, _ := s.Remember(RememberInput{Content: "Uses Postgres", Source: SourceAgent})
	imp, _ := s.Remember(RememberInput{Content: "Uses Redis", Source: SourceImport})
	mine, _ := s.Remember(RememberInput{Content: "Lives in Denver"})
	agentScoped, _ := s.Remember(RememberInput{Content: "Bot likes JSON", Source: SourceAgent, Agent: "bot"})
	session, _ := s.Remember(RememberInput{Content: "Session note", Source: SourceAgent, Session: "r1"})
	proj, _ := s.Remember(RememberInput{Content: "Builds with make", Source: SourceAgent, Project: "acme"})
	confl1, _ := s.Remember(RememberInput{Content: "Editor is vim", Key: "editor"})
	confl2, _ := s.Remember(RememberInput{Content: "Editor is emacs", Key: "editor", Source: SourceAgent})
	_ = confl1
	later := time.Now().Add(ReviewAge + time.Hour)
	q, err := s.ReviewQueue(later)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, m := range q {
		got[m.ID] = true
	}
	if !got[guess.Memory.ID] || !got[proj.Memory.ID] {
		t.Fatalf("missing unscoped guesses: %v", got)
	}
	for name, id := range map[string]string{"import": imp.Memory.ID, "user": mine.Memory.ID, "agent-scoped": agentScoped.Memory.ID,
		"session": session.Memory.ID, "conflict": confl2.Memory.ID} {
		if got[id] {
			t.Fatalf("%s memory in the review queue", name)
		}
	}
	// A use long ago doesn't keep it out; a recent one does.
	u := s.Usage()
	u[guess.Memory.ID] = Usage{Last: time.Now().Add(-200 * 24 * time.Hour), Uses: 1}
	s.writeUsage(u)
	if q, _ := s.ReviewQueue(later); len(q) != 2 {
		t.Fatalf("old use kept it out: %d", len(q))
	}
}

func TestQAReviewQueueOldestFirst(t *testing.T) {
	s := newStore(t)
	a, _ := s.Remember(RememberInput{Content: "Newer guess", Source: SourceAgent})
	b, _ := s.Remember(RememberInput{Content: "Older guess", Source: SourceAgent})
	rewriteField(t, s, a.Memory.ID, "updated_at", time.Now().Add(-100*24*time.Hour).UTC().Format(time.RFC3339Nano))
	rewriteField(t, s, b.Memory.ID, "updated_at", time.Now().Add(-200*24*time.Hour).UTC().Format(time.RFC3339Nano))
	q, _ := s.ReviewQueue(time.Now())
	if len(q) != 2 || q[0].ID != b.Memory.ID {
		t.Fatalf("queue order: %+v", q)
	}
}
