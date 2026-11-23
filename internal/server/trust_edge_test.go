package server

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/rjcuff/woolly/internal/activity"
	"github.com/rjcuff/woolly/internal/memory"
	"github.com/rjcuff/woolly/internal/recall"
)

// setupStore is setup with the store exposed, so tests can seed and inspect
// memories directly.
func setupStore(t *testing.T) (http.Handler, string, *memory.Store) {
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
	return Handler(store, recall.New(store, nil, ""), "test", token, Options{Activity: activity.Open(home)}), token, store
}

func TestCookieAndBearerTogetherCountAsAnApp(t *testing.T) {
	h, token := setup(t)
	both := map[string]string{"Authorization": "Bearer " + token, "Cookie": tokenCookie + "=" + token}
	body := `{"content":"` + attack + `","source":"user"}`
	if rec := do(h, "POST", "/api/remember", body, both); rec.Code != http.StatusBadRequest {
		t.Fatalf("cookie+bearer remember: %d %s", rec.Code, rec.Body.String())
	}
}

func TestClientHeaderCannotClaimTheReviewPage(t *testing.T) {
	h, token := setup(t)
	body := `{"content":"` + attack + `","source":"user"}`
	for _, name := range []string{ReviewPage, "review page", "Review Page", "ui"} {
		hdr := map[string]string{"Authorization": "Bearer " + token, "X-Woolly-Client": name}
		if rec := do(h, "POST", "/api/remember", body, hdr); rec.Code != http.StatusBadRequest {
			t.Fatalf("X-Woolly-Client %q: %d %s", name, rec.Code, rec.Body.String())
		}
	}
	// And without any token at all, the header gets nothing.
	if rec := do(h, "POST", "/api/remember", body, map[string]string{"X-Woolly-Client": ReviewPage}); rec.Code != http.StatusUnauthorized {
		t.Fatalf("no token: %d", rec.Code)
	}
}

func TestBogusCookieIsRejected(t *testing.T) {
	h, _ := setup(t)
	body := `{"content":"` + attack + `","source":"user"}`
	if rec := do(h, "POST", "/api/remember", body, map[string]string{"Cookie": tokenCookie + "=nope"}); rec.Code != http.StatusUnauthorized {
		t.Fatalf("bogus cookie: %d %s", rec.Code, rec.Body.String())
	}
}

func TestResolveWithCookieAndBearerIsAnApp(t *testing.T) {
	h, token, store := setupStore(t)
	store.Remember(memory.RememberInput{Content: "Lives in Austin", Key: "user.location", Source: memory.SourceUser})
	agent, _ := store.Remember(memory.RememberInput{Content: "Lives in Denver", Key: "user.location", Source: memory.SourceAgent})
	both := map[string]string{"Authorization": "Bearer " + token, "Cookie": tokenCookie + "=" + token}
	if rec := do(h, "POST", "/api/resolve", `{"keep_id":"`+agent.Memory.ID+`"}`, both); rec.Code != http.StatusBadRequest {
		t.Fatalf("resolve: %d %s", rec.Code, rec.Body.String())
	}
	if got, _ := store.Get(agent.Memory.ID); got.Status != memory.StatusConflict {
		t.Fatalf("agent memory status = %s", got.Status)
	}
}

func TestPatchChecksAppsButNotTheReviewPage(t *testing.T) {
	h, token, store := setupStore(t)
	res, _ := store.Remember(memory.RememberInput{Content: "Prefers tea"})
	path := "/api/memories/" + res.Memory.ID
	app := map[string]string{"Authorization": "Bearer " + token}
	if rec := do(h, "PATCH", path, `{"content":"`+attack+`"}`, app); rec.Code != http.StatusBadRequest {
		t.Fatalf("app PATCH with instructions: %d %s", rec.Code, rec.Body.String())
	}
	if got, _ := store.Get(res.Memory.ID); got.Content != "Prefers tea" {
		t.Fatalf("content changed: %q", got.Content)
	}
	if rec := do(h, "PATCH", path, `{"content":"Prefers green tea"}`, app); rec.Code != http.StatusOK {
		t.Fatalf("app PATCH with plain text: %d %s", rec.Code, rec.Body.String())
	}
	if rec := do(h, "PATCH", path, `{"content":"`+attack+`"}`, map[string]string{"Cookie": tokenCookie + "=" + token}); rec.Code != http.StatusOK {
		t.Fatalf("review page PATCH: %d %s", rec.Code, rec.Body.String())
	}
}

func TestV1PutInScopeWorksAndKeepsScope(t *testing.T) {
	h, token := setup(t)
	add := v1(t, h, token, "POST", "/v1/memories", `{"messages":[{"role":"user","content":"Deploys the API to Fly.io"}],"user_id":"alice","infer":false}`)
	id := results(add)[0]["id"].(string)
	put := v1(t, h, token, "PUT", "/v1/memories/"+id+"?user_id=alice", `{"text":"Deploys the API to Vercel"}`)
	newID, _ := put["id"].(string)
	if newID == "" {
		t.Fatalf("PUT response: %+v", put)
	}
	got := v1(t, h, token, "GET", "/v1/memories/"+newID+"?user_id=alice", "")
	if got["memory"] != "Deploys the API to Vercel" || got["user_id"] != "alice" {
		t.Fatalf("new version: %+v", got)
	}
	hist := do(h, "GET", "/v1/memories/"+newID+"/history?user_id=alice", "", map[string]string{"Authorization": "Token " + token})
	if hist.Code != http.StatusOK {
		t.Fatalf("history in scope: %d %s", hist.Code, hist.Body.String())
	}
}

func TestV1OutOfScopeDeleteLeavesMemory(t *testing.T) {
	h, token, store := setupStore(t)
	res, _ := store.Remember(memory.RememberInput{Content: "Alice likes tea", User: "alice", Agent: "helper", Session: "run1", Source: memory.SourceAgent})
	auth := map[string]string{"Authorization": "Token " + token}
	for _, q := range []string{"?agent_id=other", "?run_id=run2", "?user_id=alice&agent_id=other"} {
		if rec := do(h, "DELETE", "/v1/memories/"+res.Memory.ID+q, "", auth); rec.Code != http.StatusNotFound {
			t.Fatalf("DELETE %s: %d", q, rec.Code)
		}
	}
	if got, err := store.Get(res.Memory.ID); err != nil || got.Status != memory.StatusActive {
		t.Fatalf("memory was touched: %+v %v", got, err)
	}
	if trash, _ := store.Trash(); len(trash) != 0 {
		t.Fatalf("%d in trash", len(trash))
	}
}

func TestV1ScopedRequestCannotReachPersonalMemories(t *testing.T) {
	h, token, store := setupStore(t)
	mine, _ := store.Remember(memory.RememberInput{Content: "I deploy to Vercel"})
	auth := map[string]string{"Authorization": "Token " + token, "Content-Type": "application/json"}
	for _, c := range []struct{ method, path, body string }{
		{"GET", "/v1/memories/" + mine.Memory.ID + "?user_id=alice", ""},
		{"PUT", "/v1/memories/" + mine.Memory.ID + "?user_id=alice", `{"text":"I deploy to Fly"}`},
		{"DELETE", "/v1/memories/" + mine.Memory.ID + "?user_id=alice", ""},
	} {
		if rec := do(h, c.method, c.path, c.body, auth); rec.Code != http.StatusNotFound {
			t.Fatalf("%s %s: %d %s", c.method, c.path, rec.Code, rec.Body.String())
		}
	}
	if got, _ := store.Get(mine.Memory.ID); got.Status != memory.StatusActive || got.Content != "I deploy to Vercel" {
		t.Fatalf("personal memory changed: %+v", got)
	}
}

// F3: a v1 PUT is a new version of the same fact, so it keeps the code refs
// and type that let `woolly check` flag it later.
func TestV1PutKeepsGrounding(t *testing.T) {
	h, token, store := setupStore(t)
	res, err := store.Remember(memory.RememberInput{Content: "LED is on pin 4", Project: "bot", Type: "hardware",
		Sources: []memory.Ref{{Path: "main/pins.h"}}, Commit: "abc123"})
	if err != nil {
		t.Fatal(err)
	}
	put := v1(t, h, token, "PUT", "/v1/memories/"+res.Memory.ID, `{"text":"LED is on pin 5"}`)
	m, err := store.Get(put["id"].(string))
	if err != nil {
		t.Fatal(err)
	}
	if !m.Grounded() || m.Sources[0].Path != "main/pins.h" || m.Type != "hardware" || m.Commit != "abc123" || m.Project != "bot" {
		t.Fatalf("v1 PUT lost grounding: %+v", m)
	}
}

// Plan F1: "v1 PUT writes SourceAgent unless from review page". An app that
// rewrites a fact the user stated must not produce a new version that still
// carries the user's full trust.
func TestV1PutByAppDoesNotKeepUserTrust(t *testing.T) {
	h, token, store := setupStore(t)
	res, _ := store.Remember(memory.RememberInput{Content: "Prefers tea", Source: memory.SourceUser})
	put := v1(t, h, token, "PUT", "/v1/memories/"+res.Memory.ID, `{"text":"Prefers coffee"}`)
	m, err := store.Get(put["id"].(string))
	if err != nil {
		t.Fatal(err)
	}
	if m.Source == memory.SourceUser {
		t.Fatalf("app edit via v1 PUT saved with source %q; plan F1 says it should be %q", m.Source, memory.SourceAgent)
	}
}

func TestAPIRememberPlainTextFromAppStillSaves(t *testing.T) {
	h, token := setup(t)
	rec := do(h, "POST", "/api/remember", `{"content":"Prefers tea","source":"user"}`, map[string]string{"Authorization": "Bearer " + token})
	if rec.Code != http.StatusOK {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	var out struct {
		Memory struct {
			Source string
			Tags   []string
		} `json:"memory"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	// A script's claim that the user said it is saved a step below the user.
	if out.Memory.Source != "import" || len(out.Memory.Tags) != 1 || out.Memory.Tags[0] != "user-stated" {
		t.Fatalf("memory = %+v", out.Memory)
	}
}

func TestAppPatchCantOverruleTheUser(t *testing.T) {
	h, token, store := setupStore(t)
	app := map[string]string{"Authorization": "Bearer " + token}
	mine, _ := store.Remember(memory.RememberInput{Content: "Lives in Austin", Key: "user.location", Source: memory.SourceUser})

	// New text for a memory the user saved waits as a conflict.
	rec := do(h, "PATCH", "/api/memories/"+mine.Memory.ID, `{"content":"Lives in Denver"}`, app)
	if rec.Code != http.StatusOK {
		t.Fatalf("app PATCH of user text: %d %s", rec.Code, rec.Body.String())
	}
	var got memory.Memory
	json.Unmarshal(rec.Body.Bytes(), &got)
	if got.ID == mine.Memory.ID || got.Status != memory.StatusConflict || got.Source == memory.SourceUser {
		t.Fatalf("app edit should be a conflicting version: %+v", got)
	}
	if m, _ := store.Get(mine.Memory.ID); m.Status != memory.StatusActive || m.Content != "Lives in Austin" || m.Source != memory.SourceUser {
		t.Fatalf("user memory changed: %+v", m)
	}

	// Other changes to the user's memory are refused.
	for _, body := range []string{`{"key":"user.home"}`, `{"project":"other"}`, `{"tags":["x"]}`} {
		if rec := do(h, "PATCH", "/api/memories/"+mine.Memory.ID, body, app); rec.Code != http.StatusForbidden {
			t.Errorf("app PATCH %s: %d %s", body, rec.Code, rec.Body.String())
		}
	}

	// An app can't take over the user's key with one of its own memories.
	guess, _ := store.Remember(memory.RememberInput{Content: "Lives in Boston", Source: memory.SourceAgent})
	rec = do(h, "PATCH", "/api/memories/"+guess.Memory.ID, `{"key":"user.location"}`, app)
	if rec.Code != http.StatusOK {
		t.Fatalf("app PATCH key: %d %s", rec.Code, rec.Body.String())
	}
	if m, _ := store.Get(mine.Memory.ID); m.Status != memory.StatusActive {
		t.Fatalf("app key replaced the user's fact: %+v", m)
	}
	if m, _ := store.Get(guess.Memory.ID); m.Status != memory.StatusConflict || m.ConflictsWith != mine.Memory.ID {
		t.Fatalf("app memory should conflict with the user's: %+v", m)
	}

	// The review page is the user, and can change anything.
	review := map[string]string{"Cookie": tokenCookie + "=" + token}
	if rec := do(h, "PATCH", "/api/memories/"+mine.Memory.ID, `{"content":"Lives in Tampa","tags":["home"]}`, review); rec.Code != http.StatusOK {
		t.Fatalf("review page PATCH: %d %s", rec.Code, rec.Body.String())
	}
	if m, _ := store.Get(mine.Memory.ID); m.Content != "Lives in Tampa" || m.Source != memory.SourceUser {
		t.Fatalf("review page edit: %+v", m)
	}
}
