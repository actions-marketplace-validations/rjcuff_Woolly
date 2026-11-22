package recall

import (
	"strings"
	"testing"

	"github.com/rjcuff/woolly/internal/memory"
)

func TestFitKeepsBestFirstWithinBudget(t *testing.T) {
	rs := []Result{
		{Memory: memory.Memory{Content: strings.Repeat("a", 50)}},
		{Memory: memory.Memory{Content: strings.Repeat("b", 500)}},
		{Memory: memory.Memory{Content: strings.Repeat("c", 30)}},
	}
	got := Fit(rs, 100)
	if len(got) != 2 || got[0].Memory.Content[0] != 'a' || got[1].Memory.Content[0] != 'c' {
		t.Fatalf("Fit = %v", got)
	}
	if len(Fit(rs, 0)) != 3 {
		t.Fatal("no budget keeps everything")
	}
	if len(Fit(nil, 100)) != 0 {
		t.Fatal("empty in, empty out")
	}
}

func TestFitLines(t *testing.T) {
	if got := FitLines([]string{"aaaa", "bbbbbbbbbb", "cc"}, 9); len(got) != 2 || got[0] != 0 || got[1] != 2 {
		t.Fatalf("FitLines = %v", got)
	}
}
