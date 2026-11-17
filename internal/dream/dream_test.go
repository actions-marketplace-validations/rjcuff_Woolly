package dream

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/rjcuff/woolly/internal/handoff"
	"github.com/rjcuff/woolly/internal/learn"
	"github.com/rjcuff/woolly/internal/memory"
	"github.com/rjcuff/woolly/internal/recall"
)

// scriptModel answers relate prompts from a table keyed by the two facts,
// lessons prompts with a fixed reply, and anything else as learn's decide
// step would: add.
type scriptModel struct {
	relations map[string]string // "A|B" (either order) -> relation JSON
	lessons   string
	fail      bool
	asked     int
}

func (s *scriptModel) Chat(_ context.Context, system, user string, _ any) (string, error) {
	if s.fail {
		return "", errors.New("connection refused")
	}
	switch system {
	case relatePrompt:
		s.asked++
		lines := strings.Split(user, "\n")
		a := strings.TrimPrefix(lines[0], "Fact A: ")
		b := strings.TrimPrefix(lines[1], "Fact B: ")
		if r, ok := s.relations[a+"|"+b]; ok {
			return r, nil
		}
		if r, ok := s.relations[b+"|"+a]; ok {
			return r, nil
		}
		return `{"relation":"unrelated"}`, nil
	case lessonsPrompt:
		return s.lessons, nil
	}
	return `{"event":"ADD","target":-1}`, nil
}

func setup(t *testing.T, m Model) (*Dreamer, *memory.Store) {
	t.Helper()
	store, err := memory.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	engine := recall.New(store, nil, "")
	d := &Dreamer{Store: store, Engine: engine}
	if m != nil {
		d.Model = m
		d.Learner = &learn.Learner{Store: store, Engine: engine, Model: m}
	}
	return d, store
}

func save(t *testing.T, s *memory.Store, in memory.RememberInput) memory.Memory {
	t.Helper()
	if in.Source == "" {
		in.Source = memory.SourceAgent
	}
	res, err := s.Remember(in)
	if err != nil {
		t.Fatal(err)
	}
	return res.Memory
}

func current(t *testing.T, s *memory.Store, sc memory.Scope) []string {
	t.Helper()
	mems, _ := s.ListIn(sc, false)
	var out []string
	for _, m := range mems {
		out = append(out, m.Content)
	}
	return out
}

func TestFoldsRepeatsWithoutAModel(t *testing.T) {
	d, s := setup(t, nil)
	said := save(t, s, memory.RememberInput{Content: "Uses pnpm", Source: memory.SourceUser})
	save(t, s, memory.RememberInput{Content: "uses pnpm."})
	save(t, s, memory.RememberInput{Content: "Uses pnpm", User: "alice"}) // another scope, left alone

	rep, err := d.Dream(context.Background(), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if rep.Merged != 1 || !strings.Contains(rep.Note, "no chat model") {
		t.Fatalf("report: %+v", rep)
	}
	if got := current(t, s, memory.Scope{}); len(got) != 2 {
		t.Fatalf("want your fact and alice's left: %v", got)
	}
	if m, _ := s.Get(said.ID); !m.Live() {
		t.Fatalf("what you said is the one kept: %+v", m)
	}
}

func TestMergesDuplicatesAndSettlesContradictions(t *testing.T) {
	m := &scriptModel{relations: map[string]string{
		"Deploys from main|Ships to production from the main branch": `{"relation":"duplicate","text":"Deploys to production from the main branch"}`,
		"Lives in Denver|Lives in Austin":                            `{"relation":"contradiction"}`,
	}}
	d, s := setup(t, m)
	sc := memory.Scope{User: "alice"}
	a := save(t, s, memory.RememberInput{Content: "Deploys from main", User: "alice"})
	b := save(t, s, memory.RememberInput{Content: "Ships to production from the main branch", User: "alice"})
	denver := save(t, s, memory.RememberInput{Content: "Lives in Denver", User: "alice"})
	austin := save(t, s, memory.RememberInput{Content: "Lives in Austin", User: "alice"})

	rep, err := d.Dream(context.Background(), Options{Scope: sc})
	if err != nil {
		t.Fatal(err)
	}
	if rep.Merged != 1 || rep.Settled != 1 || rep.Note != "" {
		t.Fatalf("report: %+v", rep)
	}
	got := strings.Join(current(t, s, sc), "|")
	if !strings.Contains(got, "Deploys to production from the main branch") || !strings.Contains(got, "Lives in Austin") || strings.Contains(got, "Denver") {
		t.Fatalf("current: %s", got)
	}
	for _, id := range []string{a.ID, b.ID, denver.ID} {
		if old, _ := s.Get(id); old.Live() || old.InvalidAt == nil {
			t.Fatalf("should be history now: %+v", old)
		}
	}
	if m, _ := s.Get(austin.ID); !m.Live() {
		t.Fatalf("the newer fact stays: %+v", m)
	}
}

func TestNeverRetiresWhatYouSaid(t *testing.T) {
	m := &scriptModel{relations: map[string]string{"Lives in Denver|Lives in Austin": `{"relation":"contradiction"}`}}
	d, s := setup(t, m)
	said := save(t, s, memory.RememberInput{Content: "Lives in Denver", Source: memory.SourceUser})
	guess := save(t, s, memory.RememberInput{Content: "Lives in Austin"})

	rep, _ := d.Dream(context.Background(), Options{})
	if rep.Conflicts != 1 {
		t.Fatalf("want a conflict: %+v", rep)
	}
	if m, _ := s.Get(said.ID); !m.Live() || m.Status != memory.StatusActive {
		t.Fatalf("what you said stands: %+v", m)
	}
	g, _ := s.Get(guess.ID)
	if g.Status != memory.StatusConflict || g.ConflictsWith != said.ID {
		t.Fatalf("the guess waits for review: %+v", g)
	}
	// Picking your fact settles it.
	if _, replaced, err := s.Resolve(said.ID); err != nil || len(replaced) != 1 || replaced[0].ID != guess.ID {
		t.Fatalf("resolve: %+v %v", replaced, err)
	}
}

func TestDryRunAndScopes(t *testing.T) {
	m := &scriptModel{relations: map[string]string{"Lives in Denver|Lives in Austin": `{"relation":"contradiction"}`}}
	d, s := setup(t, m)
	save(t, s, memory.RememberInput{Content: "Lives in Denver", User: "alice"})
	save(t, s, memory.RememberInput{Content: "Lives in Austin", User: "bob"})
	save(t, s, memory.RememberInput{Content: "Lives in Denver", User: "carol"})
	save(t, s, memory.RememberInput{Content: "Lives in Austin", User: "carol"})

	rep, _ := d.Dream(context.Background(), Options{DryRun: true})
	if rep.Settled != 1 || rep.Steps[0].IDs == nil {
		t.Fatalf("only carol's pair is in one scope: %+v", rep)
	}
	if n := len(current(t, s, memory.Scope{})); n != 4 {
		t.Fatalf("a dry run changes nothing: %d current", n)
	}
}

func TestLearnsLessonsFromSessions(t *testing.T) {
	lessons, _ := json.Marshal(map[string]any{"lessons": []map[string]any{
		{"lesson": "Run make check before every commit", "notes": []int{0, 2}},
		{"lesson": "Made up with no notes", "notes": []int{}},
		{"lesson": "Points at a note that isn't there", "notes": []int{42}},
	}})
	m := &scriptModel{lessons: string(lessons)}
	d, s := setup(t, m)
	p := "acme"
	save(t, s, memory.RememberInput{Content: "Fixed the login bug. make check caught a race.", Key: handoff.Key, Project: p})
	save(t, s, memory.RememberInput{Content: "Use SQLite instead of Postgres. Why: one file.", Tags: []string{"decision"}, Project: p})
	save(t, s, memory.RememberInput{Content: "Added billing. Ran make check before committing.", Key: handoff.Key, Project: p})

	rep, err := d.Dream(context.Background(), Options{Scope: memory.Scope{Project: p}, Learn: true})
	if err != nil {
		t.Fatal(err)
	}
	if rep.Learned != 1 || rep.Steps[len(rep.Steps)-1].Memory != "Run make check before every commit" {
		t.Fatalf("report: %+v", rep)
	}
	mems, _ := s.ListIn(memory.Scope{Project: p}, false)
	var lesson *memory.Memory
	for i := range mems {
		if mems[i].Content == "Run make check before every commit" {
			lesson = &mems[i]
		}
	}
	if lesson == nil || lesson.Project != p || len(lesson.Tags) != 1 || lesson.Tags[0] != "learning" || lesson.Metadata["learned_from"] == nil || lesson.Via != Via {
		t.Fatalf("lesson: %+v", lesson)
	}

	// Dreaming again doesn't write the same lesson twice.
	rep, _ = d.Dream(context.Background(), Options{Scope: memory.Scope{Project: p}, Learn: true})
	if rep.Learned != 0 {
		t.Fatalf("second dream: %+v", rep)
	}
}

func TestAModelThatStopsAnswering(t *testing.T) {
	m := &scriptModel{fail: true}
	d, s := setup(t, m)
	save(t, s, memory.RememberInput{Content: "Lives in Denver"})
	save(t, s, memory.RememberInput{Content: "Lives in Austin"})
	rep, err := d.Dream(context.Background(), Options{})
	if err != nil || !strings.Contains(rep.Note, "stopped answering") {
		t.Fatalf("want a note, not an error: %+v %v", rep, err)
	}
}

func TestRemembersUnrelatedPairs(t *testing.T) {
	m := &scriptModel{}
	d, s := setup(t, m)
	save(t, s, memory.RememberInput{Content: "Likes Go"})
	save(t, s, memory.RememberInput{Content: "Likes Go and Rust"})

	rep, _ := d.Dream(context.Background(), Options{})
	if m.asked != 1 || len(rep.Unrelated) != 1 {
		t.Fatalf("first dream: asked %d, %+v", m.asked, rep)
	}
	known := map[string]bool{rep.Unrelated[0]: true}
	d.Dream(context.Background(), Options{Unrelated: known})
	if m.asked != 1 {
		t.Fatalf("a known unrelated pair isn't asked again: asked %d", m.asked)
	}
}
