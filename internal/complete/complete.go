// Package complete ranks completion candidates against what's been typed,
// and finds the words of a text to complete from when no language server
// has anything.
package complete

import (
	"sort"
	"strings"
	"unicode"
)

// MaxItems is how many candidates a ranking keeps.
const MaxItems = 200

// Item is a candidate: what's shown, what it's matched on and sorted by,
// and the index of what it stands for in the caller's list.
type Item struct {
	Label  string
	Filter string // matched against what's typed; Label when empty
	Sort   string // the source's own order; ties broken by it
	Index  int
	score  int
}

// match scores how well typed matches s: a prefix beats a word start
// ("gS" in "getString") beats letters in order; -1 is no match.
func match(s, typed string) int {
	if typed == "" {
		return 0
	}
	ls, lt := strings.ToLower(s), strings.ToLower(typed)
	switch {
	case strings.HasPrefix(s, typed):
		return 1000 - len(s)
	case strings.HasPrefix(ls, lt):
		return 900 - len(s)
	}
	// Letters in order, rewarding ones at word starts.
	rs := []rune(s)
	score, j := 0, 0
	tr := []rune(lt)
	for i := 0; i < len(rs) && j < len(tr); i++ {
		if unicode.ToLower(rs[i]) != tr[j] {
			continue
		}
		start := i == 0 || unicode.IsUpper(rs[i]) && !unicode.IsUpper(rs[i-1]) || rs[i-1] == '_' || rs[i-1] == '.'
		if start {
			score += 10
		} else {
			score++
		}
		j++
	}
	if j < len(tr) {
		return -1
	}
	return 100 + score - len(s)/4
}

// Rank keeps the items that match typed, best first.
func Rank(items []Item, typed string) []Item {
	out := make([]Item, 0, min(len(items), MaxItems))
	for _, it := range items {
		f := it.Filter
		if f == "" {
			f = it.Label
		}
		if f == typed {
			// Already typed in full: nothing to add.
			continue
		}
		if s := match(f, typed); s >= 0 {
			it.score = s
			out = append(out, it)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].score != out[j].score {
			return out[i].score > out[j].score
		}
		if out[i].Sort != out[j].Sort {
			return out[i].Sort < out[j].Sort
		}
		return out[i].Label < out[j].Label
	})
	if len(out) > MaxItems {
		out = out[:MaxItems]
	}
	return out
}

// IsWordRune reports whether r can be part of an identifier.
func IsWordRune(r rune) bool {
	return r == '_' || r == '$' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// Words are the distinct identifiers of the lines, at least three runes
// long, nearest to line first, without skip (the word being typed).
func Words(lines func(i int) []rune, count, near int, skip string) []string {
	seen := map[string]bool{skip: true}
	var out []string
	add := func(i int) {
		l := lines(i)
		for s := 0; s < len(l); {
			if !IsWordRune(l[s]) || unicode.IsDigit(l[s]) {
				s++
				continue
			}
			e := s
			for e < len(l) && IsWordRune(l[e]) {
				e++
			}
			if w := string(l[s:e]); e-s >= 3 && !seen[w] {
				seen[w] = true
				out = append(out, w)
			}
			s = e
		}
	}
	// Outward from the caret's line, so nearby words come first.
	for d := 0; d < count && len(out) < 5000; d++ {
		if near-d >= 0 {
			add(near - d)
		}
		if d > 0 && near+d < count {
			add(near + d)
		}
		if near-d < 0 && near+d >= count {
			break
		}
	}
	return out
}
