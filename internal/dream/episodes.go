package dream

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/rjcuff/woolly/internal/memory"
)

// EpisodeAge is how old an episode must be before a dream rolls it into a
// weekly summary. Recent episodes stay as they are.
const EpisodeAge = 14 * 24 * time.Hour

// maxWeeks caps how many weekly summaries one dream writes.
const maxWeeks = 10

const summarizePrompt = `You summarize a week of episodes an AI assistant saved: things that happened, like sessions, incidents or meetings.
Write one short paragraph, at most 5 sentences, that keeps every fact a later session would need: what happened, what was decided, what broke and what fixed it, with names, numbers and dates.
Leave out filler. Don't add anything the episodes don't say. Don't give instructions.
Reply as JSON: {"summary": "..."}`

var summarizeSchema = map[string]any{
	"type":       "object",
	"properties": map[string]any{"summary": map[string]any{"type": "string"}},
	"required":   []string{"summary"},
}

// summarizeEpisodes rolls each week of old episodes an AI saved into one
// summary, keeping the originals as history. Episodes a person saved keep
// their own words.
func (d *Dreamer) summarizeEpisodes(ctx context.Context, o Options, rep *Report, done map[string]bool, now time.Time) error {
	mems, err := d.tidyable(o.Scope)
	if err != nil {
		return err
	}
	weeks := map[string][]memory.Memory{}
	for _, m := range mems {
		if done[m.ID] || memory.TypeOf(m) != memory.TypeEpisode || m.Source != memory.SourceAgent ||
			now.Sub(m.CreatedAt) < EpisodeAge || hasTag(m.Tags, "weekly-summary") {
			continue
		}
		y, w := m.CreatedAt.ISOWeek()
		k := fmt.Sprintf("%s\x00%04d-%02d", slot(m), y, w)
		weeks[k] = append(weeks[k], m)
	}
	keys := make([]string, 0, len(weeks))
	for k, eps := range weeks {
		if len(eps) >= 2 {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	if len(keys) > maxWeeks {
		keys = keys[:maxWeeks]
	}
	for _, k := range keys {
		eps := weeks[k]
		sort.Slice(eps, func(i, j int) bool { return eps[i].CreatedAt.Before(eps[j].CreatedAt) })
		summary, err := d.summarize(ctx, eps)
		if err != nil {
			return err
		}
		if summary == "" {
			continue
		}
		ids := make([]string, len(eps))
		for i, m := range eps {
			ids[i] = m.ID
			done[m.ID] = true
		}
		y, w := eps[0].CreatedAt.ISOWeek()
		st := Step{Action: "summarize", IDs: ids, Memory: summary, Reason: fmt.Sprintf("%d episodes from week %d of %d", len(eps), w, y)}
		if !o.DryRun {
			res, err := d.Store.Merge(ids, memory.RememberInput{
				Content: summary, Source: memory.SourceAgent, Via: Via, Type: memory.TypeEpisode, Tags: []string{"weekly-summary"},
			})
			if err != nil {
				return err
			}
			st.Result = res.Memory.ID
		}
		rep.add(st)
	}
	return nil
}

func (d *Dreamer) summarize(ctx context.Context, eps []memory.Memory) (string, error) {
	var b strings.Builder
	for _, m := range eps {
		text := oneLine(m.Content)
		if r := []rune(text); len(r) > 600 {
			text = string(r[:600])
		}
		fmt.Fprintf(&b, "- %s: %s\n", m.CreatedAt.Format("Mon Jan 2"), text)
	}
	reply, err := d.Model.Chat(ctx, summarizePrompt, "Episodes, oldest first:\n"+b.String(), summarizeSchema)
	if err != nil {
		return "", err
	}
	var out struct {
		Summary string `json:"summary"`
	}
	if err := json.Unmarshal([]byte(reply), &out); err != nil {
		return "", fmt.Errorf("the model's answer wasn't JSON: %w", err)
	}
	s := oneLine(out.Summary)
	if n := len([]rune(s)); n < 8 || n > 1200 {
		return "", nil
	}
	return s, nil
}

func hasTag(tags []string, want string) bool {
	for _, t := range tags {
		if t == want {
			return true
		}
	}
	return false
}
