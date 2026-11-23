package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestV1IDRoutesStayInScope(t *testing.T) {
	h, token := setup(t)
	add := v1(t, h, token, "POST", "/v1/memories", `{"messages":[{"role":"user","content":"Deploys the API to Fly.io"}],"user_id":"alice","infer":false}`)
	id := results(add)[0]["id"].(string)
	auth := map[string]string{"Authorization": "Token " + token, "Content-Type": "application/json"}
	for _, c := range []struct{ method, path, body string }{
		{"GET", "/v1/memories/" + id + "?user_id=bob", ""},
		{"PUT", "/v1/memories/" + id + "?user_id=bob", `{"text":"Deploys to Vercel"}`},
		{"GET", "/v1/memories/" + id + "/history?user_id=bob", ""},
		{"DELETE", "/v1/memories/" + id + "?user_id=bob", ""},
	} {
		if rec := do(h, c.method, c.path, c.body, auth); rec.Code != http.StatusNotFound {
			t.Fatalf("%s %s: %d %s", c.method, c.path, rec.Code, rec.Body.String())
		}
	}
	if rec := do(h, "GET", "/v1/memories/"+id+"?user_id=alice", "", auth); rec.Code != http.StatusOK {
		t.Fatalf("alice: %d", rec.Code)
	}
	if rec := do(h, "GET", "/v1/memories/"+id, "", auth); rec.Code != http.StatusOK {
		t.Fatalf("no scope: %d", rec.Code)
	}
}

func TestAPIProfileAndRecallLeaveOutAppUsers(t *testing.T) {
	h, token := setup(t)
	v1(t, h, token, "POST", "/v1/memories", `{"messages":[{"role":"user","content":"Alice deploys to Fly.io"}],"user_id":"alice","infer":false}`)
	auth := map[string]string{"Authorization": "Bearer " + token}
	do(h, "POST", "/api/remember", `{"content":"I deploy to Vercel"}`, auth)
	for _, c := range []struct{ method, path, body string }{
		{"GET", "/api/profile?format=markdown", ""},
		{"POST", "/api/recall", `{"query":"deploys"}`},
	} {
		rec := do(h, c.method, c.path, c.body, auth)
		if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), "Alice") || !strings.Contains(rec.Body.String(), "Vercel") {
			t.Fatalf("%s %s: %d %s", c.method, c.path, rec.Code, rec.Body.String())
		}
	}
}

func TestV1TypeOnAddAndSearch(t *testing.T) {
	h, token := setup(t)
	v1(t, h, token, "POST", "/v1/memories", `{"messages":[{"role":"user","content":"Reset a password by emailing a link"}],"user_id":"alice","infer":false,"memory_type":"procedure"}`)
	v1(t, h, token, "POST", "/v1/memories", `{"messages":[{"role":"user","content":"Prefers password managers"}],"user_id":"alice","infer":false,"memory_type":"preference"}`)
	got := results(v1(t, h, token, "POST", "/v1/search", `{"query":"password","filters":{"user_id":"alice","type":["procedure"]}}`))
	if len(got) != 1 || got[0]["type"] != "procedure" || !strings.Contains(got[0]["memory"].(string), "Reset") {
		t.Fatalf("search: %v", got)
	}
}

func TestV1Feedback(t *testing.T) {
	h, token := setup(t)
	add := v1(t, h, token, "POST", "/v1/memories", `{"messages":[{"role":"user","content":"Deploys to Fly"}],"user_id":"alice","infer":false}`)
	id := results(add)[0]["id"].(string)
	v1(t, h, token, "POST", "/v1/memories/"+id+"/feedback?user_id=alice", `{"helpful":true}`)
	auth := map[string]string{"Authorization": "Token " + token}
	if rec := do(h, "POST", "/v1/memories/"+id+"/feedback?user_id=bob", `{"helpful":true}`, auth); rec.Code != http.StatusNotFound {
		t.Fatalf("other user: %d", rec.Code)
	}
	if rec := do(h, "POST", "/v1/memories/nope/feedback", `{"helpful":true}`, auth); rec.Code != http.StatusNotFound {
		t.Fatalf("missing: %d", rec.Code)
	}
	if rec := do(h, "POST", "/v1/memories/"+id+"/feedback", `{}`, auth); rec.Code != http.StatusBadRequest {
		t.Fatalf("no helpful: %d", rec.Code)
	}
}

func TestV1ConflictInbox(t *testing.T) {
	h, token := setup(t)
	auth := map[string]string{"Authorization": "Token " + token}
	cookie := map[string]string{"Cookie": tokenCookie + "=" + token}
	// Alice said it herself (through the review page); an agent guesses otherwise.
	do(h, "POST", "/api/remember", `{"content":"Alice lives in Austin","key":"loc","source":"user"}`, cookie)
	rec := do(h, "POST", "/api/remember", `{"content":"Alice lives in Denver","key":"loc","source":"agent"}`, auth)
	var saved struct {
		Memory struct{ ID string } `json:"memory"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &saved)
	out := v1(t, h, token, "GET", "/v1/conflicts", "")
	list := out["results"].([]any)
	if len(list) != 1 || list[0].(map[string]any)["conflicts_with"] == nil {
		t.Fatalf("conflicts: %v", out)
	}
	if rec := do(h, "POST", "/v1/conflicts/"+saved.Memory.ID+"/resolve", `{}`, auth); rec.Code != http.StatusForbidden {
		t.Fatalf("app resolve: %d %s", rec.Code, rec.Body.String())
	}
	if rec := do(h, "POST", "/v1/conflicts/"+saved.Memory.ID+"/resolve", `{}`, cookie); rec.Code != http.StatusOK {
		t.Fatalf("review page resolve: %d %s", rec.Code, rec.Body.String())
	}
	if got := v1(t, h, token, "GET", "/v1/conflicts?user_id=bob", "")["results"].([]any); len(got) != 0 {
		t.Fatalf("bob sees %v", got)
	}
}
