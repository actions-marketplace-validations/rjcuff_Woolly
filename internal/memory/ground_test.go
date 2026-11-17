package memory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseRef(t *testing.T) {
	cases := []struct {
		in   string
		want Ref
		ok   bool
	}{
		{"main/imu.c", Ref{Path: "main/imu.c"}, true},
		{"main/imu.c#imu_read_accel", Ref{Path: "main/imu.c", Symbol: "imu_read_accel"}, true},
		{"main/pins.h:12-18", Ref{Path: "main/pins.h", Start: 12, End: 18}, true},
		{"main/pins.h:12", Ref{Path: "main/pins.h", Start: 12}, true},
		{"main/pins.h:12-18@77aa01b3", Ref{Path: "main/pins.h", Start: 12, End: 18, Hash: "77aa01b3"}, true},
		{"./src/../src/a.go#Store.Remember@abcd", Ref{Path: "src/a.go", Symbol: "Store.Remember", Hash: "abcd"}, true},
		{`src\win.go`, Ref{Path: "src/win.go"}, true},
		{"/etc/passwd", Ref{}, false},
		{"../outside.c", Ref{}, false},
		{"C:/x.c", Ref{}, false},
		{"a.c:9-3", Ref{}, false},
		{"a.c#bad symbol", Ref{}, false},
		{"", Ref{}, false},
	}
	for _, c := range cases {
		got, ok := ParseRef(c.in)
		if ok != c.ok || got != c.want {
			t.Errorf("ParseRef(%q) = %+v, %v; want %+v, %v", c.in, got, ok, c.want, c.ok)
		}
		if ok {
			if again, _ := ParseRef(got.String()); again != got {
				t.Errorf("round trip of %q gave %+v", got.String(), again)
			}
		}
	}
}

func TestNormalizeType(t *testing.T) {
	for in, want := range map[string]string{"Decision": "decision", "gotchas": "gotcha", "HARDWARE": "hardware", "bugs": "bug", "stack": "", "": ""} {
		if got, _ := NormalizeType(in); got != want {
			t.Errorf("NormalizeType(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestGroundedFieldsRoundTrip(t *testing.T) {
	now := time.Date(2026, 11, 17, 15, 30, 0, 0, time.UTC)
	m := &Memory{ID: "x1", Content: "IMU CS is GPIO 5", Source: SourceAgent, Status: StatusStale,
		CreatedAt: now, UpdatedAt: now, Type: TypeHardware, Conf: 0.8,
		Sources:     []Ref{{Path: "main/imu.c", Symbol: "imu_init", Hash: "9b1e"}, {Path: "main/pins.h", Start: 12, End: 18}},
		StaleReason: "imu_init was removed", StaleCommit: "abc1234", StaleAt: &now}
	got, err := decode(encode(m))
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != TypeHardware || got.Conf != 0.8 || got.Status != StatusStale || got.StaleReason != m.StaleReason ||
		got.StaleCommit != "abc1234" || got.StaleAt == nil || !got.StaleAt.Equal(now) || len(got.Sources) != 2 ||
		got.Sources[0] != m.Sources[0] || got.Sources[1] != m.Sources[1] {
		t.Fatalf("round trip mismatch: %+v", got)
	}
	b, _ := json.Marshal(got.clone())
	var back Memory
	if err := json.Unmarshal(b, &back); err != nil || len(back.Sources) != 2 || back.Sources[0].Symbol != "imu_init" {
		t.Fatalf("JSON round trip: %v %+v (%s)", err, back, b)
	}
}

func TestOldFilesStillLoad(t *testing.T) {
	old := "---\nid: 20261110-153000-1a2b3c4d\nsource: agent\nstatus: active\ncreated_at: 2026-11-10T15:30:00Z\ncommit: 3f2c1ab\nproject: drone\n---\n\nMotors spin up in main/motor.c.\n"
	m, err := decode([]byte(old))
	if err != nil {
		t.Fatal(err)
	}
	if m.Type != "" || len(m.Sources) != 0 || m.Conf != 0 || m.Status != StatusActive || m.Confidence() != 0.6 {
		t.Fatalf("old file decoded wrong: %+v", m)
	}
	// A status from a newer version is read as active, not dropped.
	m, err = decode([]byte("---\nid: a\nstatus: someday\n---\n\nx\n"))
	if err != nil || m.Status != StatusActive {
		t.Fatalf("unknown status: %v %+v", err, m)
	}
}

func TestRememberWithTypeAndSources(t *testing.T) {
	s := newStore(t)
	res, err := s.Remember(RememberInput{Content: "IMU CS is GPIO 5", Source: SourceAgent, Project: "drone", Type: "Hardware",
		Sources: []Ref{{Path: "main/pins.h", Start: 3}, {Path: "/abs/no.c"}, {Path: "main/pins.h", Start: 3}}, Confidence: 1.7})
	if err != nil {
		t.Fatal(err)
	}
	m := res.Memory
	if m.Type != TypeHardware || len(m.Sources) != 1 || m.Conf != 1 || TypeOf(m) != TypeHardware {
		t.Fatalf("got %+v", m)
	}
	found := false
	for _, tag := range m.Tags {
		found = found || tag == "hardware"
	}
	if !found {
		t.Fatalf("type tag missing: %v", m.Tags)
	}
	// A repeat fills in sources the first one didn't have.
	plain, _ := s.Remember(RememberInput{Content: "Uses ESP-IDF 5.3", Project: "drone"})
	again, _ := s.Remember(RememberInput{Content: "uses esp-idf 5.3", Project: "drone", Sources: []Ref{{Path: "CMakeLists.txt"}}})
	if !again.Duplicate || again.Memory.ID != plain.Memory.ID || len(again.Memory.Sources) != 1 {
		t.Fatalf("repeat: %+v", again)
	}
}

func TestStaleLifecycle(t *testing.T) {
	s := newStore(t)
	res, _ := s.Remember(RememberInput{Content: "imu_read_accel returns g", Project: "drone", Sources: []Ref{{Path: "main/imu.c", Symbol: "imu_read_accel"}}})
	id := res.Memory.ID

	m, err := s.MarkStale(id, StaleInput{Reason: "imu_read_accel was removed", Commit: "abc1234"})
	if err != nil || m.Status != StatusStale || m.StaleAt == nil || !m.Live() {
		t.Fatalf("mark stale: %v %+v", err, m)
	}
	if list, _ := s.List(false); len(list) != 0 {
		t.Fatalf("stale facts should leave the default list: %+v", list)
	}
	if st, _ := s.Stale("drone"); len(st) != 1 {
		t.Fatalf("Stale(drone) = %d", len(st))
	}
	if st, _ := s.Stale("other"); len(st) != 0 {
		t.Fatalf("Stale(other) = %d", len(st))
	}

	m, err = s.Confirm(id, []Ref{{Path: "main/imu.c", Symbol: "imu_read_accel_g", Hash: "abcd"}})
	if err != nil || m.Status != StatusActive || m.StaleReason != "" || m.StaleAt != nil || m.Sources[0].Symbol != "imu_read_accel_g" {
		t.Fatalf("confirm: %v %+v", err, m)
	}

	_, _ = s.MarkStale(id, StaleInput{Reason: "gone"})
	m, err = s.Expire(id)
	if err != nil || m.Status != StatusExpired || m.InvalidAt == nil {
		t.Fatalf("retire: %v %+v", err, m)
	}
	if _, err := s.MarkStale(id, StaleInput{Reason: "x"}); err == nil {
		t.Fatal("an expired fact can't go stale")
	}
	if _, err := s.Confirm(id, nil); err == nil {
		t.Fatal("an expired fact can't be confirmed")
	}
}

func TestNewFactReplacesStaleOneWithSameKey(t *testing.T) {
	s := newStore(t)
	old, _ := s.Remember(RememberInput{Content: "Kp is 1.2", Key: "pid.kp", Project: "drone"})
	_, _ = s.MarkStale(old.Memory.ID, StaleInput{Reason: "value changed"})
	res, _ := s.Remember(RememberInput{Content: "Kp is 1.4", Key: "pid.kp", Project: "drone"})
	if len(res.Superseded) != 1 || res.Superseded[0].ID != old.Memory.ID {
		t.Fatalf("stale fact should be replaced: %+v", res)
	}
}

func TestTouchAndLastUsed(t *testing.T) {
	root := t.TempDir()
	s, err := Open(filepath.Join(root, "memories"))
	if err != nil {
		t.Fatal(err)
	}
	a, _ := s.Remember(RememberInput{Content: "one"})
	before, _ := os.ReadFile(filepath.Join(s.Dir(), a.Memory.ID+".md"))
	s.Touch(a.Memory.ID)
	if _, ok := s.LastUsed()[a.Memory.ID]; !ok {
		t.Fatal("touch not recorded")
	}
	after, _ := os.ReadFile(filepath.Join(s.Dir(), a.Memory.ID+".md"))
	if string(before) != string(after) {
		t.Fatal("touch must not rewrite the memory file")
	}
	if _, err := os.Stat(filepath.Join(root, "usage.json")); err != nil {
		t.Fatal(err)
	}
}
