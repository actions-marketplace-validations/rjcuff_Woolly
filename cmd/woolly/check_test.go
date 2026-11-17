package main

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rjcuff/woolly/internal/ground"
	"github.com/rjcuff/woolly/internal/memory"
	"github.com/rjcuff/woolly/internal/share"
)

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func writeFile(t *testing.T, dir, name, text string) {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(name))
	os.MkdirAll(filepath.Dir(p), 0o755)
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

// droneApp makes a firmware repo, a Woolly with its own home, and two
// grounded facts saved at the first commit.
func droneApp(t *testing.T) (*app, string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	t.Setenv("WOOLLY_HOME", t.TempDir())
	t.Setenv("WOOLLY_NO_EMBED", "1")
	dir := filepath.Join(t.TempDir(), "drone")
	os.MkdirAll(dir, 0o755)
	gitRun(t, dir, "init", "-q", "-b", "main")
	writeFile(t, dir, "main/imu.c", "int imu_read_accel(int axis) {\n    return spi_read(0x3B + axis);\n}\n")
	writeFile(t, dir, "main/pins.h", "#define IMU_CS GPIO_NUM_5\n")
	gitRun(t, dir, "add", ".")
	gitRun(t, dir, "commit", "-q", "-m", "first")
	a, err := open()
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range []struct{ text, ref string }{
		{"imu_read_accel returns raw counts", "main/imu.c#imu_read_accel"},
		{"The IMU chip select is GPIO 5", "main/pins.h#IMU_CS"},
	} {
		ref, _ := memory.ParseRef(f.ref)
		if _, err := a.store.Remember(memory.RememberInput{Content: f.text, Project: "drone", Commit: commitAt(t, dir), Sources: groundRefs(dir, []memory.Ref{ref})}); err != nil {
			t.Fatal(err)
		}
	}
	return a, dir
}

func commitAt(t *testing.T, dir string) string {
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--short=12", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func breakBoth(t *testing.T, dir string) {
	writeFile(t, dir, "main/imu.c", "int imu_read_accel_g(int axis) {\n    return spi_read(0x3B + axis);\n}\n")
	writeFile(t, dir, "main/pins.h", "#define IMU_CS GPIO_NUM_4\n")
	gitRun(t, dir, "commit", "-q", "-am", "rename and move cs")
}

func TestCheckRepoMarksStaleOnce(t *testing.T) {
	a, dir := droneApp(t)
	rep, err := a.checkRepo(dir, "drone", checkOptions{backfill: true})
	if err != nil || rep.Checked != 2 || rep.NewStale != 0 {
		t.Fatalf("clean repo: %+v %v", rep, err)
	}
	breakBoth(t, dir)
	rep, err = a.checkRepo(dir, "drone", checkOptions{backfill: true})
	if err != nil || rep.NewStale != 2 || rep.Results[0].Verdict != ground.Stale {
		t.Fatalf("after the change: %+v %v", rep, err)
	}
	stale, _ := a.store.Stale("drone")
	if len(stale) != 2 {
		t.Fatalf("stale facts: %+v", stale)
	}
	for _, m := range stale {
		if m.StaleCommit == "" || m.StaleReason == "" {
			t.Fatalf("missing reason or commit: %+v", m)
		}
		if strings.Contains(m.Content, "imu_read_accel") && m.Sources[0].Symbol != "imu_read_accel_g" {
			t.Fatalf("source should follow the rename: %+v", m.Sources)
		}
	}
	rep, _ = a.checkRepo(dir, "drone", checkOptions{})
	if rep.NewStale != 0 || rep.Waiting != 2 {
		t.Fatalf("second run: %+v", rep)
	}
}

func TestCheckBackfillsSources(t *testing.T) {
	a, dir := droneApp(t)
	res, _ := a.store.Remember(memory.RememberInput{Content: "IMU_CS is the chip select for the IMU", Project: "drone"})
	rep, _ := a.checkRepo(dir, "drone", checkOptions{backfill: true})
	if rep.Backfilled != 1 {
		t.Fatalf("backfill: %+v", rep)
	}
	m, _ := a.store.Get(res.Memory.ID)
	if len(m.Sources) != 1 || m.Sources[0].Symbol != "IMU_CS" || m.Sources[0].Hash == "" || m.Commit == "" {
		t.Fatalf("got %+v", m)
	}
}

func TestReviewLoopEditsAndRetires(t *testing.T) {
	a, dir := droneApp(t)
	breakBoth(t, dir)
	a.checkRepo(dir, "drone", checkOptions{})
	stale, _ := a.store.Stale("drone")
	rv := reviewer{a: a, repo: share.Find(dir)}
	in := "e\nimu_read_accel_g returns acceleration in g\nr\n"
	// Stale facts come newest first; make the edit go to the accel fact.
	if !strings.Contains(stale[0].Content, "imu_read_accel") {
		in = "r\ne\nimu_read_accel_g returns acceleration in g\n"
	}
	rv.loop(bufio.NewReader(strings.NewReader(in)), stale)

	if left, _ := a.store.Stale("drone"); len(left) != 0 {
		t.Fatalf("still stale: %+v", left)
	}
	live, _ := a.store.List(false)
	var edited *memory.Memory
	for i, m := range live {
		if m.Content == "imu_read_accel_g returns acceleration in g" {
			edited = &live[i]
		}
		if m.Content == "The IMU chip select is GPIO 5" {
			t.Fatal("retired fact is still current")
		}
	}
	if edited == nil || len(edited.Sources) != 1 || edited.Sources[0].Symbol != "imu_read_accel_g" || edited.Source != memory.SourceUser {
		t.Fatalf("edited: %+v", edited)
	}
	if rep, _ := a.checkRepo(dir, "drone", checkOptions{}); rep.NewStale != 0 {
		t.Fatalf("the edited fact went stale again: %+v", rep)
	}
}

func TestConfirmMakesItCurrentAgain(t *testing.T) {
	a, dir := droneApp(t)
	breakBoth(t, dir)
	a.checkRepo(dir, "drone", checkOptions{})
	stale, _ := a.store.Stale("drone")
	rv := reviewer{a: a, repo: share.Find(dir)}
	rv.loop(bufio.NewReader(strings.NewReader("c\nc\n")), stale)
	if left, _ := a.store.Stale("drone"); len(left) != 0 {
		t.Fatalf("still stale: %+v", left)
	}
	if rep, _ := a.checkRepo(dir, "drone", checkOptions{}); rep.NewStale != 0 {
		t.Fatalf("confirmed facts went stale again right away: %+v", rep)
	}
}

func TestStaleLine(t *testing.T) {
	a, dir := droneApp(t)
	if line := a.staleLine("drone"); line != "" {
		t.Fatalf("nothing stale yet: %q", line)
	}
	breakBoth(t, dir)
	a.checkRepo(dir, "drone", checkOptions{})
	if line := a.staleLine("drone"); line != "2 facts went stale since last session. Run woolly review, and don't rely on them until then." {
		t.Fatalf("first brief: %q", line)
	}
	if line := a.staleLine("drone"); line != "2 stale facts are waiting for review. Run woolly review." {
		t.Fatalf("next brief: %q", line)
	}
	out, _ := buildContext(nil, "drone", dir, "x went stale")
	if !strings.Contains(out, "# Stale facts\n\nx went stale") {
		t.Fatalf("brief: %s", out)
	}
}

func TestPostCommitHook(t *testing.T) {
	_, dir := droneApp(t)
	p := filepath.Join(dir, ".git", "hooks", "post-commit")
	os.WriteFile(p, []byte("#!/bin/sh\necho mine\n"), 0o755)
	if changed, err := installHook(dir); err != nil || !changed || !hasHook(dir) {
		t.Fatalf("install: %v %v", changed, err)
	}
	if changed, _ := installHook(dir); changed {
		t.Fatal("installing twice changed the hook")
	}
	b, _ := os.ReadFile(p)
	if !strings.HasPrefix(string(b), "#!/bin/sh\necho mine\n") || !strings.Contains(string(b), "check --hook >/dev/null 2>&1 &") {
		t.Fatalf("hook:\n%s", b)
	}
	if changed, _ := removeHook(dir); !changed {
		t.Fatal("remove did nothing")
	}
	if b, _ := os.ReadFile(p); string(b) != "#!/bin/sh\necho mine\n" {
		t.Fatalf("after remove:\n%s", b)
	}
	os.Remove(p)
	installHook(dir)
	removeHook(dir)
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Fatal("a hook with only Woolly's lines should be deleted")
	}
}
