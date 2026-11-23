package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rjcuff/woolly/internal/memory"
	"github.com/rjcuff/woolly/internal/recall"
)

func httpHandler(t *testing.T) (*HTTPHandler, *memory.Store) {
	t.Helper()
	store, err := memory.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	engine := recall.New(store, nil, "")
	h := &HTTPHandler{New: func(r *http.Request) (*Server, error) {
		agent := r.URL.Query().Get("agent")
		if r.URL.Query().Has("agent") && strings.TrimSpace(agent) == "" {
			return nil, fmt.Errorf("%w: agent is blank", ErrBadRequest)
		}
		return &Server{Store: store, Engine: engine, Version: "test", Agent: agent}, nil
	}}
	return h, store
}

func post(h http.Handler, url, session, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if session != "" {
		req.Header.Set(SessionHeader, session)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

const initMsg = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"open-webui","version":"1"}}}`

func initSession(t *testing.T, h http.Handler, url string) string {
	t.Helper()
	rec := post(h, url, "", initMsg)
	if rec.Code != http.StatusOK {
		t.Fatalf("initialize: %d %s", rec.Code, rec.Body.String())
	}
	id := rec.Header().Get(SessionHeader)
	if len(id) != 32 {
		t.Fatalf("session id = %q", id)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content type = %q", ct)
	}
	return id
}

func TestHTTPHandshakeRememberRecall(t *testing.T) {
	h, _ := httpHandler(t)
	id := initSession(t, h, "/mcp")

	if rec := post(h, "/mcp", id, `{"jsonrpc":"2.0","method":"notifications/initialized"}`); rec.Code != http.StatusAccepted || rec.Body.Len() != 0 {
		t.Fatalf("notification: %d %q", rec.Code, rec.Body.String())
	}
	rec := post(h, "/mcp", id, `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"remember","arguments":{"content":"Prefers concise answers","source":"user","scope":"personal"}}}`)
	if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), `"isError"`) {
		t.Fatalf("remember: %d %s", rec.Code, rec.Body.String())
	}
	rec = post(h, "/mcp", id, `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"recall","arguments":{"query":"concise"}}}`)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Prefers concise answers") {
		t.Fatalf("recall: %d %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		ID json.RawMessage `json:"id"`
	}
	if json.Unmarshal(rec.Body.Bytes(), &resp) != nil || string(resp.ID) != "3" {
		t.Fatalf("reply id = %s", resp.ID)
	}
}

func TestHTTPSessionsAreSeparate(t *testing.T) {
	h, store := httpHandler(t)
	mine := initSession(t, h, "/mcp")
	bot := initSession(t, h, "/mcp?agent=support-bot")
	if mine == bot {
		t.Fatal("sessions share an id")
	}
	post(h, "/mcp", bot, `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"remember","arguments":{"content":"Customer prefers email"}}}`)
	mems, err := store.List(false)
	if err != nil {
		t.Fatal(err)
	}
	if len(mems) != 1 || mems[0].Agent != "support-bot" || mems[0].Via != "open-webui" {
		t.Fatalf("saved = %+v", mems)
	}
	rec := post(h, "/mcp", mine, `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"recall","arguments":{"query":"email"}}}`)
	if strings.Contains(rec.Body.String(), "Customer prefers email") {
		t.Fatalf("your own session should not see the agent's memory: %s", rec.Body.String())
	}
}

func TestHTTPRejectsBadRequests(t *testing.T) {
	h, _ := httpHandler(t)
	cases := []struct {
		name, url, session, body string
		want                     int
	}{
		{"no session", "/mcp", "", `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`, http.StatusBadRequest},
		{"unknown session", "/mcp", strings.Repeat("a", 32), `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`, http.StatusNotFound},
		{"batch", "/mcp", "", `[` + initMsg + `]`, http.StatusBadRequest},
		{"blank scope", "/mcp?agent=%20", "", initMsg, http.StatusBadRequest},
		{"too large", "/mcp", "", `{"x":"` + strings.Repeat("a", maxMessage) + `"}`, http.StatusRequestEntityTooLarge},
	}
	for _, c := range cases {
		if rec := post(h, c.url, c.session, c.body); rec.Code != c.want {
			t.Errorf("%s: %d, want %d (%s)", c.name, rec.Code, c.want, rec.Body.String())
		}
	}
	if h.Sessions() != 0 {
		t.Fatalf("bad requests opened %d sessions", h.Sessions())
	}

	get := httptest.NewRecorder()
	h.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/mcp", nil))
	if get.Code != http.StatusMethodNotAllowed || get.Header().Get("Allow") != "POST, DELETE" {
		t.Fatalf("GET: %d allow=%q", get.Code, get.Header().Get("Allow"))
	}
}

func TestHTTPFactoryErrorIsNotABadRequest(t *testing.T) {
	h := &HTTPHandler{New: func(*http.Request) (*Server, error) { return nil, errors.New("store is closed") }}
	if rec := post(h, "/mcp", "", initMsg); rec.Code != http.StatusInternalServerError {
		t.Fatalf("got %d", rec.Code)
	}
}

func TestHTTPDeleteEndsSession(t *testing.T) {
	h, _ := httpHandler(t)
	id := initSession(t, h, "/mcp")
	del := func() int {
		req := httptest.NewRequest(http.MethodDelete, "/mcp", nil)
		req.Header.Set(SessionHeader, id)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}
	if code := del(); code != http.StatusOK {
		t.Fatalf("delete: %d", code)
	}
	if code := del(); code != http.StatusNotFound {
		t.Fatalf("second delete: %d", code)
	}
	if rec := post(h, "/mcp", id, `{"jsonrpc":"2.0","id":2,"method":"ping"}`); rec.Code != http.StatusNotFound {
		t.Fatalf("ping after delete: %d", rec.Code)
	}
}

func TestHTTPIdleSessionsExpireAndTheCapHolds(t *testing.T) {
	h, _ := httpHandler(t)
	now := time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)
	h.now = func() time.Time { return now }
	h.Max = 3

	first := initSession(t, h, "/mcp")
	now = now.Add(time.Minute)
	initSession(t, h, "/mcp")
	now = now.Add(time.Minute)
	initSession(t, h, "/mcp")
	now = now.Add(time.Minute)
	// Using the first one makes it the newest, so the second is closed next.
	if rec := post(h, "/mcp", first, `{"jsonrpc":"2.0","id":2,"method":"ping"}`); rec.Code != http.StatusOK {
		t.Fatalf("ping: %d", rec.Code)
	}
	initSession(t, h, "/mcp")
	if h.Sessions() != 3 {
		t.Fatalf("sessions = %d, want the cap of 3", h.Sessions())
	}
	if rec := post(h, "/mcp", first, `{"jsonrpc":"2.0","id":3,"method":"ping"}`); rec.Code != http.StatusOK {
		t.Fatalf("the recently used session was closed: %d", rec.Code)
	}

	now = now.Add(DefaultSessionTTL + time.Second)
	if rec := post(h, "/mcp", first, `{"jsonrpc":"2.0","id":4,"method":"ping"}`); rec.Code != http.StatusNotFound {
		t.Fatalf("idle session should expire: %d", rec.Code)
	}
	if h.Sessions() != 0 {
		t.Fatalf("sessions = %d after expiry", h.Sessions())
	}
}

func TestHTTPConcurrentSessions(t *testing.T) {
	h, _ := httpHandler(t)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			rec := post(h, "/mcp", "", initMsg)
			id := rec.Header().Get(SessionHeader)
			for j := 0; j < 5; j++ {
				body := fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"method":"tools/call","params":{"name":"remember","arguments":{"content":"Note %d from session %d","scope":"personal"}}}`, j+2, j, i)
				if rec := post(h, "/mcp", id, body); rec.Code != http.StatusOK {
					t.Errorf("session %d: %d", i, rec.Code)
				}
			}
		}(i)
	}
	wg.Wait()
}

func TestNoSourcesDropsRefsAndFolderTools(t *testing.T) {
	store, err := memory.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	s := &Server{Store: store, Engine: recall.New(store, nil, ""), Version: "test", Project: "acme", NoSources: true}
	reply, _ := s.Handle(context.Background(), []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
	if strings.Contains(string(reply), `"name":"why"`) || strings.Contains(string(reply), `"name":"check_memory_files"`) || !strings.Contains(string(reply), `"name":"recall"`) {
		t.Fatalf("tools/list: %s", reply)
	}
	s.Handle(context.Background(), []byte(`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"remember","arguments":{"content":"Auth lives in requireUser","scope":"project","sources":["src/auth.ts#requireUser"]}}}`))
	mems, _ := store.List(false)
	if len(mems) != 1 || len(mems[0].Sources) != 0 {
		t.Fatalf("saved = %+v", mems)
	}
}
