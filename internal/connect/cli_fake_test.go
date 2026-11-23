package connect

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// When FAKE_MCP_CLI is set, the test binary acts as an app's CLI: `mcp get`
// says Woolly is connected, `mcp remove` works and `mcp add` fails.
func TestMain(m *testing.M) {
	if os.Getenv("FAKE_MCP_CLI") != "" {
		args := strings.Join(os.Args[1:], " ")
		switch {
		case strings.HasPrefix(args, "mcp add"):
			fmt.Println("boom: config is locked")
			os.Exit(1)
		case strings.HasPrefix(args, "mcp get"), strings.HasPrefix(args, "mcp remove"):
			os.Exit(0)
		}
		os.Exit(2)
	}
	os.Exit(m.Run())
}

func fakeCLI(t *testing.T) CLIApp {
	t.Helper()
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	name := "fakecli"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	src, err := os.Open(self)
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()
	dst, err := os.OpenFile(filepath.Join(dir, name), os.O_CREATE|os.O_WRONLY, 0o755)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		t.Fatal(err)
	}
	dst.Close()
	t.Setenv("PATH", dir)
	t.Setenv("FAKE_MCP_CLI", "1")
	return CLIApp{ID: "fake", Name: "Fake App", Bin: "fakecli"}
}

func TestScopedReconnectSaysWhenWoollyWasRemoved(t *testing.T) {
	c := fakeCLI(t)
	_, err := c.Connect("woolly", false, "--agent", "bot")
	if err == nil {
		t.Fatal("a failed add should be an error")
	}
	msg := err.Error()
	for _, want := range []string{"boom", "removed from Fake App", "isn't connected now", "fakecli mcp add woolly -- woolly mcp --agent bot"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error should say %q:\n%s", want, msg)
		}
	}
}

func TestUnscopedConnectLeavesAnExistingEntry(t *testing.T) {
	c := fakeCLI(t)
	res, err := c.Connect("woolly", false)
	if err != nil || res != Unchanged {
		t.Fatalf("Connect = %v, %v; want Unchanged without touching it", res, err)
	}
}
