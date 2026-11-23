package connect

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestScopeArgs(t *testing.T) {
	cases := []struct {
		user, agent string
		want        []string
	}{
		{"", "", nil},
		{"u1", "", []string{"--user", "u1"}},
		{"", "bot", []string{"--agent", "bot"}},
		{"u1", "bot", []string{"--user", "u1", "--agent", "bot"}},
	}
	for _, c := range cases {
		if got := ScopeArgs(c.user, c.agent); !reflect.DeepEqual(got, c.want) {
			t.Errorf("ScopeArgs(%q, %q) = %v, want %v", c.user, c.agent, got, c.want)
		}
	}
}

func TestEntryWithScope(t *testing.T) {
	extra := []string{"--agent", "support bot"}
	cases := []struct {
		f    Format
		want map[string]any
	}{
		{FormatMCPServers, map[string]any{"command": "woolly", "args": []any{"mcp", "--agent", "support bot"}}},
		{FormatServers, map[string]any{"type": "stdio", "command": "woolly", "args": []any{"mcp", "--agent", "support bot"}}},
		{FormatOpenCode, map[string]any{"type": "local", "command": []any{"woolly", "mcp", "--agent", "support bot"}}},
	}
	for _, c := range cases {
		if got := Entry(c.f, "woolly", extra...); !reflect.DeepEqual(got, c.want) {
			t.Errorf("Entry(%v) = %v, want %v", c.f, got, c.want)
		}
	}
	// No extra keeps the plain entry.
	if got := Entry(FormatMCPServers, "woolly"); !reflect.DeepEqual(got["args"], []any{"mcp"}) {
		t.Errorf("plain args = %v", got["args"])
	}
}

func TestConnectScopedUpdatesPlainEntry(t *testing.T) {
	a := testApp(t, FormatMCPServers)
	if _, err := Connect(a, "woolly", false); err != nil {
		t.Fatal(err)
	}
	res, err := Connect(a, "woolly", false, "--user", "u1")
	if err != nil || res != Updated {
		t.Fatalf("scoped Connect = %v, %v; want Updated", res, err)
	}
	entry := readJSON(t, a.Config)["mcpServers"].(map[string]any)["woolly"].(map[string]any)
	if !reflect.DeepEqual(entry["args"], []any{"mcp", "--user", "u1"}) {
		t.Fatalf("args = %v", entry["args"])
	}
	res, err = Connect(a, "woolly", false, "--user", "u1")
	if err != nil || res != Unchanged {
		t.Fatalf("same scope again = %v, %v; want Unchanged", res, err)
	}
}

func TestSnippetWithScope(t *testing.T) {
	s := Snippet(FormatMCPServers, "woolly", "--agent", "bot")
	if !strings.Contains(s, `"--agent"`) || !strings.Contains(s, `"bot"`) {
		t.Fatalf("snippet missing scope:\n%s", s)
	}
}

func TestAnythingLLMPaths(t *testing.T) {
	cases := []struct {
		env  Env
		want string
	}{
		{Env{GOOS: "darwin", Home: "/Users/a"}, filepath.Join("/Users/a", "Library", "Application Support", "anythingllm-desktop", "storage", "plugins", "anythingllm_mcp_servers.json")},
		{Env{GOOS: "linux", Home: "/home/a"}, filepath.Join("/home/a", ".config", "anythingllm-desktop", "storage", "plugins", "anythingllm_mcp_servers.json")},
		{Env{GOOS: "windows", Home: `C:\Users\a`, AppData: `C:\Users\a\AppData\Roaming`}, filepath.Join(`C:\Users\a\AppData\Roaming`, "anythingllm-desktop", "storage", "plugins", "anythingllm_mcp_servers.json")},
	}
	for _, c := range cases {
		found := false
		for _, a := range Apps(c.env) {
			if a.ID != "anythingllm" {
				continue
			}
			found = true
			if a.Config != c.want || a.Format != FormatMCPServers {
				t.Errorf("%s: %s (%v), want %s", c.env.GOOS, a.Config, a.Format, c.want)
			}
		}
		if !found {
			t.Errorf("%s: no anythingllm app", c.env.GOOS)
		}
	}
}
