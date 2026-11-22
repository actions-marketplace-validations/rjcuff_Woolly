package server

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rjcuff/woolly/internal/dream"
	"github.com/rjcuff/woolly/internal/learn"
	"github.com/rjcuff/woolly/internal/memory"
	"github.com/rjcuff/woolly/internal/recall"
)

// v1Memory is a memory in the shape mem0's API uses, so apps written for
// mem0 can read Woolly's answers.
type v1Memory struct {
	ID        string         `json:"id"`
	Memory    string         `json:"memory"`
	UserID    string         `json:"user_id,omitempty"`
	AgentID   string         `json:"agent_id,omitempty"`
	RunID     string         `json:"run_id,omitempty"`
	Project   string         `json:"project,omitempty"`
	Type      string         `json:"type,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Status    memory.Status  `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	InvalidAt *time.Time     `json:"invalid_at,omitempty"`
	Score     *float64       `json:"score,omitempty"`
}

func toV1(m memory.Memory) v1Memory {
	return v1Memory{
		ID: m.ID, Memory: m.Content, UserID: m.User, AgentID: m.Agent, RunID: m.Session, Project: m.Project, Type: memory.TypeOf(m),
		Metadata: m.Metadata, Status: m.Status, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt, InvalidAt: m.InvalidAt,
	}
}

// v1Event is one entry in a memory's history.
type v1Event struct {
	ID        string    `json:"id"`
	MemoryID  string    `json:"memory_id"`
	OldMemory *string   `json:"old_memory"`
	NewMemory *string   `json:"new_memory"`
	Event     string    `json:"event"`
	CreatedAt time.Time `json:"created_at"`
}

// scopeFields are the scope fields every v1 request can carry.
type scopeFields struct {
	UserID  string `json:"user_id"`
	AgentID string `json:"agent_id"`
	RunID   string `json:"run_id"`
	Project string `json:"project"`
	// Type filters search and list results by memory type.
	Type []string `json:"type,omitempty"`
}

func (f scopeFields) scope() memory.Scope {
	return memory.Scope{User: f.UserID, Agent: f.AgentID, Session: f.RunID, Project: f.Project, Types: f.Type}.Normalized()
}

func queryScope(r *http.Request) memory.Scope {
	q := r.URL.Query()
	return memory.Scope{User: q.Get("user_id"), Agent: q.Get("agent_id"), Session: q.Get("run_id"), Project: q.Get("project")}.Normalized()
}

// getInScope loads the memory named in the path. When the request names a
// user, agent or session, a memory outside it is reported as not found, so an
// app serving several users can't reach one user's memories from another's.
func getInScope(w http.ResponseWriter, r *http.Request, store *memory.Store) (memory.Memory, bool) {
	m, err := store.Get(r.PathValue("id"))
	if err == nil {
		if sc := queryScope(r); !sc.Empty() && !sc.Matches(&m) {
			err = memory.ErrNotFound
		}
	}
	if err != nil {
		writeStoreError(w, err)
		return memory.Memory{}, false
	}
	return m, true
}

// registerV1 adds the memory layer API for apps and agents: mem0's routes
// and field names, on top of Woolly's store.
func registerV1(mux *http.ServeMux, store *memory.Store, engine *recall.Engine, opts Options, record func(*http.Request, string, []string, string, int)) {
	learner := opts.Learner
	if learner == nil {
		// Saving messages as they are needs no model.
		learner = &learn.Learner{Store: store, Engine: engine}
	}

	mux.HandleFunc("POST /v1/memories", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			scopeFields
			Messages []learn.Message `json:"messages"`
			Metadata map[string]any  `json:"metadata"`
			Infer    *bool           `json:"infer"`
			// MemType is the type of what gets saved, like "procedure".
			MemType string `json:"memory_type"`
		}
		if !readJSON(w, r, &req) {
			return
		}
		if len(req.Messages) == 0 {
			writeError(w, http.StatusBadRequest, "messages is empty")
			return
		}
		infer := req.Infer == nil || *req.Infer
		events, err := learner.Add(r.Context(), req.Messages, learn.AddOptions{
			Scope: req.scope(), Via: appOf(r), Metadata: req.Metadata, Infer: infer, Type: req.MemType,
		})
		if err != nil {
			status := http.StatusBadRequest
			if errors.Is(err, learn.ErrNoModel) || infer {
				status = http.StatusServiceUnavailable
				err = errors.New(err.Error() + ". Inferring needs Ollama with a chat model running, or send \"infer\": false")
			}
			writeError(w, status, err.Error())
			return
		}
		var ids []string
		for _, e := range events {
			if e.Event != learn.EventNone {
				ids = append(ids, e.ID)
			}
		}
		if len(ids) > 0 {
			record(r, "learn", ids, "", len(ids))
		}
		writeJSON(w, http.StatusOK, map[string]any{"results": events})
	})

	mux.HandleFunc("GET /v1/memories", func(w http.ResponseWriter, r *http.Request) {
		all := r.URL.Query().Get("show_expired") == "true"
		mems, err := store.ListIn(queryScope(r), all)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if k, err := strconv.Atoi(r.URL.Query().Get("top_k")); err == nil && k > 0 && k < len(mems) {
			mems = mems[:k]
		}
		out := make([]v1Memory, len(mems))
		for i, m := range mems {
			out[i] = toV1(m)
		}
		record(r, "list", nil, "", len(out))
		writeJSON(w, http.StatusOK, map[string]any{"results": out})
	})

	mux.HandleFunc("DELETE /v1/memories", func(w http.ResponseWriter, r *http.Request) {
		sc := queryScope(r)
		if sc.Empty() {
			writeError(w, http.StatusBadRequest, "give a user_id, agent_id or run_id to delete its memories")
			return
		}
		mems, err := store.ListIn(sc, true)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		var ids []string
		for _, m := range mems {
			if _, err := store.Forget(m.ID); err != nil {
				writeStoreError(w, err)
				return
			}
			ids = append(ids, m.ID)
		}
		record(r, "forget", ids, "", len(ids))
		writeJSON(w, http.StatusOK, map[string]any{"message": "All relevant memories deleted", "deleted": len(ids)})
	})

	mux.HandleFunc("GET /v1/memories/{id}", func(w http.ResponseWriter, r *http.Request) {
		m, ok := getInScope(w, r, store)
		if !ok {
			return
		}
		writeJSON(w, http.StatusOK, toV1(m))
	})

	mux.HandleFunc("PUT /v1/memories/{id}", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Text     *string        `json:"text"`
			Metadata map[string]any `json:"metadata"`
		}
		if !readJSON(w, r, &req) {
			return
		}
		old, ok := getInScope(w, r, store)
		if !ok {
			return
		}
		text := old.Content
		if req.Text != nil {
			text = *req.Text
		}
		meta := old.Metadata
		if req.Metadata != nil {
			meta = req.Metadata
		}
		// A new version, with the old one kept as history.
		res, err := store.Replace(old.ID, memory.RememberInput{Content: text, Source: old.Source, Via: appOf(r), Metadata: meta, Untrusted: true})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		record(r, "update", []string{res.Memory.ID}, "", 0)
		writeJSON(w, http.StatusOK, toV1(res.Memory))
	})

	mux.HandleFunc("DELETE /v1/memories/{id}", func(w http.ResponseWriter, r *http.Request) {
		in, ok := getInScope(w, r, store)
		if !ok {
			return
		}
		m, err := store.Forget(in.ID)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		record(r, "forget", []string{m.ID}, "", 0)
		writeJSON(w, http.StatusOK, map[string]any{"message": "Memory deleted successfully", "id": m.ID})
	})

	mux.HandleFunc("GET /v1/memories/{id}/history", func(w http.ResponseWriter, r *http.Request) {
		in, ok := getInScope(w, r, store)
		if !ok {
			return
		}
		versions, err := store.History(in.ID)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"results": history(versions)})
	})

	mux.HandleFunc("POST /v1/dream", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			scopeFields
			DryRun bool  `json:"dry_run"`
			Learn  *bool `json:"learn"`
		}
		if !readJSON(w, r, &req) {
			return
		}
		d := opts.Dreamer
		if d == nil {
			d = &dream.Dreamer{Store: store, Engine: engine}
		}
		rep, err := d.Dream(r.Context(), dream.Options{Scope: req.scope(), DryRun: req.DryRun, Learn: req.Learn == nil || *req.Learn})
		if errors.Is(err, dream.ErrBusy) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if !req.DryRun {
			var ids []string
			for _, st := range rep.Steps {
				if st.Result != "" {
					ids = append(ids, st.Result)
				}
			}
			if len(ids) > 0 {
				record(r, "dream", ids, "", len(ids))
			}
		}
		writeJSON(w, http.StatusOK, rep)
	})

	mux.HandleFunc("POST /v1/search", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			scopeFields
			Query       string      `json:"query"`
			Filters     scopeFields `json:"filters"`
			TopK        int         `json:"top_k"`
			Threshold   float64     `json:"threshold"`
			ShowExpired bool        `json:"show_expired"`
		}
		if !readJSON(w, r, &req) {
			return
		}
		// mem0 moved the scope into filters; either place works here.
		f := req.Filters
		if f.UserID == "" {
			f.UserID = req.UserID
		}
		if f.AgentID == "" {
			f.AgentID = req.AgentID
		}
		if f.RunID == "" {
			f.RunID = req.RunID
		}
		if f.Project == "" {
			f.Project = req.Project
		}
		if len(f.Type) == 0 {
			f.Type = req.Type
		}
		results, _, err := engine.RecallScoped(r.Context(), req.Query, req.TopK, req.ShowExpired, f.scope())
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		out := []v1Memory{}
		var ids []string
		for _, res := range results {
			if res.Score < req.Threshold {
				continue
			}
			m := toV1(res.Memory)
			score := res.Score
			m.Score = &score
			out = append(out, m)
			ids = append(ids, res.Memory.ID)
		}
		record(r, "recall", ids, req.Query, len(out))
		writeJSON(w, http.StatusOK, map[string]any{"results": out})
	})
}

// history turns a memory's versions, oldest first, into mem0-style events.
func history(versions []memory.Memory) []v1Event {
	out := []v1Event{}
	var prev *memory.Memory
	for i := range versions {
		m := versions[i]
		content := m.Content
		ev := v1Event{ID: m.ID + ":" + strconv.Itoa(len(out)), MemoryID: m.ID, NewMemory: &content, Event: "ADD", CreatedAt: m.CreatedAt}
		if prev != nil {
			old := prev.Content
			ev.Event, ev.OldMemory = "UPDATE", &old
		}
		out = append(out, ev)
		if m.Status == memory.StatusExpired {
			old := m.Content
			at := m.UpdatedAt
			if m.InvalidAt != nil {
				at = *m.InvalidAt
			}
			out = append(out, v1Event{ID: m.ID + ":" + strconv.Itoa(len(out)), MemoryID: m.ID, OldMemory: &old, Event: "DELETE", CreatedAt: at})
		}
		prev = &versions[i]
	}
	return out
}

// isV1 reports whether a path is part of the memory layer API.
func isV1(path string) bool { return strings.HasPrefix(path, "/v1/") }
