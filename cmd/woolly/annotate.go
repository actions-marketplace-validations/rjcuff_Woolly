package main

import (
	"fmt"
	"strings"

	"github.com/rjcuff/woolly/internal/audit"
	"github.com/rjcuff/woolly/internal/gitinfo"
)

// One line per stale line in an agent file, for editors, CI logs and
// GitHub's inline annotations.

const (
	formatText    = "text"
	formatCompact = "compact"
	formatGitHub  = "github"
)

// compactFindings prints each finding as path:line: message, the shape
// editors, Vim's quickfix list and problem matchers read. Memory outside
// the repo uses its full path so an editor can still open it.
func compactFindings(rep audit.Report) string {
	var b strings.Builder
	for _, f := range rep.Findings {
		where := f.Where()
		if personalSources[f.Source] && f.Path != "" {
			where = fmt.Sprintf("%s:%d", f.Path, f.N)
		}
		fmt.Fprintf(&b, "%s: %s. %q\n", where, findingWhy(f), clip(f.Text, 120))
	}
	return b.String()
}

// githubFindings prints GitHub Actions workflow commands, so each finding
// shows on its line in the pull request. Lines this change made stale are
// warnings; ones that were already stale on base are notices. Only files
// tracked in the repo get one, since GitHub can't show anything else.
func githubFindings(dir string, rep audit.Report, base string) string {
	tracked := map[string]bool{}
	for _, f := range gitinfo.Tracked(dir) {
		tracked[f] = true
	}
	now, before := splitByBase(dir, rep.Findings, base)
	var b strings.Builder
	emit := func(fs []audit.Finding, level, title string) {
		for _, f := range fs {
			if personalSources[f.Source] || !tracked[f.File] {
				continue
			}
			msg := findingWhy(f) + ". Agents still read this line as true."
			if f.Fix != "" {
				msg += " Run woolly check --fix to change it to " + f.Fix + "."
			}
			fmt.Fprintf(&b, "::%s file=%s,line=%d,title=%s::%s\n", level, ghProp(f.File), f.N, ghProp(title), ghData(msg))
		}
	}
	emit(now, "warning", "Stale agent note")
	emit(before, "notice", "Agent note already stale")
	return b.String()
}

// ghData escapes a workflow command's message.
func ghData(s string) string {
	return strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A").Replace(s)
}

// ghProp escapes a workflow command's property, like file or title.
func ghProp(s string) string {
	return strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A", ":", "%3A", ",", "%2C").Replace(s)
}
