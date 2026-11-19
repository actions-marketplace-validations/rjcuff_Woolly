package share

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	markStart = "<!-- woolly:start -->"
	markEnd   = "<!-- woolly:end -->"
)

// export keeps a Woolly section in CLAUDE.md and AGENTS.md, when the config
// asks for it, so tools without MCP read the same notes. Text outside the
// markers is never touched.
func (r *Repo) export(d doc) error {
	section := exportSection(d)
	if d.cfg.Exports("agents") {
		if err := writeSection(filepath.Join(r.Dir, "AGENTS.md"), "# AGENTS.md", section); err != nil {
			return err
		}
	}
	claude := filepath.Join(r.Dir, "CLAUDE.md")
	if !d.cfg.Exports("claude") || (d.cfg.Exports("agents") && importsAgents(claude)) {
		return nil // off, or CLAUDE.md already pulls in AGENTS.md
	}
	return writeSection(claude, "# CLAUDE.md", section)
}

func hasSection(path string) bool {
	data, err := os.ReadFile(path)
	return err == nil && strings.Contains(string(data), markStart)
}

func importsAgents(path string) bool {
	data, err := os.ReadFile(path)
	return err == nil && strings.Contains(string(data), "@AGENTS.md")
}

// maxExport keeps Woolly's section well under 32 KiB, where Codex stops
// reading AGENTS.md, leaving room for what people write themselves.
const maxExport = 16 << 10

func exportSection(d doc) string {
	var b strings.Builder
	b.WriteString(markStart + "\n## Project memory\n\n")
	b.WriteString("Kept up to date by Woolly from .woolly/woolly.md. Edit that file, not this section.\n")
	found, full := false, false
	for _, name := range SectionNames() {
		_, facts := d.in(name)
		var items []string
		for _, ln := range facts {
			// Stale facts stay in woolly.md for review, not in files agents
			// read as true.
			if ln.stale != "" {
				continue
			}
			items = append(items, "- "+ln.text)
		}
		if len(items) == 0 || full {
			continue
		}
		part := "\n### " + name + "\n\n" + strings.Join(items, "\n") + "\n"
		if b.Len()+len(part) > maxExport {
			full = true
			continue
		}
		found = true
		b.WriteString(part)
	}
	if full {
		b.WriteString("\nMore in .woolly/woolly.md, left out here to keep this file small enough for every agent to read whole.\n")
	}
	if !found {
		b.WriteString("\nNothing yet. Woolly adds notes here as AI tools learn about this project.\n")
	}
	b.WriteString(markEnd)
	return b.String()
}

func writeSection(path, title, section string) error {
	cur, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	text := string(cur)
	start, end := strings.Index(text, markStart), strings.Index(text, markEnd)
	var out string
	switch {
	case start >= 0 && end > start:
		out = text[:start] + section + text[end+len(markEnd):]
	case strings.TrimSpace(text) == "":
		out = title + "\n\n" + section + "\n"
	default:
		out = strings.TrimRight(text, "\n") + "\n\n" + section + "\n"
	}
	if out == text {
		return nil
	}
	return os.WriteFile(path, []byte(out), 0o644)
}

func removeSection(path string) error {
	cur, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	text := string(cur)
	start, end := strings.Index(text, markStart), strings.Index(text, markEnd)
	if start < 0 || end < start {
		return nil
	}
	out := strings.TrimRight(text[:start], "\n") + "\n" + strings.TrimLeft(text[end+len(markEnd):], "\n")
	title := "# " + filepath.Base(path)
	if strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(out), title)) == "" {
		return os.Remove(path)
	}
	return os.WriteFile(path, []byte(out), 0o644)
}
