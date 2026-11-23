package connect

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The Claude Code plugin in plugin/ runs the same hooks that `woolly connect`
// adds. This keeps the two from drifting apart.
func TestPluginHooksMatchConnect(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "plugin", "hooks", "hooks.json"))
	if err != nil {
		t.Fatal(err)
	}
	type hook struct {
		Type    string `json:"type"`
		Command string `json:"command"`
		Timeout int    `json:"timeout"`
	}
	type group struct {
		Matcher string `json:"matcher"`
		Hooks   []hook `json:"hooks"`
	}
	var file struct {
		Hooks map[string][]group `json:"hooks"`
	}
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("hooks.json: %v", err)
	}

	count := 0
	for _, groups := range file.Hooks {
		for _, g := range groups {
			count += len(g.Hooks)
		}
	}
	if count != len(hookSpecs) {
		t.Errorf("plugin has %d hooks, connect adds %d", count, len(hookSpecs))
	}

	for _, s := range hookSpecs {
		groups := file.Hooks[s.event]
		if len(groups) != 1 || len(groups[0].Hooks) != 1 {
			t.Errorf("%s: want one group with one hook, got %+v", s.event, groups)
			continue
		}
		g, h := groups[0], groups[0].Hooks[0]
		if g.Matcher != s.matcher {
			t.Errorf("%s: matcher %q, connect uses %q", s.event, g.Matcher, s.matcher)
		}
		if h.Type != "command" {
			t.Errorf("%s: type %q, want command", s.event, h.Type)
		}
		if h.Timeout != s.timeout {
			t.Errorf("%s: timeout %d, connect uses %d", s.event, h.Timeout, s.timeout)
		}
		if !strings.Contains(h.Command, "then woolly "+s.args+";") {
			t.Errorf("%s: command %q should run %q", s.event, h.Command, "woolly "+s.args)
		}
		if !strings.HasPrefix(h.Command, "if command -v woolly ") {
			t.Errorf("%s: command %q should check woolly is installed first", s.event, h.Command)
		}
	}
}

func TestPluginManifestRunsMCP(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "plugin", ".claude-plugin", "plugin.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m struct {
		Name       string `json:"name"`
		Version    string `json:"version"`
		MCPServers map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("plugin.json: %v", err)
	}
	if m.Name != "woolly" || m.Version == "" {
		t.Errorf("name %q version %q", m.Name, m.Version)
	}
	s, ok := m.MCPServers["woolly"]
	if !ok || s.Command != "woolly" || len(s.Args) != 1 || s.Args[0] != "mcp" {
		t.Errorf("mcpServers.woolly = %+v, want woolly mcp", s)
	}
}

func TestPluginInstalled(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", "")
	if PluginInstalled(home) {
		t.Fatal("no plugins file should mean not installed")
	}
	dir := filepath.Join(home, ".claude", "plugins")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, "installed_plugins.json"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(`{"version":2,"plugins":{"woollyish@x":[],"other@woolly":[]}}`)
	if PluginInstalled(home) {
		t.Error("only an exact plugin name should count")
	}
	write(`{"version":2,"plugins":{"woolly@woolly":[{"scope":"user"}]}}`)
	if !PluginInstalled(home) {
		t.Error("woolly@woolly should count as installed")
	}
	write(`not json`)
	if PluginInstalled(home) {
		t.Error("an unreadable file should not count")
	}
}
