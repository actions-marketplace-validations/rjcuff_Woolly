package connect

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// PluginName is the name of Woolly's Claude Code plugin.
const PluginName = "woolly"

// PluginInstalled reports whether Woolly's Claude Code plugin is installed.
// The plugin runs the same MCP server and hooks that `woolly connect` adds, so
// having both would run every hook twice.
func PluginInstalled(home string) bool {
	dir := filepath.Join(home, ".claude")
	if d := os.Getenv("CLAUDE_CONFIG_DIR"); d != "" {
		dir = d
	}
	raw, err := os.ReadFile(filepath.Join(dir, "plugins", "installed_plugins.json"))
	if err != nil {
		return false
	}
	var f struct {
		Plugins map[string]json.RawMessage `json:"plugins"`
	}
	if json.Unmarshal(raw, &f) != nil {
		return false
	}
	for key := range f.Plugins {
		if strings.HasPrefix(key, PluginName+"@") {
			return true
		}
	}
	return false
}
