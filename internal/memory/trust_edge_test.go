package memory

import (
	"errors"
	"testing"
)

// Edge cases for the trust boundary (F1) and grounding (F3) beyond the
// happy paths in trust_test.go.

func TestReplaceExplicitSourcesOverrideOldOnes(t *testing.T) {
	s := newStore(t)
	old, err := s.Remember(RememberInput{Content: "LED is on pin 4", Project: "bot", Sources: []Ref{{Path: "main/pins.h"}}})
	if err != nil {
		t.Fatal(err)
	}
	res, err := s.Replace(old.Memory.ID, RememberInput{Content: "LED is on pin 5", Sources: []Ref{{Path: "main/led.c"}}})
	if err != nil {
		t.Fatal(err)
	}
	if got := res.Memory.Sources; len(got) != 1 || got[0].Path != "main/led.c" {
		t.Fatalf("explicit sources should replace the old ones, got %+v", got)
	}
}

func TestReplaceExplicitTypeAndConfidenceWin(t *testing.T) {
	s := newStore(t)
	old, _ := s.Remember(RememberInput{Content: "LED is on pin 4", Project: "bot", Type: "hardware", Confidence: 0.9, Commit: "abc"})
	res, err := s.Replace(old.Memory.ID, RememberInput{Content: "LED pin 4 is flaky", Type: "gotcha", Confidence: 0.3, Commit: "def"})
	if err != nil {
		t.Fatal(err)
	}
	m := res.Memory
	if m.Type != "gotcha" || m.Conf != 0.3 || m.Commit != "def" {
		t.Fatalf("explicit fields lost: type=%q conf=%v commit=%q", m.Type, m.Conf, m.Commit)
	}
}

func TestReplaceKeepsOldVersionAsHistory(t *testing.T) {
	s := newStore(t)
	old, _ := s.Remember(RememberInput{Content: "LED is on pin 4", Project: "bot", Sources: []Ref{{Path: "pins.h"}}})
	res, err := s.Replace(old.Memory.ID, RememberInput{Content: "LED is on pin 5"})
	if err != nil {
		t.Fatal(err)
	}
	prev, _ := s.Get(old.Memory.ID)
	if prev.Status != StatusSuperseded || prev.SupersededBy != res.Memory.ID {
		t.Fatalf("old version: status=%s by=%s", prev.Status, prev.SupersededBy)
	}
	if len(prev.Sources) != 1 {
		t.Fatalf("old version lost its sources: %+v", prev.Sources)
	}
	h, err := s.History(res.Memory.ID)
	if err != nil || len(h) != 2 {
		t.Fatalf("history = %d versions, %v", len(h), err)
	}
}

func TestReplaceUntrustedInstructionsLeavesOldActive(t *testing.T) {
	s := newStore(t)
	old, _ := s.Remember(RememberInput{Content: "Prefers tea"})
	_, err := s.Replace(old.Memory.ID, RememberInput{Content: attack, Source: SourceUser, Untrusted: true})
	var inj *InjectionError
	if !errors.As(err, &inj) {
		t.Fatalf("err = %v, want InjectionError", err)
	}
	if got, _ := s.Get(old.Memory.ID); got.Status != StatusActive || got.Content != "Prefers tea" {
		t.Fatalf("old memory changed: %+v", got)
	}
	if all, _ := s.List(true); len(all) != 1 {
		t.Fatalf("%d memories on disk, want 1", len(all))
	}
}

func TestMergeWithMissingMemoryChangesNothing(t *testing.T) {
	s := newStore(t)
	a, _ := s.Remember(RememberInput{Content: "IMU on bus 0", Project: "bot", Sources: []Ref{{Path: "imu.c"}}, Source: SourceAgent})
	_, err := s.Merge([]string{a.Memory.ID, "20990101-000000-deadbeef"}, RememberInput{Content: "The IMU is on bus 0", Source: SourceAgent})
	if err == nil {
		t.Fatal("merge with a missing memory should fail")
	}
	if got, _ := s.Get(a.Memory.ID); got.Status != StatusActive {
		t.Fatalf("first memory status = %s, want active (no partial merge)", got.Status)
	}
	if all, _ := s.List(true); len(all) != 1 {
		t.Fatalf("%d memories on disk, want 1", len(all))
	}
}

func TestMergeDedupesSharedSourcesAndKeepsExplicitOnes(t *testing.T) {
	s := newStore(t)
	a, _ := s.Remember(RememberInput{Content: "IMU on bus 0", Project: "bot", Sources: []Ref{{Path: "main/imu.c"}}, Source: SourceAgent, Tags: []string{"imu"}})
	b, _ := s.Remember(RememberInput{Content: "IMU uses bus zero", Project: "bot", Sources: []Ref{{Path: "main/imu.c"}}, Source: SourceAgent, Tags: []string{"i2c"}})
	res, err := s.Merge([]string{a.Memory.ID, b.Memory.ID}, RememberInput{Content: "The IMU is on I2C bus 0", Source: SourceAgent,
		Sources: []Ref{{Path: "main/board.h"}}})
	if err != nil {
		t.Fatal(err)
	}
	count := map[string]int{}
	for _, r := range res.Memory.Sources {
		count[r.Path]++
	}
	if count["main/imu.c"] != 1 || count["main/board.h"] != 1 || len(res.Memory.Sources) != 2 {
		t.Fatalf("sources = %+v", res.Memory.Sources)
	}
	if !hasTag(res.Memory.Tags, "imu") || !hasTag(res.Memory.Tags, "i2c") {
		t.Fatalf("tags = %v", res.Memory.Tags)
	}
	for _, id := range []string{a.Memory.ID, b.Memory.ID} {
		if got, _ := s.Get(id); got.Status != StatusSuperseded || got.SupersededBy != res.Memory.ID {
			t.Fatalf("%s: status=%s by=%s", id, got.Status, got.SupersededBy)
		}
	}
}

func TestResolveAsKeylessPairAgainstUserIsRefused(t *testing.T) {
	for _, userPointsAtAgent := range []bool{false, true} {
		s := newStore(t)
		user, _ := s.Remember(RememberInput{Content: "Lives in Austin", Source: SourceUser})
		agent, _ := s.Remember(RememberInput{Content: "Lives in Denver", Source: SourceAgent})
		var err error
		if userPointsAtAgent {
			_, err = s.MarkConflict(user.Memory.ID, agent.Memory.ID)
		} else {
			_, err = s.MarkConflict(agent.Memory.ID, user.Memory.ID)
		}
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err := s.ResolveAs(agent.Memory.ID, SourceAgent); !errors.Is(err, ErrOutranked) {
			t.Fatalf("userPointsAtAgent=%v: err = %v, want ErrOutranked", userPointsAtAgent, err)
		}
		if got, _ := s.Get(user.Memory.ID); !got.Live() || got.Status == StatusSuperseded {
			t.Fatalf("user memory was replaced: %s", got.Status)
		}
	}
}

func TestResolveAsAgentMayPickTheUsersVersion(t *testing.T) {
	s := newStore(t)
	user, _ := s.Remember(RememberInput{Content: "Lives in Austin", Key: "user.location", Source: SourceUser})
	agent, _ := s.Remember(RememberInput{Content: "Lives in Denver", Key: "user.location", Source: SourceAgent})
	if agent.Memory.Status != StatusConflict {
		t.Fatalf("agent status = %s", agent.Memory.Status)
	}
	kept, replaced, err := s.ResolveAs(user.Memory.ID, SourceAgent)
	if err != nil {
		t.Fatalf("picking the more trusted memory should be allowed: %v", err)
	}
	if kept.ID != user.Memory.ID || kept.Status != StatusActive || len(replaced) != 1 || replaced[0].ID != agent.Memory.ID {
		t.Fatalf("kept=%+v replaced=%+v", kept, replaced)
	}
}

func TestResolveAsAgentCannotPickAgentOverImport(t *testing.T) {
	s := newStore(t)
	imp, _ := s.Remember(RememberInput{Content: "Works at Acme", Key: "user.job", Source: SourceImport})
	agent, _ := s.Remember(RememberInput{Content: "Works at Globex", Key: "user.job", Source: SourceAgent})
	if agent.Memory.Status != StatusConflict {
		t.Fatalf("agent status = %s", agent.Memory.Status)
	}
	if _, _, err := s.ResolveAs(agent.Memory.ID, SourceAgent); !errors.Is(err, ErrOutranked) {
		t.Fatalf("err = %v, want ErrOutranked", err)
	}
	if got, _ := s.Get(imp.Memory.ID); got.Status != StatusActive {
		t.Fatalf("import status = %s", got.Status)
	}
}

func TestResolveAsIgnoresSameKeyInAnotherSlot(t *testing.T) {
	s := newStore(t)
	// Alice's (app) memory with the same key must not block the user's own agents.
	alice, _ := s.Remember(RememberInput{Content: "Alice uses Postgres", Key: "db", Source: SourceUser, User: "alice"})
	a, _ := s.Remember(RememberInput{Content: "Uses SQLite", Key: "db", Source: SourceAgent})
	b, _ := s.Remember(RememberInput{Content: "Uses MySQL", Source: SourceAgent})
	if _, err := s.MarkConflict(b.Memory.ID, a.Memory.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.ResolveAs(a.Memory.ID, SourceAgent); err != nil {
		t.Fatalf("a different slot should not block: %v", err)
	}
	if got, _ := s.Get(alice.Memory.ID); got.Status != StatusActive {
		t.Fatalf("alice's memory status = %s, want untouched", got.Status)
	}
}

func TestResolveAsStaleUserFactBlocksAgent(t *testing.T) {
	s := newStore(t)
	user, _ := s.Remember(RememberInput{Content: "Deploys to Fly", Key: "deploy", Source: SourceUser})
	if _, err := s.MarkStale(user.Memory.ID, StaleInput{Reason: "file gone"}); err != nil {
		t.Fatal(err)
	}
	agent, _ := s.Remember(RememberInput{Content: "Deploys to Vercel", Key: "deploy", Source: SourceAgent})
	if _, _, err := s.ResolveAs(agent.Memory.ID, SourceAgent); !errors.Is(err, ErrOutranked) {
		t.Fatalf("err = %v, want ErrOutranked", err)
	}
}

func TestResolveAsUnknownIDIsNotFound(t *testing.T) {
	s := newStore(t)
	if _, _, err := s.ResolveAs("20990101-000000-deadbeef", SourceAgent); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestUntrustedUpdateWithPlainTextSaves(t *testing.T) {
	s := newStore(t)
	// An app's own memory: the user's are covered by TestUntrustedUpdateOfAUserMemory.
	res, _ := s.Remember(RememberInput{Content: "Prefers tea", Source: SourceAgent})
	c := "Prefers green tea"
	m, err := s.Update(res.Memory.ID, UpdateInput{Content: &c, Untrusted: true})
	if err != nil {
		t.Fatalf("plain text should save: %v", err)
	}
	if m.Content != c {
		t.Fatalf("content = %q", m.Content)
	}
	// An untrusted edit that leaves the text alone is fine too.
	key := "drink"
	if _, err := s.Update(res.Memory.ID, UpdateInput{Key: &key, Untrusted: true}); err != nil {
		t.Fatalf("key-only edit: %v", err)
	}
}

func TestTrustedUpdateKeepsInstructionLikeText(t *testing.T) {
	s := newStore(t)
	res, _ := s.Remember(RememberInput{Content: "Prefers tea"})
	c := attack
	if _, err := s.Update(res.Memory.ID, UpdateInput{Content: &c}); err != nil {
		t.Fatalf("the user's own edit should save: %v", err)
	}
}

func TestUntrustedStillRefusesSecrets(t *testing.T) {
	s := newStore(t)
	_, err := s.Remember(RememberInput{Content: "my key is sk-ant-abcdefghijklmnopqrstuvwxyz0123456789", Untrusted: true})
	var sens *SensitiveError
	if !errors.As(err, &sens) {
		t.Fatalf("err = %v, want SensitiveError", err)
	}
}

func TestPersonalScopeWithProject(t *testing.T) {
	s := newStore(t)
	s.Remember(RememberInput{Content: "Acme uses Postgres", Project: "acme"})
	s.Remember(RememberInput{Content: "Alice's acme note", Project: "acme", User: "alice"})
	s.Remember(RememberInput{Content: "Agent bot note", Agent: "bot"})
	s.Remember(RememberInput{Content: "Run note", Session: "run1"})
	s.Remember(RememberInput{Content: "Likes tea"})
	got, err := s.ListIn(Scope{Project: "acme", Personal: true}, false)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, m := range got {
		seen[m.Content] = true
	}
	if !seen["Acme uses Postgres"] || !seen["Likes tea"] || seen["Alice's acme note"] || seen["Agent bot note"] || seen["Run note"] {
		t.Fatalf("personal+project scope = %v", seen)
	}
}

// When the user rewrites an agent's guess (woolly review "edit" passes
// Source: user and no confidence), the new version is the user's own
// statement. Copying the agent's low saved confidence makes Confidence()
// report the user's fact as 0.3 instead of the user default of 1.
func TestReplaceByUserDoesNotInheritAgentConfidence(t *testing.T) {
	s := newStore(t)
	old, _ := s.Remember(RememberInput{Content: "Maybe uses Postgres", Source: SourceAgent, Confidence: 0.3})
	res, err := s.Replace(old.Memory.ID, RememberInput{Content: "Uses Postgres 16", Source: SourceUser})
	if err != nil {
		t.Fatal(err)
	}
	if c := res.Memory.Confidence(); c < 1 {
		t.Fatalf("user's rewrite has confidence %v (conf=%v copied from the agent's guess)", c, res.Memory.Conf)
	}
}

func TestUntrustedUpdateOfAUserMemory(t *testing.T) {
	s := newStore(t)
	res, _ := s.Remember(RememberInput{Content: "Prefers tea", Key: "drink"})
	c := "Prefers coffee"
	m, err := s.Update(res.Memory.ID, UpdateInput{Content: &c, Untrusted: true})
	if err != nil || m.ID == res.Memory.ID || m.Status != StatusConflict || m.ConflictsWith != res.Memory.ID {
		t.Fatalf("untrusted text should wait as a conflict: %+v %v", m, err)
	}
	if got, _ := s.Get(res.Memory.ID); got.Content != "Prefers tea" || got.Status != StatusActive || got.Source != SourceUser {
		t.Fatalf("user memory changed: %+v", got)
	}
	key := "breakfast"
	if _, err := s.Update(res.Memory.ID, UpdateInput{Key: &key, Untrusted: true}); !errors.Is(err, ErrOutranked) {
		t.Fatalf("untrusted key change: %v", err)
	}
	same := "Prefers tea"
	if m, err := s.Update(res.Memory.ID, UpdateInput{Content: &same, Untrusted: true}); err != nil || m.ID != res.Memory.ID {
		t.Fatalf("same text: %+v %v", m, err)
	}
}
