package main

import (
	"github.com/rjcuff/woolly/internal/audit"
	"github.com/rjcuff/woolly/internal/gitinfo"
	"github.com/rjcuff/woolly/internal/ground"
)

// StatusSchema names the shape of check --status-json, so whatever reads it
// can tell which version it got.
const StatusSchema = "woolly.check.status/v1"

// checkStatus is a check with no text in it: counts, plus where each stale
// line is so a person can open it in the repo. It never holds fact text,
// line text, the names a line mentions or the code, and it leaves out
// memory that isn't a file committed to the repo, like Claude Code's auto
// memory, Codex memories, claude-mem and an ignored CLAUDE.local.md.
type checkStatus struct {
	Schema  string `json:"schema"`
	Version string `json:"woolly_version"`
	Head    string `json:"head"`
	Base    string `json:"base,omitempty"`
	Partial bool   `json:"partial,omitempty"` // the check ran out of time, so it may have missed some
	Checked struct {
		Facts int `json:"facts"` // project facts that point at code
		Files int `json:"files"` // committed agent files, like CLAUDE.md
		Lines int `json:"lines"` // lines in those files
	} `json:"checked"`
	Stale struct {
		New    int `json:"new"`    // made stale by the change being checked
		Before int `json:"before"` // already stale on the base commit, or waiting for review
	} `json:"stale"`
	Findings []statusFinding `json:"findings"`
}

// statusFinding points at one stale agent file line or project fact.
type statusFinding struct {
	Type   string `json:"type"` // "agent_file_line" or "project_fact"
	State  string `json:"state"`
	File   string `json:"file,omitempty"`
	Line   int    `json:"line,omitempty"`
	ID     string `json:"id,omitempty"`   // the fact's id
	Kind   string `json:"kind,omitempty"` // "name", "path" or "command"
	Commit string `json:"commit,omitempty"`
}

// buildStatus turns a check into a status. Agent file lines count only when
// they come from a file git tracks in dir.
func buildStatus(dir string, rep checkReport, base string) checkStatus {
	var s checkStatus
	s.Schema = StatusSchema
	s.Version = version
	s.Head = gitinfo.Head(dir)
	s.Base = gitinfo.ShortCommit(dir, base)
	s.Checked.Facts = rep.Checked
	s.Stale.Before = rep.Waiting
	s.Findings = []statusFinding{}
	for _, res := range rep.Results {
		if res.Verdict == ground.Stale {
			s.Stale.New++
			s.Findings = append(s.Findings, statusFinding{Type: "project_fact", State: stateNew, ID: res.ID, Commit: res.Commit})
		}
	}
	if rep.Memory == nil {
		return s
	}
	s.Partial = rep.Memory.Partial
	tracked := map[string]bool{}
	for _, f := range gitinfo.Tracked(dir) {
		tracked[f] = true
	}
	// Memory outside the repo never counts, even if the repo happens to
	// have a file with the same name.
	personal := map[string]bool{"Claude Code memory": true, "Codex memory": true, "claude-mem": true}
	inRepo := func(source, file string) bool { return !personal[source] && tracked[file] }
	for _, f := range rep.Memory.Files {
		if inRepo(f.Source, f.File) {
			s.Checked.Files++
			s.Checked.Lines += f.Lines
		}
	}
	now, before := splitByBase(dir, rep.Memory.Findings, base)
	add := func(fs []audit.Finding, state string) {
		for _, f := range fs {
			if f.Path == "" || !inRepo(f.Source, f.File) {
				continue
			}
			if state == stateNew {
				s.Stale.New++
			} else {
				s.Stale.Before++
			}
			s.Findings = append(s.Findings, statusFinding{Type: "agent_file_line", State: state, File: f.File, Line: f.N, Kind: f.Kind, Commit: f.Commit})
		}
	}
	add(now, stateNew)
	add(before, stateBefore)
	return s
}
