package recall

// QA tests for F5 (budgets), G3 (ranking), G4 (near copies) and S1 (links),
// written against plans/memory-growth-and-fix.md.

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/rjcuff/woolly/internal/memory"
)

func res(content string) Result { return Result{Memory: memory.Memory{Content: content}} }

func TestQAFitExactlyAtBudget(t *testing.T) {
	rs := []Result{res("abc"), res("de")} // 3+3 and 2+3
	if got := Fit(rs, 6); len(got) != 1 {
		t.Fatalf("exactly at budget: %d kept", len(got))
	}
	if got := Fit(rs, 11); len(got) != 2 {
		t.Fatalf("exactly both: %d kept", len(got))
	}
	if got := Fit(rs, 10); len(got) != 1 {
		t.Fatalf("one under: %d kept", len(got))
	}
	if got := Fit(rs, 5); len(got) != 1 || got[0].Memory.Content != "de" {
		t.Fatalf("too small for the first: %v", got)
	}
	if got := Fit(rs, -1); len(got) != 2 {
		t.Fatal("negative budget keeps everything")
	}
	// Whitespace around the text doesn't count against the budget.
	if got := Fit([]Result{res("   abc \n\n")}, 6); len(got) != 1 {
		t.Fatal("padding counted")
	}
	// Order is kept: Fit never reorders.
	rs = []Result{res("ccc"), res("a"), res("bb")}
	got := Fit(rs, 100)
	for i := range rs {
		if got[i].Memory.Content != rs[i].Memory.Content {
			t.Fatalf("reordered: %v", got)
		}
	}
}

func TestQAFitLinesExactlyAtBudget(t *testing.T) {
	lines := []string{"abcd", "ef"} // 5 and 3
	if got := FitLines(lines, 5); len(got) != 1 || got[0] != 0 {
		t.Fatalf("exactly at budget: %v", got)
	}
	if got := FitLines(lines, 8); len(got) != 2 {
		t.Fatalf("exactly both: %v", got)
	}
	if got := FitLines(lines, 4); len(got) != 1 || got[0] != 1 {
		t.Fatalf("skip the long one: %v", got)
	}
	if got := FitLines(lines, 0); len(got) != 2 {
		t.Fatal("no budget keeps everything")
	}
	if got := FitLines(nil, 10); len(got) != 0 {
		t.Fatal("empty in, empty out")
	}
}

func TestQAFollowLinksNeverExceedsLimit(t *testing.T) {
	e, store := newEngine(t)
	var hubs []string
	for i := 0; i < 3; i++ {
		h, _ := store.Remember(memory.RememberInput{Content: fmt.Sprintf("Kubernetes cluster %d runs the api", i)})
		hubs = append(hubs, h.Memory.ID)
		for j := 0; j < 4; j++ {
			o, _ := store.Remember(memory.RememberInput{Content: fmt.Sprintf("Unrelated note %d-%d about lunch", i, j)})
			if _, _, err := store.Link(h.Memory.ID, o.Memory.ID); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, limit := range []int{1, 2, 5} {
		got, _, _ := e.Recall(context.Background(), "kubernetes cluster", limit, false)
		if len(got) > limit {
			t.Fatalf("limit %d, got %d", limit, len(got))
		}
	}
}

func TestQAFollowLinksSkipsSupersededAndOutOfScope(t *testing.T) {
	e, store := newEngine(t)
	hub, _ := store.Remember(memory.RememberInput{Content: "Payments service talks to Stripe"})
	old, _ := store.Remember(memory.RememberInput{Content: "Webhook secret rotation is monthly", Key: "rotation"})
	store.Link(hub.Memory.ID, old.Memory.ID)
	store.Remember(memory.RememberInput{Content: "Webhook secret rotation is weekly", Key: "rotation"})
	// An app user's memory, never linkable to the user's own.
	alice, _ := store.Remember(memory.RememberInput{Content: "Alice's card is declined often", User: "alice"})
	proj, _ := store.Remember(memory.RememberInput{Content: "Beta project deploys on Fridays", Project: "beta"})
	store.Link(hub.Memory.ID, proj.Memory.ID) // allowed across projects
	got, _, _ := e.RecallScoped(context.Background(), "payments stripe", 10, false, memory.Scope{Personal: true, Project: "alpha"})
	for _, r := range got {
		switch r.Memory.ID {
		case old.Memory.ID:
			t.Fatal("a replaced memory came in through a link")
		case alice.Memory.ID:
			t.Fatal("an app user's memory came in through a link")
		case proj.Memory.ID:
			t.Fatal("another project's memory came in through a link")
		}
	}
	// With includeAll, the replaced version may come along, at most once.
	all, _, _ := e.Recall(context.Background(), "payments stripe", 10, true)
	seen := map[string]int{}
	for _, r := range all {
		seen[r.Memory.ID]++
		if seen[r.Memory.ID] > 1 {
			t.Fatalf("%s listed twice", r.Memory.ID)
		}
	}
}

func TestQALinkedMatchKeepsBestScore(t *testing.T) {
	e, store := newEngine(t)
	strong, _ := store.Remember(memory.RememberInput{Content: "Redis cache Redis cluster Redis eviction"})
	weak, _ := store.Remember(memory.RememberInput{Content: "The cache layer was added in 2024 after a slow quarter with many other changes to infra"})
	store.Remember(memory.RememberInput{Content: "Filler about lunch"})
	before, _, _ := e.Recall(context.Background(), "redis cache", 5, false)
	var weakBefore float64
	for _, r := range before {
		if r.Memory.ID == weak.Memory.ID {
			weakBefore = r.Score
		}
	}
	store.Link(strong.Memory.ID, weak.Memory.ID)
	after, _, _ := e.Recall(context.Background(), "redis cache", 5, false)
	n := 0
	for _, r := range after {
		if r.Memory.ID == weak.Memory.ID {
			n++
			if r.Score < weakBefore-0.001 {
				t.Fatalf("linking lowered a match: %v -> %v", weakBefore, r.Score)
			}
			if r.Score < after[0].Score*0.5-0.002 {
				t.Fatalf("linked score %v below half of %v", r.Score, after[0].Score)
			}
		}
	}
	if n != 1 {
		t.Fatalf("weak match listed %d times", n)
	}
	// Results stay sorted.
	for i := 1; i < len(after); i++ {
		if after[i].Score > after[i-1].Score {
			t.Fatalf("not sorted: %+v", after)
		}
	}
}

func TestQALinkToMissingMemoryIsIgnored(t *testing.T) {
	e, store := newEngine(t)
	a, _ := store.Remember(memory.RememberInput{Content: "Grafana dashboards live in ops repo"})
	b, _ := store.Remember(memory.RememberInput{Content: "Lunch is at noon"})
	store.Link(a.Memory.ID, b.Memory.ID)
	// b's file vanishes (deleted by hand, not forgotten).
	if err := os.Remove(store.Dir() + "/" + b.Memory.ID + ".md"); err != nil {
		t.Fatal(err)
	}
	got, _, err := e.Recall(context.Background(), "grafana", 5, false)
	if err != nil || len(got) != 1 {
		t.Fatalf("got %v err %v", got, err)
	}
}

func TestQANearDupSkippingKeepsLimit(t *testing.T) {
	e, store := newEngine(t)
	for i := 0; i < 5; i++ {
		store.Remember(memory.RememberInput{Content: "Prefers dark mode in the editor", Key: fmt.Sprintf("k%d", i)})
	}
	for _, s := range []string{"Editor font is Fira Code", "Editor tab width is 4", "Editor theme syncs with OS", "Editor autosaves on blur"} {
		store.Remember(memory.RememberInput{Content: s})
	}
	got, _, _ := e.Recall(context.Background(), "editor", 4, false)
	if len(got) != 4 {
		t.Fatalf("got %d results, want 4 (enough distinct memories exist)", len(got))
	}
	dark := 0
	for _, r := range got {
		if strings.Contains(r.Memory.Content, "dark mode") {
			dark++
		}
	}
	if dark > 1 {
		t.Fatalf("%d copies of the same fact", dark)
	}
}

// NearCopy ignores stopwords, and "not" and "no" are stopwords, so a fact
// and its negation look like copies. MCP remember then answers "Already
// known" and drops the correction.
func TestQANearCopyNegation(t *testing.T) {
	pairs := [][2]string{
		{"The app uses Postgres", "The app does not use Postgres"},
		{"Deploys are allowed on Friday", "Deploys are not allowed on Friday"},
		{"Has a dog", "Has no dog"},
	}
	for _, p := range pairs {
		if NearCopy(p[0], p[1]) {
			t.Errorf("NearCopy(%q, %q) = true; a negation is not a copy", p[0], p[1])
		}
	}
}

// Recall's own near-copy skipping has the same blind spot: the negation is
// dropped from results.
func TestQARecallKeepsNegation(t *testing.T) {
	e, store := newEngine(t)
	store.Remember(memory.RememberInput{Content: "Deploys are allowed on Friday", Key: "a"})
	store.Remember(memory.RememberInput{Content: "Deploys are not allowed on Friday", Key: "b"})
	got, _, _ := e.Recall(context.Background(), "deploys friday", 5, false)
	if len(got) != 2 {
		t.Fatalf("got %d results; the negated fact was skipped as a copy: %+v", len(got), got)
	}
}

func TestQAScorePartsAreConsistent(t *testing.T) {
	e, store := newEngine(t)
	store.Remember(memory.RememberInput{Content: "Uses Terraform for infra", Source: memory.SourceAgent, Confidence: 0.3})
	got, _, _ := e.Recall(context.Background(), "terraform", 5, false)
	if len(got) != 1 || got[0].Parts == nil {
		t.Fatalf("got %+v", got)
	}
	p := got[0].Parts
	want := p.Keyword * p.Fresh * p.Confidence * p.Use * p.Status
	if d := want - got[0].Score; d > 0.01 || d < -0.01 {
		t.Fatalf("score %v vs parts product %v: %+v", got[0].Score, want, p)
	}
	if p.Confidence < 0.8 || p.Confidence > 1 || p.Use < 0.9 || p.Use > 1.1 {
		t.Fatalf("multipliers out of range: %+v", p)
	}
}

// The plan's G3 asks for an absolute relevance floor: BM25 is normalized so
// the top match is always 1.0, and a query that shares one incidental word
// with a memory still returns it.
func TestQARelevanceFloor(t *testing.T) {
	e, store := newEngine(t)
	store.Remember(memory.RememberInput{Content: "The build uses make"})
	got, _, _ := e.Recall(context.Background(), "how do I make chocolate cake frosting recipe dessert", 5, false)
	if len(got) != 0 {
		t.Fatalf("a one-word incidental match was returned with score %v (keyword %v)", got[0].Score, got[0].Parts.Keyword)
	}
}

func TestQATypeFilterInRecall(t *testing.T) {
	e, store := newEngine(t)
	store.Remember(memory.RememberInput{Content: "Deploy by running make ship", Type: "procedure"})
	store.Remember(memory.RememberInput{Content: "Likes deploy logs short", Type: "preference"})
	store.Remember(memory.RememberInput{Content: "Deploy broke on Monday", Type: "episode"})
	got, _, _ := e.RecallScoped(context.Background(), "deploy", 5, false, memory.Scope{Types: []string{"procedure", "episode"}})
	if len(got) != 2 {
		t.Fatalf("got %d: %+v", len(got), got)
	}
	for _, r := range got {
		if memory.TypeOf(r.Memory) == "preference" {
			t.Fatal("filtered type came back")
		}
	}
}

func TestQAEvalFloorHoldsWithBadFeedback(t *testing.T) {
	store, err := memory.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]string{}
	for _, m := range evalMemories {
		res, err := store.Remember(memory.RememberInput{Content: m.content})
		if err != nil {
			t.Fatal(err)
		}
		ids[res.Memory.ID] = m.key
	}
	// Every memory gets down-voted a lot: feedback is a nudge, so recall must
	// still find the right answers.
	for id := range ids {
		for i := 0; i < 5; i++ {
			store.Feedback(id, false)
		}
	}
	e := New(store, nil, "")
	hits := 0
	for _, q := range evalQueries {
		results, _, _ := e.Recall(context.Background(), q.query, 5, false)
		for _, r := range results {
			if ids[r.Memory.ID] == q.want {
				hits++
				break
			}
		}
	}
	if score := float64(hits) / float64(len(evalQueries)); score < evalFloor {
		t.Fatalf("recall@5 %.2f below the floor with feedback", score)
	}
}
