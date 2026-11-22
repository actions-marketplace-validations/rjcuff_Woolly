package memory

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rjcuff/woolly/internal/secrets"
)

// ErrNotFound is returned when no memory matches an ID.
var ErrNotFound = errors.New("memory not found")

// SensitiveError is returned when a memory looks like it contains a secret.
type SensitiveError struct {
	What string // e.g. "an OpenAI API key"
}

func (e *SensitiveError) Error() string {
	return "not saved: this looks like it contains " + e.What + ". Woolly doesn't store secrets"
}

// InjectionError is returned when an AI or an import tries to save text that
// reads like instructions to an AI rather than a fact about the user.
type InjectionError struct{}

func (e *InjectionError) Error() string {
	return "not saved: this reads like instructions to an AI, not a fact about the user"
}

func checkSensitive(text string, allow bool) error {
	if allow {
		return nil
	}
	if found := secrets.Scan(text); len(found) > 0 {
		return &SensitiveError{What: secrets.Describe(found)}
	}
	return nil
}

// Store keeps memories as Markdown files in one directory. It caches them in
// memory and reloads automatically when files change on disk, so a running
// server picks up edits made by the CLI, an editor, or a sync tool.
type Store struct {
	dir  string
	mu   sync.Mutex
	mems map[string]*Memory
	fp   string // fingerprint of the directory at last load
	// problems lists files that couldn't be read at last load, so a typo in
	// a hand edit shows up instead of the memory quietly going missing.
	problems []string

	usageMu sync.Mutex // guards usage.json
}

// Open opens (creating if needed) a store rooted at dir.
func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create memory dir: %w", err)
	}
	s := &Store{dir: dir}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

// Dir returns the directory holding the memory files.
func (s *Store) Dir() string { return s.dir }

func isMemoryFile(name string) bool {
	return strings.HasSuffix(name, ".md") && !strings.HasPrefix(name, ".")
}

func (s *Store) fingerprint() (string, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return "", err
	}
	// Every file's name, size and time, so a rename or a sync tool that
	// keeps an older time still counts as a change. ReadDir sorts by name.
	h := sha256.New()
	for _, e := range entries {
		if e.IsDir() || !isMemoryFile(e.Name()) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		fmt.Fprintf(h, "%s:%d:%d;", e.Name(), info.Size(), info.ModTime().UnixNano())
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Problems lists memory files Woolly couldn't read fully, with why. Fix or
// remove them by hand; nothing else touches them.
func (s *Store) Problems() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.refresh()
	return append([]string(nil), s.problems...)
}

func (s *Store) load() error {
	fp, err := s.fingerprint()
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return err
	}
	mems := make(map[string]*Memory, len(entries))
	var problems []string
	for _, e := range entries {
		if e.IsDir() || !isMemoryFile(e.Name()) {
			continue
		}
		path := filepath.Join(s.dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			problems = append(problems, e.Name()+": "+err.Error())
			continue
		}
		m, err := decode(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "woolly: skipping %s: %v\n", e.Name(), err)
			problems = append(problems, e.Name()+": "+err.Error())
			continue
		}
		if m.badStatus != "" {
			problems = append(problems, fmt.Sprintf("%s: unknown status %q, read as active", e.Name(), m.badStatus))
		}
		m.file = path
		mems[m.ID] = m
	}
	s.mems, s.fp, s.problems = mems, fp, problems
	return nil
}

// refresh reloads from disk if anything changed. Caller holds s.mu.
func (s *Store) refresh() error {
	fp, err := s.fingerprint()
	if err != nil {
		return err
	}
	if fp == s.fp && s.mems != nil {
		return nil
	}
	return s.load()
}

// write saves a memory atomically (temp file + rename). Caller holds s.mu.
func (s *Store) write(m *Memory) error {
	if m.file == "" {
		m.file = filepath.Join(s.dir, m.ID+".md")
	}
	tmp, err := os.CreateTemp(s.dir, ".tmp-*")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(encode(m)); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	if err := os.Rename(tmp.Name(), m.file); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	s.mems[m.ID] = m
	return nil
}

// find resolves a full ID or an unambiguous ID prefix. Caller holds s.mu.
func (s *Store) find(id string) (*Memory, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrNotFound
	}
	if m, ok := s.mems[id]; ok {
		return m, nil
	}
	var match *Memory
	for mid, m := range s.mems {
		if strings.HasPrefix(mid, id) {
			if match != nil {
				return nil, fmt.Errorf("id prefix %q matches more than one memory", id)
			}
			match = m
		}
	}
	if match == nil {
		return nil, ErrNotFound
	}
	return match, nil
}

// RememberInput describes a new memory.
type RememberInput struct {
	Content string
	Key     string   // optional stable slot, e.g. "user.location"
	Tags    []string // optional
	Source  Source   // defaults to SourceUser
	// AllowSensitive saves the memory even if it looks like it holds a secret.
	AllowSensitive bool
	// Via names the app saving the memory.
	Via string
	// Project scopes the memory to one project. Keys only clash within the
	// same project.
	Project string
	// Commit is the git commit a project note was saved at.
	Commit string
	// ID, if set and free, is used instead of a new one, so a note shared
	// through a repo keeps the same ID on every computer.
	ID string
	// User, Agent and Session scope the memory for apps built on Woolly.
	User, Agent, Session string
	// Metadata is free-form data an app attaches to the memory.
	Metadata map[string]any
	// Type is decision, convention, gotcha, bug or hardware. It also adds
	// the matching tag, which is how woolly.md files the fact.
	Type string
	// Sources are the files, symbols or line ranges the fact is about.
	Sources []Ref
	// Confidence is from 0 to 1. Zero means the default for the source.
	Confidence float64
	// Untrusted marks text that came through an AI or an app, whatever
	// source it claims. It is always checked for instructions to an AI.
	Untrusted bool
}

// UserStatedTag marks a fact an AI app or a script said came from the user.
const UserStatedTag = "user-stated"

// claimed is the source Woolly saves for what a caller claims. Woolly can't
// check that untrusted text really came from the user, so a claim of "user"
// is saved one step down, as import, tagged so it's clear who said what. It
// still beats an AI's guess but never replaces what the user saved.
func claimed(src Source, untrusted bool, tags []string) (Source, []string) {
	if untrusted && src == SourceUser {
		return SourceImport, append(append([]string(nil), tags...), UserStatedTag)
	}
	return src, tags
}

// RememberResult reports what Remember did.
type RememberResult struct {
	Memory        Memory   `json:"memory"`
	Duplicate     bool     `json:"duplicate,omitempty"`
	Superseded    []Memory `json:"superseded,omitempty"`
	ConflictsWith *Memory  `json:"conflicts_with,omitempty"`
}

// Remember saves a memory. A repeat refreshes the existing one. A newer fact
// with the same key replaces older ones, unless it comes from a less trusted
// source, in which case it's saved as a conflict for the user to review.
func (s *Store) Remember(in RememberInput) (*RememberResult, error) {
	content := strings.TrimSpace(in.Content)
	if content == "" {
		return nil, errors.New("memory content is empty")
	}
	src := in.Source
	if src == "" {
		src = SourceUser
	}
	if src.Rank() == 0 {
		return nil, fmt.Errorf("unknown source %q (use user, import or agent)", src)
	}
	src, in.Tags = claimed(src, in.Untrusted, in.Tags)
	if err := checkSensitive(content, in.AllowSensitive); err != nil {
		return nil, err
	}
	// Anything you type yourself is trusted. Text from an AI or an import
	// could have come from a malicious web page or document.
	if (src != SourceUser || in.Untrusted) && secrets.LooksLikeInstructions(content) {
		return nil, &InjectionError{}
	}
	key := NormalizeKey(in.Key)
	typ, _ := NormalizeType(in.Type)
	if typ != "" {
		in.Tags = append(in.Tags, typ)
	}
	tags := NormalizeTags(in.Tags)
	project := NormalizeProject(in.Project)
	refs := cleanRefs(in.Sources)
	conf := ClampConfidence(in.Confidence)
	slot := &Memory{Project: project, User: CleanScopeID(in.User), Agent: CleanScopeID(in.Agent), Session: CleanScopeID(in.Session)}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(); err != nil {
		return nil, err
	}
	now := time.Now().UTC()

	// Duplicate: same key and same text as a live memory.
	for _, ex := range s.mems {
		if ex.Status == StatusSuperseded || ex.Status == StatusExpired || ex.Key != key || !sameSlot(ex, slot) || !sameText(ex.Content, content) {
			continue
		}
		ex.UpdatedAt = now
		ex.Tags = mergeTags(ex.Tags, tags)
		if src.Rank() > ex.Source.Rank() {
			ex.Source = src // the user confirmed something an agent guessed
		}
		if ex.Type == "" {
			ex.Type = typ
		}
		if len(ex.Sources) == 0 {
			ex.Sources = refs
		}
		if conf > ex.Conf {
			ex.Conf = conf
		}
		if err := s.write(ex); err != nil {
			return nil, err
		}
		if ex.Status == StatusConflict {
			if other, ok := s.mems[ex.ConflictsWith]; !ok || ex.Source.Rank() >= other.Source.Rank() {
				if _, err := s.resolveLocked(ex, now); err != nil {
					return nil, err
				}
			}
		}
		return &RememberResult{Memory: ex.clone(), Duplicate: true}, nil
	}

	id := newID(now)
	if in.ID != "" && ValidID(in.ID) && s.mems[in.ID] == nil {
		id = in.ID
	}
	m := &Memory{
		ID: id, Content: content, Key: key, Tags: tags,
		Source: src, Status: StatusActive, CreatedAt: now, UpdatedAt: now,
		Via: CleanApp(in.Via), Project: project, Commit: in.Commit,
		User: slot.User, Agent: slot.Agent, Session: slot.Session, Metadata: in.Metadata,
		Type: typ, Sources: refs, Conf: conf,
	}
	res := &RememberResult{}

	if key != "" {
		var same []*Memory
		var stronger *Memory
		for _, ex := range s.mems {
			if ex.Key != key || !sameSlot(ex, m) || ex.Status == StatusSuperseded || ex.Status == StatusExpired {
				continue
			}
			same = append(same, ex)
			// A stale fact the user stated still outranks a guess: it waits for
			// their review, not for an agent to replace it.
			if (ex.Status == StatusActive || ex.Status == StatusStale) && ex.Source.Rank() > src.Rank() &&
				(stronger == nil || ex.UpdatedAt.After(stronger.UpdatedAt)) {
				stronger = ex
			}
		}
		if stronger != nil {
			m.Status = StatusConflict
			m.ConflictsWith = stronger.ID
			c := stronger.clone()
			res.ConflictsWith = &c
		} else {
			for _, ex := range same {
				ex.Status = StatusSuperseded
				ex.SupersededBy = m.ID
				ex.ConflictsWith = ""
				ex.UpdatedAt = now
				ex.InvalidAt = &now
				if err := s.write(ex); err != nil {
					return nil, err
				}
				res.Superseded = append(res.Superseded, ex.clone())
			}
		}
	}

	if err := s.write(m); err != nil {
		return nil, err
	}
	res.Memory = m.clone()
	return res, nil
}

// Get returns a memory by ID or unambiguous ID prefix.
func (s *Store) Get(id string) (Memory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(); err != nil {
		return Memory{}, err
	}
	m, err := s.find(id)
	if err != nil {
		return Memory{}, err
	}
	return m.clone(), nil
}

// List returns memories, newest first. Without includeAll it returns only
// active memories.
func (s *Store) List(includeAll bool) ([]Memory, error) {
	return s.filter(func(m *Memory) bool { return includeAll || m.Status == StatusActive })
}

// Conflicts returns memories waiting for the user to resolve them.
func (s *Store) Conflicts() ([]Memory, error) {
	return s.filter(func(m *Memory) bool { return m.Status == StatusConflict })
}

func (s *Store) filter(keep func(*Memory) bool) ([]Memory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(); err != nil {
		return nil, err
	}
	var out []Memory
	for _, m := range s.mems {
		if keep(m) {
			out = append(out, m.clone())
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].CreatedAt.After(out[j].CreatedAt)
		}
		return out[i].ID > out[j].ID
	})
	return out, nil
}

// UpdateInput changes an existing memory. Nil fields are left alone.
type UpdateInput struct {
	Content        *string
	Key            *string
	Tags           *[]string
	Project        *string
	Type           *string
	Sources        *[]Ref
	Confidence     *float64
	AllowSensitive bool
	// Untrusted text is checked for instructions to an AI, as in Remember.
	Untrusted bool
}

// Update edits a memory in place. If the memory is current and has a key,
// it becomes the current fact for that key.
func (s *Store) Update(id string, in UpdateInput) (Memory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(); err != nil {
		return Memory{}, err
	}
	m, err := s.find(id)
	if err != nil {
		return Memory{}, err
	}
	if in.Content != nil {
		c := strings.TrimSpace(*in.Content)
		if c == "" {
			return Memory{}, errors.New("memory content is empty")
		}
		if err := checkSensitive(c, in.AllowSensitive); err != nil {
			return Memory{}, err
		}
		if in.Untrusted && secrets.LooksLikeInstructions(c) {
			return Memory{}, &InjectionError{}
		}
		m.Content = c
	}
	if in.Key != nil {
		m.Key = NormalizeKey(*in.Key)
	}
	if in.Tags != nil {
		m.Tags = NormalizeTags(*in.Tags)
	}
	if in.Project != nil {
		m.Project = NormalizeProject(*in.Project)
	}
	if in.Type != nil {
		m.Type, _ = NormalizeType(*in.Type)
	}
	if in.Sources != nil {
		m.Sources = cleanRefs(*in.Sources)
	}
	if in.Confidence != nil {
		m.Conf = ClampConfidence(*in.Confidence)
	}
	m.UpdatedAt = time.Now().UTC()
	if m.Status == StatusActive && m.Key != "" {
		if _, err := s.resolveLocked(m, m.UpdatedAt); err != nil {
			return Memory{}, err
		}
	} else if err := s.write(m); err != nil {
		return Memory{}, err
	}
	return m.clone(), nil
}

// trashDir holds forgotten memories until the trash is emptied.
func (s *Store) trashDir() string { return filepath.Join(filepath.Dir(s.dir), "trash") }

// Forget moves a memory to the trash, where Restore can bring it back.
func (s *Store) Forget(id string) (Memory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(); err != nil {
		return Memory{}, err
	}
	m, err := s.find(id)
	if err != nil {
		return Memory{}, err
	}
	if err := os.MkdirAll(s.trashDir(), 0o700); err != nil {
		return Memory{}, err
	}
	// The new time lets other devices see that forgetting is the newest change.
	m.UpdatedAt = time.Now().UTC()
	if err := writeFile(filepath.Join(s.trashDir(), m.ID+".md"), encode(m)); err != nil {
		return Memory{}, err
	}
	if err := os.Remove(m.file); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Memory{}, err
	}
	delete(s.mems, m.ID)
	// A conflict with a memory that's gone has nothing left to settle, and
	// a link to it leads nowhere.
	for _, ex := range s.mems {
		linked := hasID(ex.Links, m.ID)
		if ex.ConflictsWith != m.ID && !linked {
			continue
		}
		if ex.ConflictsWith == m.ID {
			ex.ConflictsWith = ""
			if ex.Status == StatusConflict {
				ex.Status = StatusActive
			}
		}
		if linked {
			ex.Links = removeID(ex.Links, m.ID)
		}
		ex.UpdatedAt = m.UpdatedAt
		if err := s.write(ex); err != nil {
			return Memory{}, err
		}
	}
	return m.clone(), nil
}

// Trash lists forgotten memories, newest first.
func (s *Store) Trash() ([]Memory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entries, err := os.ReadDir(s.trashDir())
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []Memory
	for _, e := range entries {
		if e.IsDir() || !isMemoryFile(e.Name()) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.trashDir(), e.Name()))
		if err != nil {
			continue
		}
		if m, err := decode(data); err == nil {
			out = append(out, m.clone())
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}

// Restore brings a forgotten memory back from the trash.
func (s *Store) Restore(id string) (Memory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(); err != nil {
		return Memory{}, err
	}
	trash := s.trashDir()
	entries, err := os.ReadDir(trash)
	if err != nil {
		return Memory{}, ErrNotFound
	}
	var match string
	for _, e := range entries {
		name := strings.TrimSuffix(e.Name(), ".md")
		if name == id || (id != "" && strings.HasPrefix(name, id)) {
			if match != "" && name != id {
				return Memory{}, fmt.Errorf("id prefix %q matches more than one memory", id)
			}
			match = e.Name()
		}
	}
	if match == "" {
		return Memory{}, ErrNotFound
	}
	src := filepath.Join(trash, match)
	data, err := os.ReadFile(src)
	if err != nil {
		return Memory{}, err
	}
	m, err := decode(data)
	if err != nil {
		return Memory{}, err
	}
	if _, err := os.Stat(filepath.Join(s.dir, match)); err == nil || s.mems[m.ID] != nil {
		return Memory{}, fmt.Errorf("memory %s already exists", m.ID)
	}
	m.UpdatedAt = time.Now().UTC()
	// While it was in the trash, a newer fact may have taken its key. Then it
	// comes back as a conflict for the user to settle, never as a second
	// current fact.
	if m.Status == StatusActive && m.Key != "" {
		for _, ex := range s.mems {
			if ex.Status == StatusActive && ex.Key == m.Key && sameSlot(ex, m) {
				m.Status = StatusConflict
				m.ConflictsWith = ex.ID
				break
			}
		}
	}
	if err := writeFile(filepath.Join(s.dir, match), encode(m)); err != nil {
		return Memory{}, err
	}
	if err := os.Remove(src); err != nil {
		return Memory{}, err
	}
	if err := s.load(); err != nil {
		return Memory{}, err
	}
	got, ok := s.mems[strings.TrimSuffix(match, ".md")]
	if !ok {
		return Memory{}, ErrNotFound
	}
	return got.clone(), nil
}

// EmptyTrash permanently deletes everything in the trash.
func (s *Store) EmptyTrash() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entries, err := os.ReadDir(s.trashDir())
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	n := 0
	for _, e := range entries {
		if e.IsDir() || !isMemoryFile(e.Name()) {
			continue
		}
		if err := os.Remove(filepath.Join(s.trashDir(), e.Name())); err == nil {
			n++
			s.addTombstone(strings.TrimSuffix(e.Name(), ".md"))
		}
	}
	return n, nil
}

// ErrOutranked is returned when an AI tries to make a memory win over one
// that a more trusted source saved.
var ErrOutranked = errors.New("not resolved: the other memory comes from a more trusted source, so only the user can choose this one")

// Resolve makes the given memory the current fact for its key and marks every
// other live memory with that key as superseded by it. It is for the user,
// through the command line or the review page.
func (s *Store) Resolve(id string) (Memory, []Memory, error) {
	return s.ResolveAs(id, SourceUser)
}

// ResolveAs is Resolve on behalf of actor. Anyone but the user may only pick
// a memory at least as trusted as each memory it would replace.
func (s *Store) ResolveAs(id string, actor Source) (Memory, []Memory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(); err != nil {
		return Memory{}, nil, err
	}
	keep, err := s.find(id)
	if err != nil {
		return Memory{}, nil, err
	}
	if actor != SourceUser {
		if next, ok := s.mems[keep.SupersededBy]; ok && !keep.Live() && next.Source.Rank() > keep.Source.Rank() {
			return Memory{}, nil, ErrOutranked
		}
		for _, ex := range s.mems {
			if ex.ID == keep.ID || !ex.Live() || ex.Source.Rank() <= keep.Source.Rank() {
				continue
			}
			paired := ex.ID == keep.ConflictsWith || ex.ConflictsWith == keep.ID
			sameKey := keep.Key != "" && ex.Key == keep.Key && sameSlot(ex, keep)
			if paired || sameKey {
				return Memory{}, nil, ErrOutranked
			}
		}
	}
	replaced, err := s.resolveLocked(keep, time.Now().UTC())
	if err != nil {
		return Memory{}, nil, err
	}
	return keep.clone(), replaced, nil
}

func (s *Store) resolveLocked(keep *Memory, now time.Time) ([]Memory, error) {
	var replaced []Memory
	// A conflict flagged without a key pairs two memories directly.
	for _, ex := range s.mems {
		if ex.ID == keep.ID || !ex.Live() || (ex.ID != keep.ConflictsWith && ex.ConflictsWith != keep.ID) {
			continue
		}
		if ex.Key != "" && ex.Key == keep.Key {
			continue // the key pass below handles it
		}
		ex.Status = StatusSuperseded
		ex.SupersededBy = keep.ID
		ex.ConflictsWith = ""
		ex.UpdatedAt = now
		ex.InvalidAt = &now
		if err := s.write(ex); err != nil {
			return nil, err
		}
		replaced = append(replaced, ex.clone())
	}
	if keep.Key != "" {
		for _, ex := range s.mems {
			if ex.ID == keep.ID || ex.Key != keep.Key || !sameSlot(ex, keep) || ex.Status == StatusSuperseded || ex.Status == StatusExpired {
				continue
			}
			ex.Status = StatusSuperseded
			ex.SupersededBy = keep.ID
			ex.ConflictsWith = ""
			ex.UpdatedAt = now
			ex.InvalidAt = &now
			if err := s.write(ex); err != nil {
				return nil, err
			}
			replaced = append(replaced, ex.clone())
		}
	}
	keep.Status = StatusActive
	keep.ConflictsWith = ""
	keep.SupersededBy = ""
	keep.InvalidAt = nil
	keep.UpdatedAt = now
	if err := s.write(keep); err != nil {
		return nil, err
	}
	return replaced, nil
}

// ListIn returns the memories in a scope, newest first. Without includeAll
// it returns only active memories.
func (s *Store) ListIn(sc Scope, includeAll bool) ([]Memory, error) {
	sc = sc.Normalized()
	return s.filter(func(m *Memory) bool {
		return (includeAll || m.Status == StatusActive) && sc.Matches(m)
	})
}

// Replace saves in as the newer version of the memory with ID old: the old
// one is kept as history, marked superseded from now on. The new memory
// takes the old one's scope, key, metadata, sources, type and commit unless
// in sets its own, and keeps the old tags along with in's. An edit from a
// less trusted source than the old version is saved as a conflict instead.
func (s *Store) Replace(old string, in RememberInput) (*RememberResult, error) {
	prev, err := s.Get(old)
	if err != nil {
		return nil, err
	}
	if prev.Status == StatusSuperseded || prev.Status == StatusExpired {
		return nil, fmt.Errorf("memory %s is no longer current", prev.ID)
	}
	if in.Project == "" {
		in.Project = prev.Project
	}
	if in.User == "" && in.Agent == "" && in.Session == "" {
		in.User, in.Agent, in.Session = prev.User, prev.Agent, prev.Session
	}
	if in.Key == "" {
		in.Key = prev.Key
	}
	if in.Metadata == nil {
		in.Metadata = prev.Metadata
	}
	// A new version is about the same code, so it keeps what ties the old
	// one to it; without this, a merged or edited fact never goes stale.
	if in.Sources == nil {
		in.Sources = prev.Sources
	}
	if in.Type == "" {
		in.Type = prev.Type
	}
	if in.Commit == "" {
		in.Commit = prev.Commit
	}
	src := in.Source
	if src == "" {
		src = SourceUser
	}
	src, _ = claimed(src, in.Untrusted, nil)
	// A more trusted source restating a guess is sure of it: the old
	// confidence belonged to the guess.
	if in.Confidence == 0 && src.Rank() <= prev.Source.Rank() {
		in.Confidence = prev.Conf
	}
	in.Tags = mergeTags(prev.Tags, in.Tags)
	res, err := s.Remember(in)
	if err != nil {
		return nil, err
	}
	if res.Memory.ID == prev.ID {
		return res, nil // the same text again
	}
	// A less trusted edit never replaces the old version: it waits as a
	// conflict for the user to settle.
	if src.Rank() < prev.Source.Rank() {
		if res.Memory.Status != StatusConflict {
			m, err := s.MarkConflict(res.Memory.ID, prev.ID)
			if err != nil {
				return nil, err
			}
			res.Memory = m
		}
		p := prev
		res.ConflictsWith = &p
		return res, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(); err != nil {
		return nil, err
	}
	ex, ok := s.mems[prev.ID]
	if !ok || ex.Status == StatusSuperseded {
		return res, nil // already replaced, through its key
	}
	now := time.Now().UTC()
	ex.Status = StatusSuperseded
	ex.SupersededBy = res.Memory.ID
	ex.ConflictsWith = ""
	ex.UpdatedAt = now
	ex.InvalidAt = &now
	if err := s.write(ex); err != nil {
		return nil, err
	}
	res.Superseded = append(res.Superseded, ex.clone())
	return res, nil
}

// Expire marks a memory as no longer true, with nothing to replace it. It
// stays on disk as history, so a recall of the past can still find it.
func (s *Store) Expire(id string) (Memory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(); err != nil {
		return Memory{}, err
	}
	m, err := s.find(id)
	if err != nil {
		return Memory{}, err
	}
	if m.Status == StatusExpired {
		return m.clone(), nil
	}
	now := time.Now().UTC()
	m.Status = StatusExpired
	m.ConflictsWith = ""
	m.UpdatedAt = now
	m.InvalidAt = &now
	if err := s.write(m); err != nil {
		return Memory{}, err
	}
	return m.clone(), nil
}

// History returns every version of a memory, oldest first: the memories it
// replaced, the memory itself, and the ones that replaced it.
func (s *Store) History(id string) ([]Memory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(); err != nil {
		return nil, err
	}
	m, err := s.find(id)
	if err != nil {
		return nil, err
	}
	// Walk back to the first version: whatever was superseded by the one we have.
	first := m
	for seen := map[string]bool{first.ID: true}; ; {
		var prev *Memory
		for _, ex := range s.mems {
			if ex.SupersededBy == first.ID && !seen[ex.ID] {
				prev = ex
				break
			}
		}
		if prev == nil {
			break
		}
		seen[prev.ID] = true
		first = prev
	}
	var out []Memory
	for cur, seen := first, map[string]bool{}; cur != nil && !seen[cur.ID]; cur = s.mems[cur.SupersededBy] {
		seen[cur.ID] = true
		out = append(out, cur.clone())
		if cur.SupersededBy == "" {
			break
		}
	}
	return out, nil
}

// Supersede marks the memory old as replaced by the memory by, which must
// already exist. Both stay on disk; old is kept as history from now on.
func (s *Store) Supersede(old, by string) (Memory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(); err != nil {
		return Memory{}, err
	}
	o, err := s.find(old)
	if err != nil {
		return Memory{}, err
	}
	b, err := s.find(by)
	if err != nil {
		return Memory{}, err
	}
	if o.ID == b.ID {
		return Memory{}, errors.New("a memory can't replace itself")
	}
	if !b.Live() {
		return Memory{}, fmt.Errorf("memory %s is no longer current", b.ID)
	}
	if b.SupersededBy == o.ID {
		return Memory{}, errors.New("these two memories would replace each other")
	}
	now := time.Now().UTC()
	o.Status = StatusSuperseded
	o.SupersededBy = b.ID
	o.ConflictsWith = ""
	o.UpdatedAt = now
	o.InvalidAt = &now
	if err := s.write(o); err != nil {
		return Memory{}, err
	}
	return o.clone(), nil
}

// Merge saves in as one memory that replaces every memory in ids, which are
// kept as history. The new memory takes the scope, type, commit and
// confidence of the first one unless in sets its own. Its sources and tags
// are in's plus those of every memory merged.
func (s *Store) Merge(ids []string, in RememberInput) (*RememberResult, error) {
	if len(ids) < 2 {
		return nil, errors.New("merging needs at least two memories")
	}
	// The merged memory points at everything its parts pointed at.
	refs := append([]Ref(nil), in.Sources...)
	for _, id := range ids {
		m, err := s.Get(id)
		if err != nil {
			return nil, err
		}
		refs = append(refs, m.Sources...)
		in.Tags = mergeTags(in.Tags, m.Tags)
	}
	if len(refs) > 0 {
		in.Sources = refs
	}
	res, err := s.Replace(ids[0], in)
	if err != nil {
		return nil, err
	}
	for _, id := range ids[1:] {
		if id == res.Memory.ID {
			continue
		}
		m, err := s.Supersede(id, res.Memory.ID)
		if err != nil {
			return nil, err
		}
		res.Superseded = append(res.Superseded, m)
	}
	return res, nil
}

// Link ties two memories together both ways, so recalling one brings the
// other along. Linking twice changes nothing.
func (s *Store) Link(a, b string) (Memory, Memory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(); err != nil {
		return Memory{}, Memory{}, err
	}
	ma, err := s.find(a)
	if err != nil {
		return Memory{}, Memory{}, err
	}
	mb, err := s.find(b)
	if err != nil {
		return Memory{}, Memory{}, err
	}
	if ma.ID == mb.ID {
		return Memory{}, Memory{}, errors.New("a memory can't link to itself")
	}
	if !sameSlot(&Memory{User: ma.User, Agent: ma.Agent, Session: ma.Session}, &Memory{User: mb.User, Agent: mb.Agent, Session: mb.Session}) {
		return Memory{}, Memory{}, errors.New("only memories with the same user, agent and run can be linked")
	}
	now := time.Now().UTC()
	for _, p := range [][2]*Memory{{ma, mb}, {mb, ma}} {
		if hasID(p[0].Links, p[1].ID) {
			continue
		}
		p[0].Links = append(p[0].Links, p[1].ID)
		p[0].UpdatedAt = now
		if err := s.write(p[0]); err != nil {
			return Memory{}, Memory{}, err
		}
	}
	return ma.clone(), mb.clone(), nil
}

func hasID(ids []string, id string) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

func removeID(ids []string, id string) []string {
	var out []string
	for _, x := range ids {
		if x != id {
			out = append(out, x)
		}
	}
	return out
}

// MarkConflict flags the memory id as disagreeing with the memory with, so
// the user picks which is true. Resolve settles it.
func (s *Store) MarkConflict(id, with string) (Memory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(); err != nil {
		return Memory{}, err
	}
	m, err := s.find(id)
	if err != nil {
		return Memory{}, err
	}
	w, err := s.find(with)
	if err != nil {
		return Memory{}, err
	}
	if m.ID == w.ID {
		return Memory{}, errors.New("a memory can't conflict with itself")
	}
	m.Status = StatusConflict
	m.ConflictsWith = w.ID
	m.UpdatedAt = time.Now().UTC()
	if err := s.write(m); err != nil {
		return Memory{}, err
	}
	return m.clone(), nil
}
