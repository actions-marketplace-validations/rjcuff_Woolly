package main

import (
	"strings"
	"testing"
)

func TestMarkdownCheckHasFooterUnlessTurnedOff(t *testing.T) {
	t.Setenv("WOOLLY_HOME", t.TempDir())
	t.Setenv("WOOLLY_NO_EMBED", "1")
	dir, base := staleRepo(t)

	out := captureStdout(t, func() {
		if err := cmdCheck([]string{"--markdown", "--no-backfill", "--dir", dir, "--base", base}); err != nil && err != exitError(1) {
			t.Errorf("check: %v", err)
		}
	})
	if !strings.HasSuffix(out, commentFooter) {
		t.Errorf("comment should end with the footer:\n%s", out)
	}
	if strings.Count(out, "usewoolly.com/team") != 1 {
		t.Errorf("footer should link Woolly Team once:\n%s", out)
	}

	out = captureStdout(t, func() {
		_ = cmdCheck([]string{"--markdown", "--no-backfill", "--no-footer", "--dir", dir, "--base", base})
	})
	if strings.Contains(out, "usewoolly.com/team") {
		t.Errorf("--no-footer should leave the footer out:\n%s", out)
	}
	if !strings.Contains(out, "went stale") {
		t.Errorf("report body missing:\n%s", out)
	}
}
