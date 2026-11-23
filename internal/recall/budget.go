package recall

import "strings"

// Fit keeps results, best first, while their text adds up to at most
// maxChars. A result too long for what's left is skipped so a shorter one
// after it can still fit. maxChars <= 0 keeps everything.
func Fit(results []Result, maxChars int) []Result {
	if maxChars <= 0 {
		return results
	}
	var out []Result
	used := 0
	for _, r := range results {
		n := len(strings.TrimSpace(r.Memory.Content)) + 3 // "- " and a newline
		if used+n > maxChars {
			continue
		}
		used += n
		out = append(out, r)
	}
	return out
}

// FitLines keeps lines in order while they add up to at most maxChars, and
// returns the indexes it kept. maxChars <= 0 keeps everything.
func FitLines(lines []string, maxChars int) []int {
	var kept []int
	used := 0
	for i, l := range lines {
		n := len(l) + 1
		if maxChars > 0 && used+n > maxChars {
			continue
		}
		used += n
		kept = append(kept, i)
	}
	return kept
}
