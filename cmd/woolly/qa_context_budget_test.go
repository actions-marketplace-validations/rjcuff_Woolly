package main

// QA tests for F5 brief budgets, written against
// plans/memory-growth-and-fix.md: "brief trims by section priority, logs
// only shown IDs".

import (
	"fmt"
	"strings"
	"testing"

	"github.com/rjcuff/woolly/internal/memory"
)

// Every ID the brief logs as used must be in the text it prints.
func checkLoggedShown(t *testing.T, out string, ids []string, mems []memory.Memory) {
	t.Helper()
	byID := map[string]memory.Memory{}
	for _, m := range mems {
		byID[m.ID] = m
	}
	for _, id := range ids {
		m, ok := byID[id]
		if !ok {
			t.Fatalf("logged unknown id %s", id)
		}
		if !strings.Contains(out, strings.TrimSpace(m.Content)) {
			t.Fatalf("id %s logged as used but its text was cut from the brief", id)
		}
	}
}

func TestQABriefWhenProfileAloneExceedsLimit(t *testing.T) {
	var mems []memory.Memory
	for i := 0; i < 400; i++ {
		mems = append(mems, memory.Memory{ID: fmt.Sprintf("me%03d", i), Status: memory.StatusActive,
			Content: fmt.Sprintf("Personal fact %d that is long enough to fill a good part of the brief", i)})
	}
	for i := 0; i < 5; i++ {
		mems = append(mems, memory.Memory{ID: fmt.Sprintf("p%d", i), Project: "acme", Status: memory.StatusActive,
			Content: fmt.Sprintf("acme note %d", i)})
	}
	out, ids := buildContext(mems, "acme", "", "", "")
	if len(out) > contextLimit+200 {
		t.Fatalf("brief is %d characters", len(out))
	}
	for i := 0; i < 5; i++ {
		if !strings.Contains(out, fmt.Sprintf("acme note %d", i)) {
			t.Fatalf("a huge profile pushed out project notes")
		}
	}
	checkLoggedShown(t, out, ids, mems)
}

func TestQABriefManyDecisionsLogsOnlyShown(t *testing.T) {
	var mems []memory.Memory
	for i := 0; i < 300; i++ {
		mems = append(mems, memory.Memory{ID: fmt.Sprintf("d%03d", i), Project: "acme", Status: memory.StatusActive, Tags: []string{"decision"},
			Content: fmt.Sprintf("Decision %03d: we chose option A over option B because of reason number %d", i, i)})
	}
	out, ids := buildContext(mems, "acme", "", "", "")
	if len(out) > contextLimit+200 {
		t.Fatalf("brief is %d characters", len(out))
	}
	checkLoggedShown(t, out, ids, mems)
}

func TestQABriefExactlyFitsAndEmpty(t *testing.T) {
	out, ids := buildContext(nil, "acme", "", "", "")
	if len(ids) != 0 || len(out) > contextLimit {
		t.Fatalf("empty brief: %d ids, %d chars", len(ids), len(out))
	}
	mems := []memory.Memory{{ID: "one", Project: "acme", Status: memory.StatusActive, Content: "acme uses make"}}
	out, ids = buildContext(mems, "acme", "", "", "")
	if len(ids) != 1 || !strings.Contains(out, "acme uses make") {
		t.Fatalf("one note: %v\n%s", ids, out)
	}
}
