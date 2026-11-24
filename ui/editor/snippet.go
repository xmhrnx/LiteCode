package editor

import "strings"

// plainSnippet turns snippet text ("Println(${1:a})$0") into plain text
// and where the caret goes in it, in runes: the first tab stop, else the
// final one ($0), else the end.
func plainSnippet(s string) (text string, caret int) {
	var b strings.Builder
	n, first, final := 0, -1, -1
	stop := func(num string) {
		switch {
		case num == "0":
			final = n
		case first < 0:
			first = n
		}
	}
	r := []rune(s)
	for i := 0; i < len(r); i++ {
		switch {
		case r[i] == '\\' && i+1 < len(r):
			i++
			b.WriteRune(r[i])
			n++
		case r[i] == '$' && i+1 < len(r) && r[i+1] >= '0' && r[i+1] <= '9':
			j := i + 1
			for j < len(r) && r[j] >= '0' && r[j] <= '9' {
				j++
			}
			stop(string(r[i+1 : j]))
			i = j - 1
		case r[i] == '$' && i+1 < len(r) && r[i+1] == '{':
			j := i + 2
			for j < len(r) && r[j] >= '0' && r[j] <= '9' {
				j++
			}
			stop(string(r[i+2 : j]))
			// The placeholder's text, up to the matching brace.
			depth, k := 1, j
			if k < len(r) && (r[k] == ':' || r[k] == '|') {
				k++
			}
			start := k
			for k < len(r) && depth > 0 {
				if r[k] == '{' {
					depth++
				} else if r[k] == '}' {
					depth--
				}
				k++
			}
			inner, _ := plainSnippet(string(r[start:max(start, k-1)]))
			if j < len(r) && r[j] == '|' {
				inner = strings.SplitN(inner, ",", 2)[0]
			}
			b.WriteString(inner)
			n += len([]rune(inner))
			i = k - 1
		default:
			b.WriteRune(r[i])
			n++
		}
	}
	switch {
	case first >= 0:
		return b.String(), first
	case final >= 0:
		return b.String(), final
	}
	return b.String(), n
}
