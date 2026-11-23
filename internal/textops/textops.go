// Package textops implements the editor's line actions on plain text.
// Positions are rune offsets; a selection is [Start, End) with Start <= End
// (Start == End is a bare caret).
package textops

import "strings"

// Edit is a text with its selection after an action.
type Edit struct {
	Text       string
	Start, End int
}

// LineCol returns the 1-based line and column of a rune offset.
func LineCol(text string, pos int) (line, col int) {
	line, col = 1, 1
	for i, r := range []rune(text) {
		if i >= pos {
			break
		}
		if r == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
}

// Offset returns the rune offset of a 1-based line and column, clamped to the text.
func Offset(text string, line, col int) int {
	r := []rune(text)
	l := 1
	i := 0
	for i < len(r) && l < line {
		if r[i] == '\n' {
			l++
		}
		i++
	}
	for c := 1; c < col && i < len(r) && r[i] != '\n'; c++ {
		i++
	}
	return i
}

// lines returns the offsets of the start of the first line touched by
// [start, end) and the end of the last one, including its newline.
func lines(r []rune, start, end int) (ls, le int) {
	ls = start
	for ls > 0 && r[ls-1] != '\n' {
		ls--
	}
	if end > start && end > 0 && r[end-1] == '\n' {
		end-- // a selection ending at a line start doesn't take that line
	}
	le = end
	for le < len(r) && r[le] != '\n' {
		le++
	}
	if le < len(r) {
		le++
	}
	return ls, le
}

func splice(r []rune, from, to int, insert []rune) []rune {
	out := make([]rune, 0, len(r)-(to-from)+len(insert))
	out = append(out, r[:from]...)
	out = append(out, insert...)
	return append(out, r[to:]...)
}

// DuplicateLine copies the selection after itself, or the caret's line below it.
func DuplicateLine(text string, start, end int) Edit {
	r := []rune(text)
	if start != end {
		sel := append([]rune(nil), r[start:end]...)
		return Edit{Text: string(splice(r, end, end, sel)), Start: end, End: end + len(sel)}
	}
	ls, le := lines(r, start, end)
	line := append([]rune(nil), r[ls:le]...)
	if len(line) == 0 || line[len(line)-1] != '\n' {
		line = append([]rune{'\n'}, line...)
		caret := start + len(line)
		return Edit{Text: string(splice(r, le, le, line)), Start: caret, End: caret}
	}
	caret := start + len(line)
	return Edit{Text: string(splice(r, le, le, line)), Start: caret, End: caret}
}

// DeleteLine removes every line the selection touches.
func DeleteLine(text string, start, end int) Edit {
	r := []rune(text)
	ls, le := lines(r, start, end)
	if le == len(r) && ls > 0 && (le == 0 || r[le-1] != '\n') {
		ls-- // last line: take the newline before it instead
	}
	out := splice(r, ls, le, nil)
	caret := min(ls, len(out))
	return Edit{Text: string(out), Start: caret, End: caret}
}

// ToggleLineComment comments the touched lines out with prefix, or back in
// when they all are comments already. A single-line toggle moves the caret
// to the next line.
func ToggleLineComment(text string, start, end int, prefix string) Edit {
	if prefix == "" {
		return Edit{Text: text, Start: start, End: end}
	}
	r := []rune(text)
	ls, le := lines(r, start, end)
	block := string(r[ls:le])
	trailingNL := strings.HasSuffix(block, "\n")
	ls2 := strings.Split(strings.TrimSuffix(block, "\n"), "\n")
	allCommented := true
	for _, l := range ls2 {
		if strings.TrimSpace(l) != "" && !strings.HasPrefix(strings.TrimLeft(l, " \t"), prefix) {
			allCommented = false
		}
	}
	for i, l := range ls2 {
		if allCommented {
			trimmed := strings.TrimLeft(l, " \t")
			indent := l[:len(l)-len(trimmed)]
			rest := strings.TrimPrefix(trimmed, prefix)
			rest = strings.TrimPrefix(rest, " ")
			ls2[i] = indent + rest
		} else if strings.TrimSpace(l) != "" || len(ls2) == 1 {
			ls2[i] = prefix + " " + l
		}
	}
	newBlock := strings.Join(ls2, "\n")
	if trailingNL {
		newBlock += "\n"
	}
	nb := []rune(newBlock)
	out := splice(r, ls, le, nb)
	if start == end {
		next := min(ls+len(nb), len(out))
		_, col := LineCol(text, start)
		caret := Offset(string(out), lineOf(out, next), col)
		return Edit{Text: string(out), Start: caret, End: caret}
	}
	return Edit{Text: string(out), Start: ls, End: ls + len(nb)}
}

func lineOf(r []rune, pos int) int {
	n := 1
	for i := 0; i < pos && i < len(r); i++ {
		if r[i] == '\n' {
			n++
		}
	}
	return n
}

// MoveLines moves the touched lines up (dir < 0) or down one line.
func MoveLines(text string, start, end, dir int) Edit {
	r := []rune(text)
	if len(r) == 0 || r[len(r)-1] != '\n' {
		r = append(r, '\n')
	}
	ls, le := lines(r, start, end)
	block := append([]rune(nil), r[ls:le]...)
	if dir < 0 {
		if ls == 0 {
			return Edit{Text: text, Start: start, End: end}
		}
		ps, _ := lines(r, ls-1, ls-1)
		prev := append([]rune(nil), r[ps:ls]...)
		out := splice(r, ps, le, append(block, prev...))
		shift := len(prev)
		return trimEdit(out, text, start-shift, end-shift)
	}
	if le >= len(r) {
		return Edit{Text: text, Start: start, End: end}
	}
	_, ne := lines(r, le, le)
	next := append([]rune(nil), r[le:ne]...)
	out := splice(r, ls, ne, append(next, block...))
	shift := len(next)
	return trimEdit(out, text, start+shift, end+shift)
}

// trimEdit drops the newline MoveLines added to a text that had none at its end.
func trimEdit(out []rune, orig string, start, end int) Edit {
	s := string(out)
	if !strings.HasSuffix(orig, "\n") {
		s = strings.TrimSuffix(s, "\n")
	}
	n := len([]rune(s))
	return Edit{Text: s, Start: min(start, n), End: min(end, n)}
}

// Indent adds one level of indentation to the touched lines, or inserts it at
// the caret when nothing is selected.
func Indent(text string, start, end int, unit string) Edit {
	r := []rune(text)
	u := []rune(unit)
	if start == end {
		out := splice(r, start, start, u)
		return Edit{Text: string(out), Start: start + len(u), End: start + len(u)}
	}
	ls, le := lines(r, start, end)
	parts := strings.Split(string(r[ls:le]), "\n")
	added := 0
	for i, p := range parts {
		if p != "" {
			parts[i] = unit + p
			added += len(u)
		}
	}
	nb := []rune(strings.Join(parts, "\n"))
	out := splice(r, ls, le, nb)
	return Edit{Text: string(out), Start: ls, End: ls + len(nb)}
}

// Unindent removes up to one level of indentation from the touched lines.
func Unindent(text string, start, end int, unit string) Edit {
	r := []rune(text)
	ls, le := lines(r, start, end)
	parts := strings.Split(string(r[ls:le]), "\n")
	removed := 0
	for i, p := range parts {
		switch {
		case strings.HasPrefix(p, unit):
			parts[i] = p[len(unit):]
			removed += len([]rune(unit))
		case strings.HasPrefix(p, "\t"):
			parts[i] = p[1:]
			removed++
		default:
			trimmed := strings.TrimLeft(p, " ")
			n := min(len(p)-len(trimmed), len(unit))
			parts[i] = p[n:]
			removed += n
		}
	}
	nb := []rune(strings.Join(parts, "\n"))
	out := splice(r, ls, le, nb)
	if start == end {
		caret := max(ls, start-(len(r[ls:le])-len(nb)))
		return Edit{Text: string(out), Start: caret, End: caret}
	}
	return Edit{Text: string(out), Start: ls, End: ls + len(nb)}
}

// NewLine inserts a line break at the caret (replacing the selection) that
// keeps the current line's indentation, one level deeper after an opening
// bracket or a colon.
func NewLine(text string, start, end int, unit string) Edit {
	r := []rune(text)
	ls := start
	for ls > 0 && r[ls-1] != '\n' {
		ls--
	}
	i := ls
	for i < start && (r[i] == ' ' || r[i] == '\t') {
		i++
	}
	indent := string(r[ls:i])
	before := strings.TrimRight(string(r[ls:start]), " \t")
	if strings.HasSuffix(before, "{") || strings.HasSuffix(before, "(") || strings.HasSuffix(before, "[") || strings.HasSuffix(before, ":") {
		indent += unit
	}
	ins := []rune("\n" + indent)
	out := splice(r, start, end, ins)
	caret := start + len(ins)
	return Edit{Text: string(out), Start: caret, End: caret}
}
