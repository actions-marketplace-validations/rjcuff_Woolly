package main

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

// Terminal colors, only when printing to a terminal that wants them. Pipes,
// NO_COLOR, TERM=dumb and older Windows consoles get plain text.

type style string

const (
	bold  style = "1"
	dim   style = "2"
	red   style = "31"
	green style = "32"
	blue  style = "38;2;110;140;255" // Woolly blue, lightened to read on dark terminals
)

var colorOn = wantColor()

func wantColor() bool {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}
	if runtime.GOOS == "windows" && os.Getenv("WT_SESSION") == "" && os.Getenv("TERM_PROGRAM") == "" {
		return false
	}
	st, err := os.Stdout.Stat()
	return err == nil && st.Mode()&os.ModeCharDevice != 0
}

func paint(s style, text string) string {
	if !colorOn || text == "" {
		return text
	}
	return "\x1b[" + string(s) + "m" + text + "\x1b[0m"
}

// spin shows a small spinner on stderr while slow work runs, when stderr is
// a terminal. The returned function clears it.
func spin(label string) func() {
	st, err := os.Stderr.Stat()
	if err != nil || st.Mode()&os.ModeCharDevice == 0 || !colorOn {
		return func() {}
	}
	done := make(chan struct{})
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		t := time.NewTicker(80 * time.Millisecond)
		defer t.Stop()
		// Quick checks never show it.
		select {
		case <-done:
			return
		case <-time.After(400 * time.Millisecond):
		}
		for i := 0; ; i++ {
			fmt.Fprintf(os.Stderr, "\r%s %s", paint(blue, frames[i%len(frames)]), paint(dim, label+"..."))
			select {
			case <-done:
				fmt.Fprint(os.Stderr, "\r\x1b[2K")
				return
			case <-t.C:
			}
		}
	}()
	return func() {
		close(done)
		<-finished
	}
}
