package main

import (
	"strings"
	"testing"

	"github.com/rjcuff/woolly/internal/dream"
)

func TestRecapShowsOncePerProject(t *testing.T) {
	t.Setenv("WOOLLY_HOME", t.TempDir())
	t.Setenv("WOOLLY_NO_EMBED", "1")
	a, err := open()
	if err != nil {
		t.Fatal(err)
	}
	if s := a.awaySection("drone"); s != "" {
		t.Fatalf("no dream yet: %q", s)
	}

	// A dream that did nothing leaves no recap.
	a.saveRecap(&recapState{})
	if s := a.awaySection("drone"); s != "" {
		t.Fatalf("empty dream: %q", s)
	}

	r := &recapState{}
	r.add(&dream.Report{Merged: 4, Settled: 1}, "")
	r.add(&dream.Report{Steps: []dream.Step{{Action: "learn", Memory: "Run make check before every commit"}}}, "drone")
	r.add(&dream.Report{Steps: []dream.Step{{Action: "learn", Memory: "The site deploys from main"}}}, "web")
	a.saveRecap(r)

	s := a.awaySection("drone")
	for _, want := range []string{"# While you were away", "Woolly dreamed overnight. It folded 4 repeats together and settled 1 fact that disagreed.", "It learned 1 lesson from recent sessions:", "- Run make check before every commit"} {
		if !strings.Contains(s, want) {
			t.Fatalf("recap is missing %q:\n%s", want, s)
		}
	}
	if strings.Contains(s, "deploys from main") {
		t.Fatal("another project's lesson showed up")
	}
	if again := a.awaySection("drone"); again != "" {
		t.Fatalf("the same recap showed twice: %q", again)
	}
	if web := a.awaySection("web"); !strings.Contains(web, "The site deploys from main") {
		t.Fatalf("each project sees it once: %q", web)
	}
}
