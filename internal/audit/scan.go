package audit

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// scanPresent finds which of words appear as whole identifiers in the repo's
// files, reading each file once on every CPU. It stops when every word is
// found. It reports false when it ran out of time, so nothing should be
// taken as missing.
func scanPresent(dir string, files, words, skip []string, deadline time.Time) (map[string]bool, bool) {
	want := map[string]bool{}
	for _, w := range words {
		want[w] = true
	}
	found := map[string]bool{}
	if len(want) == 0 {
		return found, true
	}
	// A test file like webcrypto_test.ts counts as its name being there.
	for _, f := range files {
		base := f[strings.LastIndex(f, "/")+1:]
		if i := strings.Index(base, "."); i > 0 {
			base = base[:i]
		}
		if want[base] && !skipped(f, skip) {
			found[base] = true
		}
	}
	if len(found) == len(want) {
		return found, true
	}
	var mu sync.Mutex
	done := false
	jobs := make(chan string, 256)
	var wg sync.WaitGroup
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range jobs {
				b := readCode(filepath.Join(dir, filepath.FromSlash(p)))
				if b == nil {
					continue
				}
				var hits []string
				eachIdent(b, func(id []byte) {
					if want[string(id)] {
						hits = append(hits, string(id))
					}
				})
				if len(hits) == 0 {
					continue
				}
				mu.Lock()
				for _, h := range hits {
					found[h] = true
				}
				if len(found) == len(want) {
					done = true
				}
				mu.Unlock()
			}
		}()
	}
	ok := true
	for _, p := range files {
		if skipped(p, skip) {
			continue
		}
		mu.Lock()
		stop := done
		mu.Unlock()
		if stop {
			break
		}
		if time.Now().After(deadline) {
			ok = false
			break
		}
		jobs <- p
	}
	close(jobs)
	wg.Wait()
	return found, ok
}

func skipped(p string, skip []string) bool {
	for _, s := range skip {
		if p == s || strings.HasPrefix(p, s+"/") {
			return true
		}
	}
	return strings.Contains(p, "node_modules/") || strings.HasSuffix(p, ".md") || strings.HasSuffix(p, ".mdx") || strings.HasSuffix(p, ".txt")
}

// readCode reads a text file up to 2 MB, or returns nil for anything else.
func readCode(p string) []byte {
	st, err := os.Stat(p)
	if err != nil || !st.Mode().IsRegular() || st.Size() > 2<<20 || st.Size() == 0 {
		return nil
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return nil
	}
	head := b
	if len(head) > 8000 {
		head = head[:8000]
	}
	if bytes.IndexByte(head, 0) >= 0 {
		return nil
	}
	return b
}

func identByte(c byte) bool {
	return c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
}

// eachIdent calls fn with every run of letters, digits and underscores
// that starts with a letter or underscore.
func eachIdent(b []byte, fn func([]byte)) {
	for i := 0; i < len(b); {
		if !identByte(b[i]) {
			i++
			continue
		}
		j := i
		for j < len(b) && identByte(b[j]) {
			j++
		}
		if c := b[i]; c < '0' || c > '9' {
			fn(b[i:j])
		}
		i = j
	}
}
