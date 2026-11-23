package memory

import (
	"os"
	"path/filepath"
	"strings"
)

// LinksEnv picks how the links: line is written. "wikilinks" writes each ID
// as a quoted [[id]] in a YAML list, so Obsidian shows the links. "plain"
// writes bare IDs. Unset, Woolly writes wikilinks when the memories folder
// is inside an Obsidian vault and plain IDs otherwise. Both styles are
// always read.
const LinksEnv = "WOOLLY_LINKS"

// encode writes m the way this store writes links.
func (s *Store) encode(m *Memory) []byte { return encodeAs(m, s.wikilinks()) }

func (s *Store) wikilinks() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(LinksEnv))) {
	case "wikilinks":
		return true
	case "plain":
		return false
	}
	return inObsidianVault(s.dir)
}

// inObsidianVault reports whether dir or one of the two folders above it is
// an Obsidian vault, which has a .obsidian folder at its top. That covers
// opening ~/.woolly/memories or ~/.woolly as a vault, or keeping
// WOOLLY_HOME in a vault folder.
func inObsidianVault(dir string) bool {
	for i := 0; i < 3; i++ {
		if dirExists(filepath.Join(dir, ".obsidian")) {
			return true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return false
}

func dirExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

// formatLinks writes linked IDs for the links: line. Wikilinks are written
// as a YAML list of quoted strings, the only form Obsidian reads as links in
// front matter.
func formatLinks(ids []string, wiki bool) string {
	if len(ids) == 0 {
		return ""
	}
	if !wiki {
		return strings.Join(ids, ", ")
	}
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = `"[[` + id + `]]"`
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// parseLinks reads a links: line in any style Woolly writes, or that a
// person or Obsidian might: a, b / [[a]], [[b]] / ["[[a]]", "[[b]]"] /
// [a, b]. Only valid IDs are kept.
func parseLinks(v string) []string {
	v = strings.TrimSpace(v)
	if strings.HasPrefix(v, "[") && !strings.HasPrefix(v, "[[") && strings.HasSuffix(v, "]") {
		v = v[1 : len(v)-1] // a YAML flow list
	}
	var ids []string
	for _, id := range strings.Split(v, ",") {
		id = strings.Trim(strings.TrimSpace(id), `"'`)
		id = strings.TrimSuffix(strings.TrimPrefix(id, "[["), "]]")
		if id = strings.TrimSpace(id); ValidID(id) {
			ids = append(ids, id)
		}
	}
	return ids
}
