package server

// QA tests for S5 (/v1/conflicts), S1 links, G1 type filters and G2
// feedback over the local API, written against plans/memory-growth-and-fix.md.

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/rjcuff/woolly/internal/memory"
)

func appAuth(token string) map[string]string {
	return map[string]string{"Authorization": "Token " + token, "Content-Type": "application/json"}
}

func TestQAConflictInboxAgentAndRunScopes(t *testing.T) {
	h, token, store := setupStore(t)
	store.Remember(memory.RememberInput{Content: "Bot replies in English", Key: "lang", Agent: "bot", Source: memory.SourceImport})
	store.Remember(memory.RememberInput{Content: "Bot replies in French", Key: "lang", Agent: "bot", Source: memory.SourceAgent})
	store.Remember(memory.RememberInput{Content: "Run uses GPU", Key: "hw", Session: "r1", Source: memory.SourceImport})
	store.Remember(memory.RememberInput{Content: "Run uses CPU", Key: "hw", Session: "r1", Source: memory.SourceAgent})
	count := func(q string) int {
		return len(v1(t, h, token, "GET", "/v1/conflicts"+q, "")["results"].([]any))
	}
	if n := count("?agent_id=bot"); n != 1 {
		t.Fatalf("agent bot: %d", n)
	}
	if n := count("?run_id=r1"); n != 1 {
		t.Fatalf("run r1: %d", n)
	}
	if n := count("?agent_id=other"); n != 0 {
		t.Fatalf("agent other: %d", n)
	}
	if n := count("?agent_id=bot&run_id=r1"); n != 0 {
		t.Fatalf("agent bot in run r1: %d", n)
	}
	// Resolve through the wrong scope is not found.
	conf, _ := store.Conflicts()
	for _, c := range conf {
		if c.Agent == "bot" {
			rec := do(h, "POST", "/v1/conflicts/"+c.ID+"/resolve?agent_id=other", `{}`, appAuth(token))
			if rec.Code != http.StatusNotFound {
				t.Fatalf("resolve out of scope: %d", rec.Code)
			}
			// The app can't pick the less trusted version in its own scope either.
			rec = do(h, "POST", "/v1/conflicts/"+c.ID+"/resolve?agent_id=bot", `{}`, appAuth(token))
			if rec.Code != http.StatusForbidden {
				t.Fatalf("app resolve over import: %d %s", rec.Code, rec.Body.String())
			}
		}
	}
}

// The conflict inbox resolve must not let an app bring back a fact the user
// replaced or retired. Neither of these is a conflict.
func TestQAAppCannotReviveReplacedOrRetiredFact(t *testing.T) {
	h, token, store := setupStore(t)
	old, _ := store.Remember(memory.RememberInput{Content: "Lives in Austin", Key: "loc"})
	now, _ := store.Remember(memory.RememberInput{Content: "Lives in Denver", Key: "loc"})
	gone, _ := store.Remember(memory.RememberInput{Content: "Works at Acme"})
	store.Expire(gone.Memory.ID)

	rec := do(h, "POST", "/v1/conflicts/"+old.Memory.ID+"/resolve", `{}`, appAuth(token))
	if rec.Code == http.StatusOK {
		cur, _ := store.Get(now.Memory.ID)
		t.Errorf("an app revived the user's replaced fact; the user's newer fact is now %s", cur.Status)
	}
	rec = do(h, "POST", "/v1/conflicts/"+gone.Memory.ID+"/resolve", `{}`, appAuth(token))
	if rec.Code == http.StatusOK {
		g, _ := store.Get(gone.Memory.ID)
		t.Errorf("an app un-retired a fact the user expired; status is now %s", g.Status)
	}
}

// Same for the store directly, which MCP resolve_conflict uses.
func TestQAResolveAsAgentOnNonConflicts(t *testing.T) {
	_, _, store := setupStore(t)
	old, _ := store.Remember(memory.RememberInput{Content: "Lives in Austin", Key: "loc"})
	store.Remember(memory.RememberInput{Content: "Lives in Denver", Key: "loc"})
	if _, _, err := store.ResolveAs(old.Memory.ID, memory.SourceAgent); err == nil {
		t.Error("ResolveAs(agent) rolled back the user's own newer fact")
	}
}

func TestQALinksEndpointScopes(t *testing.T) {
	h, token, store := setupStore(t)
	a, _ := store.Remember(memory.RememberInput{Content: "Alice resets by email", User: "alice"})
	b, _ := store.Remember(memory.RememberInput{Content: "Alice likes SMS", User: "alice"})
	bob, _ := store.Remember(memory.RememberInput{Content: "Bob likes push", User: "bob"})
	mine, _ := store.Remember(memory.RememberInput{Content: "I use 1Password"})

	if rec := do(h, "POST", "/v1/memories/"+a.Memory.ID+"/links?user_id=alice", `{"to":"`+bob.Memory.ID+`"}`, appAuth(token)); rec.Code != http.StatusNotFound {
		t.Fatalf("link to bob from alice's scope: %d", rec.Code)
	}
	if rec := do(h, "POST", "/v1/memories/"+a.Memory.ID+"/links", `{"to":"`+mine.Memory.ID+`"}`, appAuth(token)); rec.Code == http.StatusOK {
		t.Fatal("linked an app user's memory to the user's own")
	}
	if rec := do(h, "POST", "/v1/memories/"+a.Memory.ID+"/links?user_id=alice", `{"to":"`+a.Memory.ID+`"}`, appAuth(token)); rec.Code != http.StatusBadRequest {
		t.Fatalf("self link: %d", rec.Code)
	}
	if rec := do(h, "POST", "/v1/memories/"+a.Memory.ID+"/links?user_id=alice", `{}`, appAuth(token)); rec.Code == http.StatusOK {
		t.Fatal("link with no target")
	}
	out := v1(t, h, token, "POST", "/v1/memories/"+a.Memory.ID+"/links?user_id=alice", `{"to":"`+b.Memory.ID+`"}`)
	if links, _ := out["links"].([]any); len(links) != 1 {
		t.Fatalf("links = %v", out)
	}
	got := v1(t, h, token, "GET", "/v1/memories/"+b.Memory.ID+"?user_id=alice", "")
	if links, _ := got["links"].([]any); len(links) != 1 {
		t.Fatalf("other side: %v", got)
	}
}

func TestQAFeedbackScopesByAgent(t *testing.T) {
	h, token, store := setupStore(t)
	m, _ := store.Remember(memory.RememberInput{Content: "Bot answers in haiku", Agent: "bot"})
	if rec := do(h, "POST", "/v1/memories/"+m.Memory.ID+"/feedback?agent_id=other", `{"helpful":true}`, appAuth(token)); rec.Code != http.StatusNotFound {
		t.Fatalf("other agent: %d", rec.Code)
	}
	if rec := do(h, "POST", "/v1/memories/"+m.Memory.ID+"/feedback?agent_id=bot", `{"helpful":false}`, appAuth(token)); rec.Code != http.StatusOK {
		t.Fatalf("own agent: %d", rec.Code)
	}
	if rec := do(h, "POST", "/v1/memories/"+m.Memory.ID+"/feedback?agent_id=bot", `{"helpful":"yes"}`, appAuth(token)); rec.Code != http.StatusBadRequest {
		t.Fatalf("string helpful: %d", rec.Code)
	}
	if u := store.Usage()[m.Memory.ID]; u.Down != 1 || u.Up != 0 {
		t.Fatalf("usage = %+v", u)
	}
}

func TestQASearchUnknownTypeFilter(t *testing.T) {
	h, token, store := setupStore(t)
	store.Remember(memory.RememberInput{Content: "Reset password by emailing a link", User: "alice", Type: "procedure"})
	rec := do(h, "POST", "/v1/search", `{"query":"password","filters":{"user_id":"alice","type":["fact"]}}`, appAuth(token))
	if rec.Code != http.StatusOK {
		return // refusing an unknown type is fine
	}
	var out map[string]any
	json.Unmarshal(rec.Body.Bytes(), &out)
	if len(results(out)) != 0 {
		t.Fatalf(`filters.type ["fact"] matched a procedure: %v`, out)
	}
}

func TestQASearchTouchesOnlyReturned(t *testing.T) {
	h, token, store := setupStore(t)
	a, _ := store.Remember(memory.RememberInput{Content: "Uses Postgres for orders", User: "alice"})
	b, _ := store.Remember(memory.RememberInput{Content: "Uses Postgres for users", User: "bob"})
	v1(t, h, token, "POST", "/v1/search", `{"query":"postgres","filters":{"user_id":"alice"}}`)
	u := store.Usage()
	if u[a.Memory.ID].Uses != 1 || u[b.Memory.ID].Uses != 0 {
		t.Fatalf("usage = %+v", u)
	}
}

func TestQAReviewQueueAndProblemsEndpoints(t *testing.T) {
	h, token, _ := setupStore(t)
	if rec := do(h, "GET", "/api/review-queue", "", nil); rec.Code != http.StatusUnauthorized {
		t.Fatalf("review queue without a token: %d", rec.Code)
	}
	if rec := do(h, "GET", "/api/problems", "", nil); rec.Code != http.StatusUnauthorized {
		t.Fatalf("problems without a token: %d", rec.Code)
	}
	auth := map[string]string{"Authorization": "Bearer " + token}
	rec := do(h, "GET", "/api/review-queue", "", auth)
	var q struct {
		Memories []any `json:"memories"`
	}
	if rec.Code != http.StatusOK || json.Unmarshal(rec.Body.Bytes(), &q) != nil || q.Memories == nil {
		t.Fatalf("review queue: %d %s", rec.Code, rec.Body.String())
	}
	rec = do(h, "GET", "/api/problems", "", auth)
	var p struct {
		Problems []string `json:"problems"`
	}
	if rec.Code != http.StatusOK || json.Unmarshal(rec.Body.Bytes(), &p) != nil || p.Problems == nil {
		t.Fatalf("problems: %d %s", rec.Code, rec.Body.String())
	}
}
