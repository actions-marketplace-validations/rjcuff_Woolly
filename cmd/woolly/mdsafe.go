package main

import (
	"strings"

	"github.com/rjcuff/woolly/internal/secrets"
)

// Pull request comments quote lines from agent files and project facts.
// Anyone who can open a pull request can write those lines, so Woolly's
// comment must not turn them into mentions, links, images or HTML, and must
// not repeat a secret someone left in a file.

// codeSpan wraps s in inline code, so GitHub shows it as plain text. The
// fence is longer than any run of backticks inside s.
func codeSpan(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if s == "" {
		return ""
	}
	longest, run := 0, 0
	for _, r := range s {
		if r == '`' {
			run++
			if run > longest {
				longest = run
			}
		} else {
			run = 0
		}
	}
	fence := strings.Repeat("`", longest+1)
	if strings.HasPrefix(s, "`") || strings.HasSuffix(s, "`") {
		s = " " + s + " "
	}
	return fence + s + fence
}

// quoted is a line of someone's text, safe to put in a comment: inline code,
// or a note in its place when it holds something like an API key.
func quoted(s string) string {
	if found := secrets.Scan(s); len(found) > 0 {
		return "_(not shown: this line has " + secrets.Describe(found) + ")_"
	}
	return codeSpan(s)
}

// inert makes Woolly's own sentences safe when they name things from the
// repo, like a file under packages/@acme: no mentions, HTML or links.
var inertReplacer = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	"@", "@&#8203;",
	"[", "&#91;",
	"]", "&#93;",
)

func inert(s string) string { return inertReplacer.Replace(s) }
