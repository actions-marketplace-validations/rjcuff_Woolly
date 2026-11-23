package recall

import (
	"context"
	"testing"
	"time"

	"github.com/rjcuff/woolly/internal/memory"
)

// The eval is a fixed set of memories and questions with the memory each
// question should find. It runs in keyword mode, with no Ollama, so it is
// fast and the same everywhere. TestRecallEval fails if recall@5 drops below
// the floor; raise the floor when ranking gets better.

var evalMemories = []struct{ key, content string }{
	{"loc", "Lives in Denver, Colorado"},
	{"job", "Works as a backend engineer at a fintech startup"},
	{"lang", "Prefers Go for services and TypeScript for web code"},
	{"editor", "Uses Neovim with the LazyVim setup"},
	{"indent", "Indents Python with four spaces, never tabs"},
	{"answers", "Wants short answers without long explanations"},
	{"coffee", "Drinks oat milk flat whites, no sugar"},
	{"diet", "Is vegetarian and avoids fish"},
	{"pet", "Has a border collie named Pixel"},
	{"kids", "Has two kids, ages 6 and 9"},
	{"tz", "Works in Mountain Time and avoids meetings before 9am"},
	{"os", "Uses a MacBook Pro with macOS and a Linux desktop at home"},
	{"db", "The billing service uses Postgres 16 with pgbouncer"},
	{"deploy", "Deploys the API to Fly.io from GitHub Actions"},
	{"tests", "Runs tests with make check before every commit"},
	{"lint", "Uses golangci-lint with the default rules plus gofumpt"},
	{"commits", "Writes commit messages in conventional commits format"},
	{"review", "Wants code review comments ranked by severity"},
	{"cloud", "Avoids AWS; prefers Fly.io and Cloudflare"},
	{"auth", "The web app signs users in with passkeys through WebAuthn"},
	{"queue", "Background jobs run on a Redis queue with asynq"},
	{"logs", "Logs go to Grafana Loki and alerts to PagerDuty"},
	{"run", "Runs three times a week and is training for a half marathon"},
	{"music", "Listens to lo-fi hip hop while coding"},
	{"books", "Reads science fiction, favorite author is Ursula K. Le Guin"},
	{"travel", "Is planning a trip to Japan in April"},
	{"allergy", "Is allergic to peanuts"},
	{"car", "Drives a 2019 Subaru Outback"},
	{"bday", "Birthday is March 14"},
	{"spanish", "Is learning Spanish on Duolingo"},
	{"keyboard", "Types on a split ergonomic keyboard with a Colemak layout"},
	{"theme", "Likes dark themes and the Catppuccin color scheme"},
	{"docs", "Writes docs in plain American English without em dashes"},
	{"api", "Designs REST APIs with cursor pagination and problem+json errors"},
	{"cache", "Caches API responses in Redis for five minutes"},
	{"mobile", "Builds the mobile app with React Native and Expo"},
	{"ci", "CI runs on GitHub Actions with a matrix for Linux and macOS"},
	{"secrets", "Keeps secrets in 1Password and loads them with op run"},
	{"monorepo", "The monorepo uses pnpm workspaces and Turborepo"},
	{"style", "Prefers small functions and early returns over nesting"},
}

var evalQueries = []struct{ query, want string }{
	{"where does the user live", "loc"},
	{"what city", "loc"},
	{"what is their job", "job"},
	{"which programming language for services", "lang"},
	{"what editor do they use", "editor"},
	{"tabs or spaces in python", "indent"},
	{"how long should answers be", "answers"},
	{"coffee order", "coffee"},
	{"food preferences vegetarian", "diet"},
	{"dog name", "pet"},
	{"children", "kids"},
	{"time zone meetings", "tz"},
	{"which database does billing use", "db"},
	{"how is the api deployed", "deploy"},
	{"how to run the tests", "tests"},
	{"linter", "lint"},
	{"commit message format", "commits"},
	{"cloud provider", "cloud"},
	{"how do users log in", "auth"},
	{"background jobs", "queue"},
	{"where do logs go", "logs"},
	{"exercise running", "run"},
	{"favorite author", "books"},
	{"upcoming trip", "travel"},
	{"allergies", "allergy"},
	{"what car", "car"},
	{"birthday", "bday"},
	{"keyboard layout", "keyboard"},
	{"color scheme", "theme"},
	{"pagination style", "api"},
	{"mobile app framework", "mobile"},
	{"where are secrets stored", "secrets"},
	{"package manager monorepo", "monorepo"},
	{"code style nesting", "style"},
}

// evalFloor is the recall@5 the suite must keep. Raise it as ranking improves.
const evalFloor = 0.75

func TestRecallEval(t *testing.T) {
	store, err := memory.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]string{}
	for _, m := range evalMemories {
		res, err := store.Remember(memory.RememberInput{Content: m.content})
		if err != nil {
			t.Fatal(err)
		}
		ids[res.Memory.ID] = m.key
	}
	e := New(store, nil, "")
	start := time.Now()
	hits := 0
	var misses []string
	for _, q := range evalQueries {
		results, _, err := e.Recall(context.Background(), q.query, 5, false)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, r := range results {
			if ids[r.Memory.ID] == q.want {
				found = true
			}
		}
		if found {
			hits++
		} else {
			misses = append(misses, q.query)
		}
	}
	score := float64(hits) / float64(len(evalQueries))
	t.Logf("recall@5 = %.2f (%d/%d) in %v; missed: %v", score, hits, len(evalQueries), time.Since(start), misses)
	if score < evalFloor {
		t.Fatalf("recall@5 %.2f is below the floor %.2f", score, evalFloor)
	}
}

func TestRecallEvalEmptyStore(t *testing.T) {
	store, err := memory.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	results, _, err := New(store, nil, "").Recall(context.Background(), "anything", 5, false)
	if err != nil || len(results) != 0 {
		t.Fatalf("results %v, err %v", results, err)
	}
}
