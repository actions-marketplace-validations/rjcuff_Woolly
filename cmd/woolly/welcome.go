package main

import (
	"fmt"
	"strings"
)

// bigMammoth is the website icon at full size, for the first run.
const bigMammoth = `                      ▄████▄
                     ████████▄▄
                    ▄▀  ▀▀ ▀████▄
         ▄▄▄▄▄▄▄▄▄▄███▄▄▄▄▄███████▄
      ▄▀▀▀▀▀▀███████████████▄██████
   ▄█▀       ▄█████████████████████
 ▄▄███████████████████████████▄▄███
███████████████████████████████████▄
███████████████████████████▀ ▀██████
 ██████████████████████████▄  ███████
  ▀████████████████████████▀▄ ▀▀▀▀ ██
   ████████████████████████ ▀▄▄▄▄▄███
   ████████████████████████   ▄█████
   ████████████▀███████████  █████▀
   ▀████▀▀████▀ ▀████▀████▀   ▀▀▀`

// cmdWelcome is what the installer shows once Woolly is in place.
func cmdWelcome(args []string) error {
	fs := newFlags("welcome")
	if _, err := parse(fs, args); err != nil {
		return err
	}
	fmt.Print(welcomeText())
	return nil
}

func welcomeText() string {
	var b strings.Builder
	b.WriteString("\n")
	for _, ln := range strings.Split(bigMammoth, "\n") {
		b.WriteString("  " + paint(blue, ln) + "\n")
	}
	fmt.Fprintf(&b, "\n  %s %s\n", paint(bold, "Woolly"), paint(dim, version))
	b.WriteString("  Memory that knows when it's wrong.\n\n")
	steps := [][2]string{
		{"woolly check", "in any repo: which lines in CLAUDE.md, AGENTS.md and your agents' memory name code that's gone"},
		{"woolly connect", "share one memory with Claude Code, Cursor, Codex and the rest"},
		{"woolly init --hooks", "keep checking after every commit, and warn agents at the start of each session"},
	}
	for i, s := range steps {
		fmt.Fprintf(&b, "  %s %s\n     %s\n", paint(blue, fmt.Sprintf("%d.", i+1)), paint(bold, s[0]), paint(dim, s[1]))
	}
	b.WriteString("\n  " + paint(dim, "Docs: https://usewoolly.com/docs") + "\n\n")
	return b.String()
}
