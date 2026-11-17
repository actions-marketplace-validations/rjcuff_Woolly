package ground

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rjcuff/woolly/internal/memory"
)

const imuC = `#include "pins.h"

// Reads the accelerometer.
int16_t imu_read_accel(int axis) {
    int16_t raw = spi_read(IMU_CS, 0x3B + axis * 2);
    return raw;
}

static void imu_reset(void) {
    spi_write(IMU_CS, 0x6B, 0x80);
}
`

const pinsH = `#ifndef PINS_H
#define PINS_H

#define IMU_CS GPIO_NUM_5
#define MOTOR_PWM_HZ 20000

typedef enum {
    MOTOR_FL = 0,
    MOTOR_FR,
    MOTOR_BL,
    MOTOR_BR,
} motor_t;

#endif
`

func run(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func write(t *testing.T, dir, name, text string) {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(name))
	os.MkdirAll(filepath.Dir(p), 0o755)
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

// drone makes a small firmware repo and returns it with its first commit.
func drone(t *testing.T) (string, string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	run(t, dir, "init", "-q", "-b", "main")
	write(t, dir, "main/imu.c", imuC)
	write(t, dir, "main/pins.h", pinsH)
	run(t, dir, "add", ".")
	run(t, dir, "commit", "-q", "-m", "first")
	return dir, run(t, dir, "rev-parse", "--short", "HEAD")
}

func commit(t *testing.T, dir, msg string) string {
	t.Helper()
	run(t, dir, "add", "-A")
	run(t, dir, "commit", "-q", "-m", msg)
	return run(t, dir, "rev-parse", "--short", "HEAD")
}

// fact is a memory saved at commit c with resolved sources.
func fact(t *testing.T, dir, c, text string, refs ...string) memory.Memory {
	t.Helper()
	var parsed []memory.Ref
	for _, r := range refs {
		ref, ok := memory.ParseRef(r)
		if !ok {
			t.Fatalf("bad ref %q", r)
		}
		parsed = append(parsed, ref)
	}
	got := Open(dir).Resolve(parsed)
	if len(got) != len(parsed) {
		t.Fatalf("refs %v resolved to %v", refs, got)
	}
	return memory.Memory{ID: "f1", Content: text, Commit: c, Sources: got, CreatedAt: time.Now()}
}

func TestDefsC(t *testing.T) {
	defs := Defs("x.c", []byte(imuC+pinsH))
	for _, want := range []string{"imu_read_accel", "imu_reset", "IMU_CS", "MOTOR_PWM_HZ", "MOTOR_FL", "MOTOR_BR", "motor_t", "PINS_H"} {
		if _, ok := Find(defs, want); !ok {
			t.Errorf("missing %s in %+v", want, defs)
		}
	}
	d, _ := Find(defs, "imu_read_accel")
	if d.Start != 4 || d.End != 7 {
		t.Errorf("imu_read_accel spans %d-%d, want 4-7", d.Start, d.End)
	}
	for _, d := range defs {
		if d.Name == "spi_read" || d.Name == "return" || d.Name == "raw" {
			t.Errorf("a call or local was taken for a definition: %+v", d)
		}
	}
}

func TestDefsOtherLanguages(t *testing.T) {
	cases := map[string][2]string{
		"store.go":  {"package x\n\ntype Store struct{ n int }\n\nfunc (s *Store) Remember(x int) error {\n\treturn nil\n}\n\nconst Limit = 5\n", "Store.Remember,Remember,Store,Limit"},
		"pid.py":    {"KP = 1.2\n\nclass Pid:\n    def step(self, err):\n        return err * KP\n\ndef tune():\n    pass\n", "KP,Pid,step,tune"},
		"api.ts":    {"export const LIMIT = 20;\nexport function recall(q: string) {\n  return q;\n}\nexport class Client {\n  async search(q: string): Promise<void> {\n    return;\n  }\n}\ninterface Opts { n: number }\n", "LIMIT,recall,Client,search,Opts"},
		"motor.rs":  {"pub const MAX_RPM: u32 = 9000;\npub fn spin(rpm: u32) {\n}\nstruct Motor { pin: u8 }\n", "MAX_RPM,spin,Motor"},
		"motor.cpp": {"class Motor {\npublic:\n  void spin(int rpm);\n};\n\nvoid Motor::spin(int rpm) {\n  set(rpm);\n}\n", "Motor,Motor::spin,spin"},
	}
	for name, c := range cases {
		defs := Defs(name, []byte(c[0]))
		for _, want := range strings.Split(c[1], ",") {
			if _, ok := Find(defs, want); !ok {
				t.Errorf("%s: missing %s in %+v", name, want, defs)
			}
		}
	}
}

func TestFingerprintIgnoresCommentsAndSpacing(t *testing.T) {
	a := "int f(void) {\n  return 1; // one\n}\n"
	b := "int f(void)  {\n\treturn 1; /* still one */\n}\n"
	if Fingerprint("a.c", a) != Fingerprint("a.c", b) {
		t.Fatal("comments and spacing changed the fingerprint")
	}
	if Fingerprint("a.c", a) == Fingerprint("a.c", strings.Replace(a, "1", "2", 1)) {
		t.Fatal("a code change kept the fingerprint")
	}
}

func check(t *testing.T, dir string, m memory.Memory) Result {
	t.Helper()
	return Open(dir).Check(m)
}

func TestUnchangedCodeIsOK(t *testing.T) {
	dir, c := drone(t)
	m := fact(t, dir, c, "imu_read_accel returns raw counts", "main/imu.c#imu_read_accel")
	if r := check(t, dir, m); r.Verdict != OK || r.Update {
		t.Fatalf("got %+v", r)
	}
}

func TestRenameNamedInFactIsStale(t *testing.T) {
	dir, c := drone(t)
	m := fact(t, dir, c, "imu_read_accel returns raw counts; divide by 16384 for g", "main/imu.c#imu_read_accel")
	write(t, dir, "main/imu.c", strings.ReplaceAll(imuC, "imu_read_accel", "imu_read_accel_g"))
	broke := commit(t, dir, "rename")
	r := check(t, dir, m)
	if r.Verdict != Stale || r.Reason != "main/imu.c: imu_read_accel was renamed to imu_read_accel_g" || r.Commit != broke {
		t.Fatalf("got %+v (want commit %s)", r, broke)
	}
	if r.Refs[0].Symbol != "imu_read_accel_g" {
		t.Fatalf("source should follow the rename: %+v", r.Refs)
	}
}

func TestRenameNotNamedInFactJustMoves(t *testing.T) {
	dir, c := drone(t)
	m := fact(t, dir, c, "The accelerometer read takes an axis from 0 to 2", "main/imu.c#imu_read_accel")
	write(t, dir, "main/imu.c", strings.ReplaceAll(imuC, "imu_read_accel", "accel_read"))
	commit(t, dir, "rename")
	r := check(t, dir, m)
	if r.Verdict != Moved || !r.Update || r.Refs[0].Symbol != "accel_read" {
		t.Fatalf("got %+v", r)
	}
}

func TestDeletedFileIsStale(t *testing.T) {
	dir, c := drone(t)
	m := fact(t, dir, c, "IMU reset writes 0x80 to register 0x6B", "main/imu.c#imu_reset")
	os.Remove(filepath.Join(dir, "main/imu.c"))
	broke := commit(t, dir, "drop imu")
	r := check(t, dir, m)
	if r.Verdict != Stale || r.Reason != "main/imu.c was deleted" || r.Commit != broke {
		t.Fatalf("got %+v", r)
	}
}

func TestRenamedFileMoves(t *testing.T) {
	dir, c := drone(t)
	m := fact(t, dir, c, "IMU reset writes 0x80", "main/imu.c#imu_reset")
	run(t, dir, "mv", "main/imu.c", "main/imu_spi.c")
	commit(t, dir, "move")
	r := check(t, dir, m)
	if r.Verdict != Moved || r.Refs[0].Path != "main/imu_spi.c" {
		t.Fatalf("got %+v", r)
	}
}

func TestChangedValueIsStale(t *testing.T) {
	dir, c := drone(t)
	m := fact(t, dir, c, "The IMU chip select is GPIO 5", "main/pins.h#IMU_CS")
	write(t, dir, "main/pins.h", strings.Replace(pinsH, "GPIO_NUM_5", "GPIO_NUM_4", 1))
	broke := commit(t, dir, "move cs")
	r := check(t, dir, m)
	if r.Verdict != Stale || r.Reason != "main/pins.h#IMU_CS: 5 is no longer there" || r.Commit != broke {
		t.Fatalf("got %+v", r)
	}
}

func TestLineRangeValueChangeIsStale(t *testing.T) {
	dir, c := drone(t)
	m := fact(t, dir, c, "Motor PWM runs at 20000 Hz", "main/pins.h:5")
	write(t, dir, "main/pins.h", strings.Replace(pinsH, "20000", "24000", 1))
	commit(t, dir, "faster pwm")
	if r := check(t, dir, m); r.Verdict != Stale || !strings.Contains(r.Reason, "20000 is no longer there") {
		t.Fatalf("got %+v", r)
	}
}

func TestShiftedLinesMove(t *testing.T) {
	dir, c := drone(t)
	m := fact(t, dir, c, "Motor PWM runs at 20000 Hz", "main/pins.h:5")
	write(t, dir, "main/pins.h", strings.Replace(pinsH, "#define IMU_CS", "// Pins for the v2 board.\n#define LED GPIO_NUM_2\n#define IMU_CS", 1))
	commit(t, dir, "led")
	r := check(t, dir, m)
	if r.Verdict != Moved || r.Refs[0].Start != 7 {
		t.Fatalf("got %+v", r)
	}
}

func TestCommentOnlyEditIsOK(t *testing.T) {
	dir, c := drone(t)
	m := fact(t, dir, c, "imu_read_accel returns raw counts", "main/imu.c#imu_read_accel")
	write(t, dir, "main/imu.c", strings.Replace(imuC, "    return raw;", "    // Raw counts, not g.\n    return raw;", 1))
	commit(t, dir, "comment")
	if r := check(t, dir, m); r.Verdict != OK {
		t.Fatalf("got %+v", r)
	}
}

func TestBodyChangeKeepingValuesIsChanged(t *testing.T) {
	dir, c := drone(t)
	m := fact(t, dir, c, "imu_read_accel returns raw counts", "main/imu.c#imu_read_accel")
	write(t, dir, "main/imu.c", strings.Replace(imuC, "    return raw;", "    if (raw == INT16_MIN) raw++;\n    return raw;", 1))
	commit(t, dir, "clamp")
	if r := check(t, dir, m); r.Verdict != Changed {
		t.Fatalf("got %+v", r)
	}
}

func TestFunctionMovedToAnotherFile(t *testing.T) {
	dir, c := drone(t)
	m := fact(t, dir, c, "IMU reset writes 0x80 to register 0x6B", "main/imu.c#imu_reset")
	reset := "static void imu_reset(void) {\n    spi_write(IMU_CS, 0x6B, 0x80);\n}\n"
	write(t, dir, "main/imu.c", strings.Replace(imuC, reset, "", 1))
	write(t, dir, "main/imu_init.c", "#include \"pins.h\"\n\n"+reset)
	commit(t, dir, "split")
	r := check(t, dir, m)
	if r.Verdict != Moved || r.Refs[0].Path != "main/imu_init.c" {
		t.Fatalf("got %+v", r)
	}
}

func TestRemovedSymbolIsStale(t *testing.T) {
	dir, c := drone(t)
	m := fact(t, dir, c, "Call imu_reset before reading", "main/imu.c#imu_reset")
	write(t, dir, "main/imu.c", strings.Replace(imuC, "static void imu_reset(void) {\n    spi_write(IMU_CS, 0x6B, 0x80);\n}\n", "", 1))
	commit(t, dir, "drop reset")
	if r := check(t, dir, m); r.Verdict != Stale || r.Reason != "main/imu.c: imu_reset is gone" {
		t.Fatalf("got %+v", r)
	}
}

func TestUncommittedChangeHasNoCommit(t *testing.T) {
	dir, c := drone(t)
	m := fact(t, dir, c, "The IMU chip select is GPIO 5", "main/pins.h#IMU_CS")
	write(t, dir, "main/pins.h", strings.Replace(pinsH, "GPIO_NUM_5", "GPIO_NUM_4", 1))
	if r := check(t, dir, m); r.Verdict != Stale || r.Commit != "" {
		t.Fatalf("got %+v", r)
	}
}

func TestResolveDropsMissingSources(t *testing.T) {
	dir, _ := drone(t)
	refs := Open(dir).Resolve([]memory.Ref{{Path: "main/imu.c", Symbol: "imu_reset"}, {Path: "main/nope.c"}, {Path: "main/imu.c", Symbol: "no_such_fn"}, {Path: "main/pins.h", Start: 900}})
	if len(refs) != 1 || refs[0].Hash == "" {
		t.Fatalf("got %+v", refs)
	}
}

func TestSuggestFindsSources(t *testing.T) {
	dir, _ := drone(t)
	r := Open(dir)
	got := r.Suggest("Call imu_reset in main/imu.c before the first read")
	if len(got) != 1 || got[0].Path != "main/imu.c" || got[0].Symbol != "imu_reset" {
		t.Fatalf("with a path: %+v", got)
	}
	got = r.Suggest("MOTOR_PWM_HZ must stay above 16000")
	if len(got) != 1 || got[0].Path != "main/pins.h" || got[0].Symbol != "MOTOR_PWM_HZ" {
		t.Fatalf("by name: %+v", got)
	}
	if got := r.Suggest("The team prefers small pull requests"); len(got) != 0 {
		t.Fatalf("prose: %+v", got)
	}
}

func TestFactValues(t *testing.T) {
	got := strings.Join(factValues("The IMU CS is GPIO 5, Kp is 1.2, and `spi_read` uses SPI2 at 0x3B"), ",")
	if got != "5,1.2,spi_read,SPI2,0x3B" {
		t.Fatalf("got %s", got)
	}
}
