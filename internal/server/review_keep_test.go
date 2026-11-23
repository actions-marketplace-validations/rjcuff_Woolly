package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/rjcuff/woolly/internal/memory"
)

func TestKeepTakesAGuessOffTheReviewQueue(t *testing.T) {
	h, token, store := setupStore(t)
	auth := map[string]string{"Authorization": "Bearer " + token}
	res, err := store.Remember(memory.RememberInput{Content: "Probably deploys on Fridays", Source: memory.SourceAgent})
	if err != nil {
		t.Fatal(err)
	}
	id := res.Memory.ID

	if rec := do(h, "POST", "/api/review-queue/"+id+"/keep", "", nil); rec.Code != http.StatusUnauthorized {
		t.Fatalf("keep without a token: %d", rec.Code)
	}
	if rec := do(h, "POST", "/api/review-queue/no-such-id/keep", "", auth); rec.Code != http.StatusNotFound {
		t.Fatalf("keep unknown id: %d %s", rec.Code, rec.Body.String())
	}
	if rec := do(h, "POST", "/api/review-queue/"+id+"/keep", "", auth); rec.Code != http.StatusOK {
		t.Fatalf("keep: %d %s", rec.Code, rec.Body.String())
	}
	if last := store.Usage()[id].Last; time.Since(last) > time.Minute {
		t.Fatalf("keep should count as a use now, last = %v", last)
	}
}
