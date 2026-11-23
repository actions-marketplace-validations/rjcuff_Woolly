package mcp

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"
)

// SessionHeader carries the session ID in MCP's Streamable HTTP transport.
const SessionHeader = "Mcp-Session-Id"

const (
	// DefaultSessionTTL is how long an idle HTTP session is kept.
	DefaultSessionTTL = 30 * time.Minute
	// DefaultMaxSessions caps how many HTTP sessions are open at once. The
	// one idle longest is closed to make room.
	DefaultMaxSessions = 64
	// maxMessage is the largest request body accepted.
	maxMessage = 1 << 20
)

// ErrBadRequest wraps errors a session factory returns for bad input, like
// a blank ?agent=. They become 400 responses.
var ErrBadRequest = errors.New("bad request")

// HTTPHandler serves MCP over Streamable HTTP: one JSON-RPC message per
// POST, answered with plain JSON. It never streams, so GET is refused.
// Each session gets its own Server from New, because a Server keeps
// per-connection state.
type HTTPHandler struct {
	// New makes the Server for a session from the initialize request. It
	// may read the URL, for example to scope the session to one agent.
	New func(r *http.Request) (*Server, error)
	// TTL and Max default to DefaultSessionTTL and DefaultMaxSessions.
	TTL time.Duration
	Max int
	// now is time.Now, replaceable in tests.
	now func() time.Time

	mu       sync.Mutex
	sessions map[string]*session
}

type session struct {
	mu   sync.Mutex // one message at a time per session
	srv  *Server
	last time.Time
}

func (h *HTTPHandler) clock() time.Time {
	if h.now != nil {
		return h.now()
	}
	return time.Now()
}

func (h *HTTPHandler) ttl() time.Duration {
	if h.TTL > 0 {
		return h.TTL
	}
	return DefaultSessionTTL
}

func (h *HTTPHandler) max() int {
	if h.Max > 0 {
		return h.Max
	}
	return DefaultMaxSessions
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.post(w, r)
	case http.MethodDelete:
		if h.drop(r.Header.Get(SessionHeader)) {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "unknown session", http.StatusNotFound)
	default:
		// No server-to-client stream: every answer comes back on its POST.
		w.Header().Set("Allow", "POST, DELETE")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *HTTPHandler) post(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, maxMessage+1))
	if err != nil {
		http.Error(w, "couldn't read the request", http.StatusBadRequest)
		return
	}
	if len(body) > maxMessage {
		http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
		return
	}
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) > 0 && trimmed[0] == '[' {
		// Batches were dropped from MCP in 2025-06-18.
		writeRPC(w, http.StatusBadRequest, mustJSON(response{JSONRPC: "2.0", ID: json.RawMessage("null"), Error: &rpcError{-32600, "batches aren't supported"}}))
		return
	}
	var peek struct {
		Method string `json:"method"`
	}
	_ = json.Unmarshal(trimmed, &peek)

	var s *session
	if peek.Method == "initialize" {
		srv, err := h.New(r)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, ErrBadRequest) {
				status = http.StatusBadRequest
			}
			http.Error(w, err.Error(), status)
			return
		}
		id, created, err := h.open(srv)
		if err != nil {
			http.Error(w, "couldn't start a session", http.StatusInternalServerError)
			return
		}
		w.Header().Set(SessionHeader, id)
		// Use the session just made, even if another initialize closed it
		// to make room in the meantime.
		s = created
	} else {
		id := r.Header.Get(SessionHeader)
		if id == "" {
			http.Error(w, "missing "+SessionHeader+" header; send initialize first", http.StatusBadRequest)
			return
		}
		if s = h.get(id); s == nil {
			// Tells the client to start a new session.
			http.Error(w, "unknown or expired session", http.StatusNotFound)
			return
		}
	}

	s.mu.Lock()
	reply, ok := s.srv.Handle(r.Context(), trimmed)
	s.mu.Unlock()
	if !ok {
		w.WriteHeader(http.StatusAccepted)
		return
	}
	writeRPC(w, http.StatusOK, reply)
}

func writeRPC(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

// open stores a new session and returns its ID, first closing sessions that
// went idle and, when full, the one idle longest.
func (h *HTTPHandler) open(srv *Server) (string, *session, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", nil, err
	}
	id := hex.EncodeToString(b[:])
	now := h.clock()
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.sessions == nil {
		h.sessions = map[string]*session{}
	}
	h.sweepLocked(now)
	for len(h.sessions) >= h.max() {
		oldest := ""
		for k, s := range h.sessions {
			if oldest == "" || s.last.Before(h.sessions[oldest].last) {
				oldest = k
			}
		}
		delete(h.sessions, oldest)
	}
	s := &session{srv: srv, last: now}
	h.sessions[id] = s
	return id, s, nil
}

// get returns a live session and marks it used, or nil.
func (h *HTTPHandler) get(id string) *session {
	now := h.clock()
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sweepLocked(now)
	s := h.sessions[id]
	if s != nil {
		s.last = now
	}
	return s
}

func (h *HTTPHandler) drop(id string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.sessions[id]; !ok || id == "" {
		return false
	}
	delete(h.sessions, id)
	return true
}

func (h *HTTPHandler) sweepLocked(now time.Time) {
	for k, s := range h.sessions {
		if now.Sub(s.last) > h.ttl() {
			delete(h.sessions, k)
		}
	}
}

// Sessions is how many sessions are open, for tests and diagnostics.
func (h *HTTPHandler) Sessions() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.sessions)
}
