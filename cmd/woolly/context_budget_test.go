package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/rjcuff/woolly/internal/memory"
)

func TestBusyProjectKeepsTheProfile(t *testing.T) {
	var mems []memory.Memory
	for i := 0; i < 500; i++ {
		mems = append(mems, memory.Memory{ID: fmt.Sprintf("p%03d", i), Project: "acme", Status: memory.StatusActive,
			Content: fmt.Sprintf("acme note number %d about some part of the build that matters", i)})
	}
	mems = append(mems, memory.Memory{ID: "me", Content: "Prefers short answers", Status: memory.StatusActive})
	out, ids := buildContext(mems, "acme", "", "", "")
	if len(out) > contextLimit+200 {
		t.Fatalf("brief is %d characters", len(out))
	}
	if !strings.Contains(out, "Prefers short answers") {
		t.Fatalf("the profile was pushed out:\n%s", out[len(out)-500:])
	}
	// Every ID logged as used is actually in the brief.
	shown := 0
	for _, id := range ids {
		if id == "me" {
			continue
		}
		var n int
		fmt.Sscanf(id, "p%03d", &n)
		if !strings.Contains(out, fmt.Sprintf("acme note number %d about", n)) {
			t.Fatalf("id %s logged but not shown", id)
		}
		shown++
	}
	if shown == 0 || shown == 500 {
		t.Fatalf("%d project notes shown", shown)
	}
}
