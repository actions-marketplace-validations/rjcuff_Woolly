package peer

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/rjcuff/woolly/internal/memory"
)

// Folder is a folder that devices sync through, like a Syncthing folder or a
// git repo you push and pull yourself. Woolly only reads and writes files in
// it; moving them between computers is up to the tool that owns the folder.
type Folder struct {
	Path     string    `json:"path"`
	LastSync time.Time `json:"last_sync,omitempty"`
	LastErr  string    `json:"last_error,omitempty"`
}

// FolderDir is the folder Woolly keeps inside a sync folder.
const FolderDir = "woolly-sync"

// Everything Woolly writes to a sync folder is sealed with the join key, so
// whoever holds the folder, like a git host, sees only ciphertext. File names
// are keyed hashes of memory IDs, so they don't reveal them either.
const (
	memoriesDir = "memories"
	deadDir     = "deleted"
	fileExt     = ".woolly"
)

// AddFolder remembers a sync folder, once. path should be absolute.
func (c *Config) AddFolder(path string) bool {
	for _, f := range c.Folders {
		if samePath(f.Path, path) {
			return false
		}
	}
	c.Folders = append(c.Folders, Folder{Path: path})
	return true
}

// RemoveFolder forgets a sync folder. The files in it are left alone.
func (c *Config) RemoveFolder(path string) bool {
	for i, f := range c.Folders {
		if samePath(f.Path, path) {
			c.Folders = append(c.Folders[:i:i], c.Folders[i+1:]...)
			return true
		}
	}
	return false
}

func samePath(a, b string) bool {
	return filepath.Clean(a) == filepath.Clean(b) || strings.EqualFold(filepath.Clean(a), filepath.Clean(b)) && os.PathSeparator == '\\'
}

// SyncFolder brings the store and the sync folder at path up to date with
// each other.
func SyncFolder(ctx context.Context, store *memory.Store, cfg *Config, path string) (Result, error) {
	key, err := cfg.key()
	if err != nil {
		return Result{Peer: path}, err
	}
	st, err := os.Stat(path)
	if err != nil || !st.IsDir() {
		return Result{Peer: path}, fmt.Errorf("can't open the sync folder %s. Is it still there?", path)
	}
	root := filepath.Join(path, FolderDir)
	for _, d := range []string{root, filepath.Join(root, memoriesDir), filepath.Join(root, deadDir)} {
		if err := realDir(d); err != nil {
			return Result{Peer: path}, err
		}
	}
	f := &folder{root: root, nameKey: subkey(key, "woolly folder name key v1"), sealKey: subkey(key, "woolly folder seal key v1")}
	res, err := syncWith(ctx, store, f, path)
	res.Skipped = f.skipped
	return res, err
}

// realDir makes dir if it's missing, and refuses one that is a link or not a
// folder, so files can't be written somewhere else through the sync folder.
func realDir(dir string) error {
	st, err := os.Lstat(dir)
	if errors.Is(err, os.ErrNotExist) {
		return os.Mkdir(dir, 0o700)
	}
	if err != nil {
		return err
	}
	if !st.IsDir() {
		return fmt.Errorf("%s in the sync folder isn't a plain folder, so Woolly won't use it", dir)
	}
	return nil
}

// subkey derives a key for one use from the join code's key, so names and
// encryption never share a key.
func subkey(key []byte, purpose string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(purpose))
	return h.Sum(nil)
}

// folder is a sync folder as a sync target.
type folder struct {
	root    string
	nameKey []byte
	sealKey []byte
	// skipped counts files that didn't open with the key: changed, from
	// another join code, or a sync tool's conflict copy.
	skipped int
}

// name is the file name for a memory ID: a keyed hash, the same on every
// device with the key.
func (f *folder) name(id string) string {
	h := hmac.New(sha256.New, f.nameKey)
	h.Write([]byte(id))
	return hex.EncodeToString(h.Sum(nil)[:16]) + fileExt
}

// validName matches the names Woolly writes. Anything else, like a sync
// tool's "(conflicted copy)" or ".sync-conflict-" file, is left alone.
var validName = regexp.MustCompile(`^[0-9a-f]{32}\.woolly$`)

// The label ties each sealed file to its name, so one can't be swapped for
// another.
func fileLabel(name string) string { return "folder-file:" + name }
func deadLabel(name string) string { return "folder-deleted:" + name }

// sealedFile is what a memory file in the folder holds once opened.
type sealedFile struct {
	ID    string `json:"id"`
	Data  []byte `json:"data"`
	Trash bool   `json:"trash,omitempty"`
}

type sealedDead struct {
	ID string `json:"id"`
}

// read hands each sealed file in one of the folder's directories to each,
// which opens it with the key. Only regular files with names Woolly writes
// are read, and only up to maxBody, so a planted link, pipe or huge file
// can't stall or exhaust a sync.
func (f *folder) read(ctx context.Context, dir string, each func(name string, sealed []byte)) error {
	entries, err := os.ReadDir(filepath.Join(f.root, dir))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		if !e.Type().IsRegular() || !validName.MatchString(e.Name()) {
			continue
		}
		data, ok := readCapped(filepath.Join(f.root, dir, e.Name()))
		if !ok {
			f.skipped++
			continue
		}
		each(e.Name(), data)
	}
	return nil
}

// readCapped reads a regular file of at most maxBody bytes.
func readCapped(path string) ([]byte, bool) {
	st, err := os.Lstat(path)
	if err != nil || !st.Mode().IsRegular() || st.Size() > maxBody {
		return nil, false
	}
	fh, err := os.Open(path)
	if err != nil {
		return nil, false
	}
	defer fh.Close()
	data, err := io.ReadAll(io.LimitReader(fh, maxBody+1))
	if err != nil || len(data) > maxBody {
		return nil, false
	}
	return data, true
}

// ErrOtherKey means a sync folder holds files that none open with this
// device's join code.
var ErrOtherKey = errors.New("this folder holds memories from a different join code. Run `woolly sync folder` on it again with `--code` and the code from your first device")

func (f *folder) manifest(ctx context.Context) ([]memory.Item, []string, string, error) {
	var items []memory.Item
	seen, opened := 0, 0
	err := f.read(ctx, memoriesDir, func(name string, sealed []byte) {
		seen++
		var sf sealedFile
		if open(f.sealKey, fileLabel(name), sealed, &sf) != nil || f.name(sf.ID) != name {
			f.skipped++
			return
		}
		opened++
		if it, err := memory.ItemOf(sf.Data, sf.Trash); err == nil && it.ID == sf.ID {
			items = append(items, it)
		}
	})
	if err != nil {
		return nil, nil, "", err
	}
	if seen > 0 && opened == 0 {
		return nil, nil, "", ErrOtherKey
	}
	var dead []string
	err = f.read(ctx, deadDir, func(name string, sealed []byte) {
		var sd sealedDead
		if open(f.sealKey, deadLabel(name), sealed, &sd) == nil && memory.ValidID(sd.ID) && f.name(sd.ID) == name {
			dead = append(dead, sd.ID)
		} else {
			f.skipped++
		}
	})
	if err != nil {
		return nil, nil, "", err
	}
	return items, dead, "", ctx.Err()
}

func (f *folder) get(ctx context.Context, ids []string) ([]File, error) {
	var files []File
	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return files, err
		}
		name := f.name(id)
		data, ok := readCapped(filepath.Join(f.root, memoriesDir, name))
		if !ok {
			continue
		}
		var sf sealedFile
		if open(f.sealKey, fileLabel(name), data, &sf) != nil || sf.ID != id {
			continue
		}
		files = append(files, File{Data: sf.Data, Trash: sf.Trash})
	}
	return files, nil
}

func (f *folder) put(ctx context.Context, files []File, tombstones []string) (int, error) {
	for _, id := range tombstones {
		if !memory.ValidID(id) {
			continue
		}
		name := f.name(id)
		if err := f.write(deadDir, name, deadLabel(name), sealedDead{ID: id}); err != nil {
			return 0, err
		}
		if err := os.Remove(filepath.Join(f.root, memoriesDir, name)); err != nil && !errors.Is(err, os.ErrNotExist) {
			return 0, err
		}
	}
	n := 0
	for _, file := range files {
		if err := ctx.Err(); err != nil {
			return n, err
		}
		it, err := memory.ItemOf(file.Data, file.Trash)
		if err != nil {
			continue
		}
		name := f.name(it.ID)
		if err := f.write(memoriesDir, name, fileLabel(name), sealedFile{ID: it.ID, Data: file.Data, Trash: file.Trash}); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

// write seals v and saves it atomically, so a sync tool never copies half a
// file.
func (f *folder) write(dir, name, label string, v any) error {
	sealed, err := seal(f.sealKey, label, v)
	if err != nil {
		return err
	}
	path := filepath.Join(f.root, dir, name)
	tmp, err := os.CreateTemp(filepath.Join(f.root, dir), ".tmp-*")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(sealed); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return nil
}
