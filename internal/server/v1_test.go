package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func v1(t *testing.T, h http.Handler, token, method, path, body string) map[string]any {
	t.Helper()
	rec := do(h, method, path, body, map[string]string{"Authorization": "Token " + token, "Content-Type": "application/json"})
	if rec.Code != http.StatusOK {
		t.Fatalf("%s %s: %d %s", method, path, rec.Code, rec.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func results(out map[string]any) []map[string]any {
	var list []map[string]any
	for _, r := range out["results"].([]any) {
		list = append(list, r.(map[string]any))
	}
	return list
}

func TestV1NeedsTheToken(t *testing.T) {
	h, token := setup(t)
	if code := do(h, "GET", "/v1/memories", "", nil).Code; code != http.StatusUnauthorized {
		t.Fatalf("no token: %d", code)
	}
	for _, auth := range []string{"Bearer " + token, "Token " + token} {
		if code := do(h, "GET", "/v1/memories", "", map[string]string{"Authorization": auth}).Code; code != http.StatusOK {
			t.Fatalf("%q: %d", auth, code)
		}
	}
}

func TestV1AddListSearchUpdateHistoryDelete(t *testing.T) {
	h, token := setup(t)

	// Without a model, infer: false saves user messages as they are.
	add := v1(t, h, token, "POST", "/v1/memories", `{"messages":[{"role":"user","content":"Deploys the API to Fly.io"},{"role":"assistant","content":"ok"}],"user_id":"alice","infer":false,"metadata":{"team":"web"}}`)
	got := results(add)
	if len(got) != 1 || got[0]["event"] != "ADD" || got[0]["memory"] != "Deploys the API to Fly.io" {
		t.Fatalf("add: %v", got)
	}
	id := got[0]["id"].(string)
	v1(t, h, token, "POST", "/v1/memories", `{"messages":[{"role":"user","content":"Deploys the API to Render"}],"user_id":"bob","infer":false}`)

	// Lists and searches stay inside the user.
	list := results(v1(t, h, token, "GET", "/v1/memories?user_id=alice", ""))
	if len(list) != 1 || list[0]["user_id"] != "alice" || list[0]["metadata"].(map[string]any)["team"] != "web" {
		t.Fatalf("alice's list: %v", list)
	}
	found := results(v1(t, h, token, "POST", "/v1/search", `{"query":"where does the API deploy","filters":{"user_id":"bob"}}`))
	if len(found) != 1 || !strings.Contains(found[0]["memory"].(string), "Render") || found[0]["score"] == nil {
		t.Fatalf("bob's search: %v", found)
	}

	// An update is a new version, and the history shows both.
	up := v1(t, h, token, "PUT", "/v1/memories/"+id, `{"text":"Deploys the API to Fly.io from main"}`)
	newID := up["id"].(string)
	if newID == id || up["memory"] != "Deploys the API to Fly.io from main" || up["user_id"] != "alice" {
		t.Fatalf("update: %v", up)
	}
	hist := results(v1(t, h, token, "GET", "/v1/memories/"+newID+"/history", ""))
	if len(hist) != 2 || hist[0]["event"] != "ADD" || hist[1]["event"] != "UPDATE" || hist[1]["old_memory"] != "Deploys the API to Fly.io" {
		t.Fatalf("history: %v", hist)
	}
	old := v1(t, h, token, "GET", "/v1/memories/"+id, "")
	if old["status"] != "superseded" || old["invalid_at"] == nil {
		t.Fatalf("old version: %v", old)
	}

	// Deleting by scope needs a scope, and leaves other users alone.
	if code := do(h, "DELETE", "/v1/memories", "", map[string]string{"Authorization": "Bearer " + token}).Code; code != http.StatusBadRequest {
		t.Fatalf("delete everything should need a scope: %d", code)
	}
	v1(t, h, token, "DELETE", "/v1/memories?user_id=alice", "")
	if left := results(v1(t, h, token, "GET", "/v1/memories?user_id=alice", "")); len(left) != 0 {
		t.Fatalf("alice should be empty: %v", left)
	}
	if left := results(v1(t, h, token, "GET", "/v1/memories?user_id=bob", "")); len(left) != 1 {
		t.Fatalf("bob should be untouched: %v", left)
	}
}

func TestV1InferWithoutAModel(t *testing.T) {
	h, token := setup(t)
	rec := do(h, "POST", "/v1/memories", `{"messages":[{"role":"user","content":"I live in Denver"}],"user_id":"alice"}`,
		map[string]string{"Authorization": "Bearer " + token})
	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), `\"infer\": false`) {
		t.Fatalf("want 503 with a hint: %d %s", rec.Code, rec.Body.String())
	}
}

func TestV1DreamFoldsRepeats(t *testing.T) {
	h, token := setup(t)
	for _, text := range []string{"Uses pnpm", "uses pnpm."} {
		v1(t, h, token, "POST", "/v1/memories", `{"messages":[{"role":"user","content":"`+text+`"}],"user_id":"alice","infer":false}`)
	}
	dry := v1(t, h, token, "POST", "/v1/dream", `{"user_id":"alice","dry_run":true}`)
	if dry["merged"].(float64) != 1 || len(results(v1(t, h, token, "GET", "/v1/memories?user_id=alice", ""))) != 2 {
		t.Fatalf("dry run: %v", dry)
	}
	rep := v1(t, h, token, "POST", "/v1/dream", `{"user_id":"alice"}`)
	if rep["merged"].(float64) != 1 || len(results(v1(t, h, token, "GET", "/v1/memories?user_id=alice", ""))) != 1 {
		t.Fatalf("dream: %v", rep)
	}
}
