package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/rjcuff/woolly/internal/decision"
	"github.com/rjcuff/woolly/internal/memory"
	"github.com/rjcuff/woolly/internal/why"
)

func TestWhyFindsFactsAndDecisions(t *testing.T) {
	a, dir := droneApp(t)
	ref, _ := memory.ParseRef("main/imu.c#imu_read_accel")
	if _, err := a.store.Remember(memory.RememberInput{
		Content: decision.Format("Read the IMU over SPI", "I2C was too slow at 1 kHz", []string{"I2C: too slow"}),
		Tags:    []string{decision.Tag}, Project: "drone", Via: "Claude Code", Sources: groundRefs(dir, []memory.Ref{ref}),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.store.Remember(memory.RememberInput{Content: "Keep imu.c free of RTOS calls", Project: "drone"}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.store.Remember(memory.RememberInput{Content: "Prefers short answers"}); err != nil {
		t.Fatal(err)
	}
	mems, _ := a.store.List(true)

	// Line 2 is inside imu_read_accel.
	ans := why.Find(mems, "drone", dir, why.Target{Path: "main/imu.c", Line: 2})
	if ans.In != "imu_read_accel" || len(ans.Matches) != 3 {
		t.Fatalf("line 2: %+v", ans)
	}
	first := ans.Matches[0]
	if first.Choice != "Read the IMU over SPI" || first.Why != "I2C was too slow at 1 kHz" || first.Rejected != "I2C: too slow" || first.Still != "true" {
		t.Fatalf("decisions come first, split into parts: %+v", first)
	}
	if ans.Matches[2].How != "mentions" {
		t.Fatalf("a fact that only names the file comes last: %+v", ans.Matches[2])
	}
	text := why.Text(ans)
	for _, want := range []string{"main/imu.c:2, in imu_read_accel", "Decided", "in Claude Code: Read the IMU over SPI.", "Rejected: I2C: too slow.", "Still true at"} {
		if !strings.Contains(text, want) {
			t.Fatalf("text is missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "short answers") {
		t.Fatal("personal facts never show up")
	}

	// pins.h has its own fact and nothing else.
	ans = why.Find(mems, "drone", dir, why.Target{Path: "main/pins.h"})
	if len(ans.Matches) != 1 || ans.Matches[0].Memory.Content != "The IMU chip select is GPIO 5" {
		t.Fatalf("pins.h: %+v", ans.Matches)
	}

	// After the code changes and a check runs, the answer says so.
	breakBoth(t, dir)
	if _, err := a.checkRepo(dir, "drone", checkOptions{}); err != nil {
		t.Fatal(err)
	}
	mems, _ = a.store.List(true)
	ans = why.Find(mems, "drone", dir, why.Target{Path: "main/pins.h"})
	if len(ans.Matches) != 1 || ans.Matches[0].Still != "stale" || !strings.Contains(ans.Matches[0].Reason, "5 is no longer there") {
		t.Fatalf("stale pins.h: %+v", ans.Matches)
	}
	if !strings.Contains(why.Text(ans), "Stale: ") {
		t.Fatal("stale facts say they're stale")
	}

	// Nothing known: a helpful line, not an empty answer.
	ans = why.Find(mems, "drone", dir, why.Target{Path: "main/radio.c"})
	if len(ans.Matches) != 0 || !strings.Contains(why.Text(ans), "doesn't know anything about main/radio.c yet") {
		t.Fatalf("unknown file: %+v", ans)
	}
}

func TestWhyParseTarget(t *testing.T) {
	root := filepath.FromSlash("/repo")
	for _, c := range []struct {
		in, cwd, path, sym string
		line               int
	}{
		{"main/imu.c", "/repo", "main/imu.c", "", 0},
		{"imu.c:12", "/repo/main", "main/imu.c", "", 12},
		{"main/imu.c#imu_read_accel", "/repo", "main/imu.c", "imu_read_accel", 0},
	} {
		got, err := why.ParseTarget(c.in, root, filepath.FromSlash(c.cwd))
		if err != nil || got.Path != c.path || got.Line != c.line || got.Symbol != c.sym {
			t.Fatalf("%q from %s: %+v %v", c.in, c.cwd, got, err)
		}
	}
	if _, err := why.ParseTarget("../other/x.go", root, root); err == nil {
		t.Fatal("paths outside the repo are refused")
	}
}
