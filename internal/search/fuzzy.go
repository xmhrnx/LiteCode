// Package search finds files by name (Go to File, Search Anything) and
// text in files (Find in Files).
package search

import (
	"path"
	"sort"
	"strings"
	"unicode"
)

// FileMatch is a file that matched a name query.
type FileMatch struct {
	Path  string
	Score int
}

// MatchFiles ranks files against a query the way Go to File does: the query's
// characters must appear in order in the file name (or, with a "/", in the
// path); matches at word starts, consecutive runs and exact name prefixes
// score higher. At most limit matches are returned, best first.
func MatchFiles(files []string, query string, limit int) []FileMatch {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil
	}
	usePath := strings.Contains(q, "/")
	var out []FileMatch
	for _, f := range files {
		target := path.Base(f)
		if usePath {
			target = f
		}
		if s, ok := score(target, q); ok {
			// Shorter paths win ties: a file near the root is the likelier one.
			out = append(out, FileMatch{Path: f, Score: s*1000 - len(f)})
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

// score matches q (lower case) as a subsequence of name.
func score(name, q string) (int, bool) {
	lower := strings.ToLower(name)
	if strings.HasPrefix(lower, q) {
		return 100 + len(q)*4, true
	}
	runes := []rune(name)
	lr := []rune(lower)
	qi, total, run := 0, 0, 0
	qr := []rune(q)
	for i := 0; i < len(lr) && qi < len(qr); i++ {
		if lr[i] != qr[qi] {
			run = 0
			continue
		}
		pts := 1
		if i == 0 || isBoundary(runes[i-1], runes[i]) {
			pts += 3
		}
		run++
		pts += run
		total += pts
		qi++
	}
	if qi < len(qr) {
		return 0, false
	}
	return total, true
}

func isBoundary(prev, cur rune) bool {
	return !unicode.IsLetter(prev) && !unicode.IsDigit(prev) || unicode.IsLower(prev) && unicode.IsUpper(cur)
}
