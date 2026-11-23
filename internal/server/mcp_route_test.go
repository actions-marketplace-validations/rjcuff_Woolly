package server

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rjcuff/woolly/internal/mcp"
	"github.com/rjcuff/woolly/internal/memory"
	"github.com/rjcuff/woolly/internal/recall"
)

const mcpInit = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"open-webui","version":"1"}}}`

func setupMCP(t *testing.T, withMCP bool) (http.Handler, string) {
	t.Helper()
	home := t.TempDir()
	store, err := memory.Open(filepath.Join(home, "memories"))
	if err != nil {
		t.Fatal(err)
	}
	token, err := EnsureToken(home)
	if err != nil {
		t.Fatal(err)
	}
	engine := recall.New(store, nil, "")
	opts := Options{}
	if withMCP {
		opts.MCP = &mcp.HTTPHandler{New: func(*http.Request) (*mcp.Server, error) {
			return &mcp.Server{Store: store, Engine: engine, Version: "test"}, nil
		}}
	}
	return Handler(store, engine, "test", token, opts), token
}

func TestMCPOverHTTPNeedsTheToken(t *testing.T) {
	h, token := setupMCP(t, true)
	for _, hdr := range []map[string]string{nil, {"Authorization": "Bearer wrong"}} {
		if rec := do(h, "POST", MCPPath, mcpInit, hdr); rec.Code != http.StatusUnauthorized {
			t.Fatalf("%v: %d", hdr, rec.Code)
		}
	}
	rec := do(h, "POST", MCPPath, mcpInit, map[string]string{"Authorization": "Bearer " + token})
	if rec.Code != http.StatusOK || rec.Header().Get(mcp.SessionHeader) == "" || !strings.Contains(rec.Body.String(), `"serverInfo"`) {
		t.Fatalf("with token: %d %s", rec.Code, rec.Body.String())
	}
	id := rec.Header().Get(mcp.SessionHeader)
	rec = do(h, "POST", MCPPath, `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`, map[string]string{"Authorization": "Bearer " + token, mcp.SessionHeader: id})
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"recall"`) {
		t.Fatalf("tools/list: %d %s", rec.Code, rec.Body.String())
	}
	// A session ID alone is not enough.
	if rec := do(h, "POST", MCPPath, `{"jsonrpc":"2.0","id":3,"method":"tools/list"}`, map[string]string{mcp.SessionHeader: id}); rec.Code != http.StatusUnauthorized {
		t.Fatalf("session without token: %d", rec.Code)
	}
}

func TestMCPOverHTTPBlocksOtherSites(t *testing.T) {
	h, token := setupMCP(t, true)
	auth := "Bearer " + token
	for _, hdr := range []map[string]string{
		{"Authorization": auth, "Origin": "https://evil.example"},
		{"Authorization": auth, "Sec-Fetch-Site": "cross-site"},
	} {
		if rec := do(h, "POST", MCPPath, mcpInit, hdr); rec.Code != http.StatusForbidden {
			t.Fatalf("%v: %d", hdr, rec.Code)
		}
	}
	// DNS rebinding: a page on another name that resolves to 127.0.0.1.
	req := httptest.NewRequest("POST", MCPPath, strings.NewReader(mcpInit))
	req.Host = "evil.example:11435"
	req.Header.Set("Authorization", auth)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("non-local host: %d", rec.Code)
	}
}

func TestMCPRouteIsOffWithoutTheOption(t *testing.T) {
	h, token := setupMCP(t, false)
	if rec := do(h, "POST", MCPPath, mcpInit, map[string]string{"Authorization": "Bearer " + token}); rec.Code == http.StatusOK {
		t.Fatalf("/mcp should not answer without Options.MCP: %d %s", rec.Code, rec.Body.String())
	}
}
