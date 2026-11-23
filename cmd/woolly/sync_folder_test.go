package main

import (
	"strings"
	"testing"

	"github.com/rjcuff/woolly/internal/peer"
)

func loadSync(t *testing.T, home string) *peer.Config {
	t.Helper()
	cfg, err := peer.Load(home)
	if err != nil || cfg == nil {
		t.Fatalf("sync config: %v %v", cfg, err)
	}
	return cfg
}

// The device that ran `woolly sync init` never learns who joined it, so
// adding a folder must not stop it listening for them.
func TestAddingAFolderKeepsTheNetworkListener(t *testing.T) {
	home := t.TempDir()
	t.Setenv("WOOLLY_HOME", home)
	t.Setenv("WOOLLY_NO_EMBED", "1")
	captureStdout(t, func() {
		if err := syncInit(); err != nil {
			t.Error(err)
		}
	})
	captureStdout(t, func() {
		if err := syncFolderCmd([]string{t.TempDir()}); err != nil {
			t.Error(err)
		}
	})
	if cfg := loadSync(t, home); !listens(cfg) || len(cfg.Folders) != 1 {
		t.Fatalf("after init then folder: listens=%v folders=%d", listens(cfg), len(cfg.Folders))
	}
}

func TestFolderOnlyDeviceDoesNotListenUntilInit(t *testing.T) {
	home := t.TempDir()
	t.Setenv("WOOLLY_HOME", home)
	t.Setenv("WOOLLY_NO_EMBED", "1")
	captureStdout(t, func() {
		if err := syncFolderCmd([]string{t.TempDir()}); err != nil {
			t.Error(err)
		}
	})
	if listens(loadSync(t, home)) {
		t.Fatal("a folder-only device should not listen")
	}
	captureStdout(t, func() { _ = syncInit() })
	if !listens(loadSync(t, home)) {
		t.Fatal("after init it should listen")
	}
}

func TestSyncStatusHidesTheCodeUnlessAsked(t *testing.T) {
	home := t.TempDir()
	t.Setenv("WOOLLY_HOME", home)
	t.Setenv("WOOLLY_NO_EMBED", "1")
	captureStdout(t, func() { _ = syncFolderCmd([]string{t.TempDir()}) })
	code := loadSync(t, home).JoinCode()
	if out := captureStdout(t, func() { _ = syncStatus(false) }); strings.Contains(out, code) {
		t.Fatalf("status printed the join code:\n%s", out)
	}
	if out := captureStdout(t, func() { _ = syncStatus(true) }); !strings.Contains(out, code) {
		t.Fatalf("--show-code should print it:\n%s", out)
	}
}

func TestJoinRefusesAnotherCodeOnAFolderDevice(t *testing.T) {
	home := t.TempDir()
	t.Setenv("WOOLLY_HOME", home)
	t.Setenv("WOOLLY_NO_EMBED", "1")
	captureStdout(t, func() { _ = syncFolderCmd([]string{t.TempDir()}) })
	other, _ := peer.NewConfig()
	err := syncJoin("127.0.0.1:1", other.JoinCode())
	if err == nil || !strings.Contains(err.Error(), "different group") {
		t.Fatalf("err = %v", err)
	}
}
