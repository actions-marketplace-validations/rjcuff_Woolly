package peer

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rjcuff/woolly/internal/memory"
)

func openStore(t *testing.T) *memory.Store {
	t.Helper()
	s, err := memory.Open(filepath.Join(t.TempDir(), "memories"))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func folderFiles(t *testing.T, shared string) []string {
	t.Helper()
	var out []string
	filepath.WalkDir(filepath.Join(shared, FolderDir), func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			out = append(out, p)
		}
		return nil
	})
	return out
}

func TestFolderSyncBetweenTwoDevices(t *testing.T) {
	cfg, _ := NewConfig()
	shared := t.TempDir()
	a, b := openStore(t), openStore(t)
	ctx := context.Background()
	a.Remember(memory.RememberInput{Content: "Lives in Denver"})
	b.Remember(memory.RememberInput{Content: "Likes tea"})

	if res, err := SyncFolder(ctx, a, cfg, shared); err != nil || res.Sent != 1 || res.Received != 0 {
		t.Fatalf("a first: %+v %v", res, err)
	}
	if res, err := SyncFolder(ctx, b, cfg, shared); err != nil || res.Received != 1 || res.Sent != 1 {
		t.Fatalf("b: %+v %v", res, err)
	}
	if res, err := SyncFolder(ctx, a, cfg, shared); err != nil || res.Received != 1 {
		t.Fatalf("a second: %+v %v", res, err)
	}
	if len(contents(t, a)) != 2 || len(contents(t, b)) != 2 {
		t.Fatalf("a %v, b %v", contents(t, a), contents(t, b))
	}
	if res, _ := SyncFolder(ctx, a, cfg, shared); res.Received+res.Sent+res.Deleted != 0 {
		t.Fatalf("nothing changed, so nothing should move: %+v", res)
	}
}

func TestFolderSyncNewerWinsAndDeletesCarry(t *testing.T) {
	cfg, _ := NewConfig()
	shared := t.TempDir()
	a, b := openStore(t), openStore(t)
	ctx := context.Background()
	r, _ := a.Remember(memory.RememberInput{Content: "Uses Postgres"})
	gone, _ := a.Remember(memory.RememberInput{Content: "Old note to delete"})
	SyncFolder(ctx, a, cfg, shared)
	SyncFolder(ctx, b, cfg, shared)

	time.Sleep(10 * time.Millisecond)
	text := "Uses Postgres 16"
	if _, err := b.Update(r.Memory.ID, memory.UpdateInput{Content: &text}); err != nil {
		t.Fatal(err)
	}
	if err := b.Delete(gone.Memory.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := SyncFolder(ctx, b, cfg, shared); err != nil {
		t.Fatal(err)
	}
	res, err := SyncFolder(ctx, a, cfg, shared)
	if err != nil || res.Deleted != 1 {
		t.Fatalf("a: %+v %v", res, err)
	}
	got := contents(t, a)
	if len(got) != 1 || got[0] != text {
		t.Fatalf("a = %v, want only %q", got, text)
	}
	// The deleted memory's file left the folder too.
	for _, p := range folderFiles(t, shared) {
		if strings.Contains(filepath.ToSlash(p), "/memories/") {
			data, _ := os.ReadFile(p)
			if bytes.Contains(data, []byte("Old note")) {
				t.Fatal("deleted memory still in the folder")
			}
		}
	}
	if n := len(folderFiles(t, filepath.Join(shared))); n != 2 {
		t.Fatalf("want 1 memory file and 1 deleted marker, got %d", n)
	}
}

func TestFolderHoldsOnlyCiphertext(t *testing.T) {
	cfg, _ := NewConfig()
	shared := t.TempDir()
	a := openStore(t)
	r, _ := a.Remember(memory.RememberInput{Content: "Secret project codename Bluebird"})
	if _, err := SyncFolder(context.Background(), a, cfg, shared); err != nil {
		t.Fatal(err)
	}
	files := folderFiles(t, shared)
	if len(files) != 1 {
		t.Fatalf("files = %v", files)
	}
	for _, p := range files {
		data, _ := os.ReadFile(p)
		for _, leak := range []string{"Bluebird", r.Memory.ID, "content", "updated_at"} {
			if bytes.Contains(data, []byte(leak)) || strings.Contains(p, leak) {
				t.Fatalf("%s leaks %q", p, leak)
			}
		}
	}
	// Another key can't read it, and sees nothing to sync.
	other, _ := NewConfig()
	b := openStore(t)
	res, err := SyncFolder(context.Background(), b, other, shared)
	if !errors.Is(err, ErrOtherKey) || res.Received != 0 || len(contents(t, b)) != 0 {
		t.Fatalf("wrong key: %+v %v %v", res, err, contents(t, b))
	}
}

func TestFolderRefusesTamperedAndSwappedFiles(t *testing.T) {
	cfg, _ := NewConfig()
	shared := t.TempDir()
	a := openStore(t)
	a.Remember(memory.RememberInput{Content: "First"})
	a.Remember(memory.RememberInput{Content: "Second"})
	SyncFolder(context.Background(), a, cfg, shared)
	files := folderFiles(t, shared)
	if len(files) != 2 {
		t.Fatalf("files = %v", files)
	}
	// Flip a byte in one, and copy the other over a third name.
	data, _ := os.ReadFile(files[0])
	data[len(data)-1] ^= 0xff
	os.WriteFile(files[0], data, 0o600)
	other, _ := os.ReadFile(files[1])
	os.WriteFile(filepath.Join(filepath.Dir(files[1]), "0000000000000000000000000000000"+fileExt), other, 0o600)

	b := openStore(t)
	res, err := SyncFolder(context.Background(), b, cfg, shared)
	if err != nil {
		t.Fatal(err)
	}
	if res.Received != 1 || len(contents(t, b)) != 1 {
		t.Fatalf("only the untouched file should arrive: %+v %v", res, contents(t, b))
	}
}

func TestFolderMustExist(t *testing.T) {
	cfg, _ := NewConfig()
	_, err := SyncFolder(context.Background(), openStore(t), cfg, filepath.Join(t.TempDir(), "missing"))
	if err == nil || !strings.Contains(err.Error(), "can't open the sync folder") {
		t.Fatalf("err = %v", err)
	}
}

func TestConfigFolders(t *testing.T) {
	c := &Config{}
	if !c.AddFolder("/a/b") || c.AddFolder("/a/b/") || len(c.Folders) != 1 {
		t.Fatalf("folders = %+v", c.Folders)
	}
	if !c.RemoveFolder("/a/b") || c.RemoveFolder("/a/b") || len(c.Folders) != 0 {
		t.Fatalf("folders after remove = %+v", c.Folders)
	}
}

func TestFolderIgnoresLinksHugeFilesAndConflictCopies(t *testing.T) {
	cfg, _ := NewConfig()
	shared := t.TempDir()
	a := openStore(t)
	a.Remember(memory.RememberInput{Content: "Real memory"})
	if _, err := SyncFolder(context.Background(), a, cfg, shared); err != nil {
		t.Fatal(err)
	}
	mems := filepath.Join(shared, FolderDir, memoriesDir)
	files := folderFiles(t, shared)
	real, _ := os.ReadFile(files[0])

	// A sync tool's conflict copy of the real file: a name Woolly never writes.
	conflict := strings.TrimSuffix(filepath.Base(files[0]), fileExt) + ".sync-conflict-20261122-1.woolly"
	os.WriteFile(filepath.Join(mems, conflict), real, 0o600)
	// A file bigger than any sync message, with a valid name.
	big := filepath.Join(mems, strings.Repeat("b", 32)+fileExt)
	fh, _ := os.Create(big)
	fh.Truncate(maxBody + 1)
	fh.Close()
	// A link to somewhere else, if this system allows making one.
	link := filepath.Join(mems, strings.Repeat("c", 32)+fileExt)
	linked := os.Symlink(files[0], link) == nil

	b := openStore(t)
	res, err := SyncFolder(context.Background(), b, cfg, shared)
	if err != nil {
		t.Fatal(err)
	}
	if res.Received != 1 || len(contents(t, b)) != 1 {
		t.Fatalf("only the real file should arrive: %+v %v", res, contents(t, b))
	}
	if res.Skipped != 1 {
		t.Fatalf("the oversized file should count as skipped, got %d", res.Skipped)
	}
	if linked {
		t.Log("symlink planted and ignored")
	}
}

func TestFolderRefusesLinkedSubfolders(t *testing.T) {
	cfg, _ := NewConfig()
	shared, elsewhere := t.TempDir(), t.TempDir()
	if err := os.Symlink(elsewhere, filepath.Join(shared, FolderDir)); err != nil {
		t.Skip("can't make symlinks here:", err)
	}
	_, err := SyncFolder(context.Background(), openStore(t), cfg, shared)
	if err == nil || !strings.Contains(err.Error(), "isn't a plain folder") {
		t.Fatalf("err = %v", err)
	}
	if entries, _ := os.ReadDir(elsewhere); len(entries) != 0 {
		t.Fatalf("wrote through the link: %v", entries)
	}
}

func TestFolderSkippedCountsChangedFiles(t *testing.T) {
	cfg, _ := NewConfig()
	shared := t.TempDir()
	a := openStore(t)
	a.Remember(memory.RememberInput{Content: "One"})
	a.Remember(memory.RememberInput{Content: "Two"})
	SyncFolder(context.Background(), a, cfg, shared)
	files := folderFiles(t, shared)
	data, _ := os.ReadFile(files[0])
	data[len(data)-1] ^= 1
	os.WriteFile(files[0], data, 0o600)
	res, err := SyncFolder(context.Background(), openStore(t), cfg, shared)
	if err != nil || res.Skipped != 1 || res.Received != 1 {
		t.Fatalf("res = %+v err = %v", res, err)
	}
}
