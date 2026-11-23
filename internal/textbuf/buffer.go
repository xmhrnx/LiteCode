// Package textbuf is the editor's text: lines of runes, positions in them,
// edits and their undo history. It knows nothing about drawing.
package textbuf

import "strings"

// Pos is a place in the text: a 0-based line and a 0-based rune column.
type Pos struct {
	Line, Col int
}

// Range is the text between two positions, Start <= End.
type Range struct {
	Start, End Pos
}

// Buffer is the text as lines, without their "\n".
type Buffer struct {
	lines [][]rune
	rev   int    // goes up on every edit
	text  string // Text() as of textRev
	tRev  int
}

// Less reports whether p comes before q.
func (p Pos) Less(q Pos) bool {
	return p.Line < q.Line || p.Line == q.Line && p.Col < q.Col
}

// Ordered returns the range from the earlier of a and b to the later.
func Ordered(a, b Pos) Range {
	if b.Less(a) {
		return Range{b, a}
	}
	return Range{a, b}
}

// Empty reports whether the range holds no text.
func (r Range) Empty() bool { return r.Start == r.End }

// New makes a buffer holding text ("\n" line endings).
func New(text string) *Buffer {
	b := &Buffer{}
	b.SetText(text)
	return b
}

// SetText replaces everything.
func (b *Buffer) SetText(text string) {
	parts := strings.Split(text, "\n")
	b.lines = make([][]rune, len(parts))
	for i, p := range parts {
		b.lines[i] = []rune(p)
	}
	b.rev++
}

// Rev goes up on every change.
func (b *Buffer) Rev() int { return b.rev }

// Lines is the number of lines; an empty text has one.
func (b *Buffer) Lines() int { return len(b.lines) }

// Line is line i's runes; callers mustn't change them.
func (b *Buffer) Line(i int) []rune { return b.lines[i] }

// Text is the whole text.
func (b *Buffer) Text() string {
	if b.tRev == b.rev && b.rev != 0 {
		return b.text
	}
	var sb strings.Builder
	for i, l := range b.lines {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(string(l))
	}
	b.text, b.tRev = sb.String(), b.rev
	return b.text
}

// Clamp moves p into the text.
func (b *Buffer) Clamp(p Pos) Pos {
	p.Line = max(0, min(p.Line, len(b.lines)-1))
	p.Col = max(0, min(p.Col, len(b.lines[p.Line])))
	return p
}

// End is the position after the last rune.
func (b *Buffer) End() Pos {
	n := len(b.lines) - 1
	return Pos{n, len(b.lines[n])}
}

// Slice is the text in r.
func (b *Buffer) Slice(r Range) string {
	r = Range{b.Clamp(r.Start), b.Clamp(r.End)}
	if r.Start.Line == r.End.Line {
		return string(b.lines[r.Start.Line][r.Start.Col:r.End.Col])
	}
	var sb strings.Builder
	sb.WriteString(string(b.lines[r.Start.Line][r.Start.Col:]))
	for i := r.Start.Line + 1; i < r.End.Line; i++ {
		sb.WriteByte('\n')
		sb.WriteString(string(b.lines[i]))
	}
	sb.WriteByte('\n')
	sb.WriteString(string(b.lines[r.End.Line][:r.End.Col]))
	return sb.String()
}

// Replace puts s in place of r and returns where s ends.
func (b *Buffer) Replace(r Range, s string) Pos {
	r = Range{b.Clamp(r.Start), b.Clamp(r.End)}
	head := b.lines[r.Start.Line][:r.Start.Col]
	tail := b.lines[r.End.Line][r.End.Col:]
	parts := strings.Split(s, "\n")
	repl := make([][]rune, len(parts))
	for i, p := range parts {
		repl[i] = []rune(p)
	}
	last := len(repl) - 1
	end := Pos{r.Start.Line + last, len(repl[last])}
	if last == 0 {
		end.Col += len(head)
	}
	first := make([]rune, 0, len(head)+len(repl[0]))
	repl[0] = append(append(first, head...), repl[0]...)
	repl[last] = append(repl[last], tail...)
	b.lines = append(b.lines[:r.Start.Line], append(repl, b.lines[r.End.Line+1:]...)...)
	b.rev++
	return end
}

// Offset is p as a rune offset into Text().
func (b *Buffer) Offset(p Pos) int {
	p = b.Clamp(p)
	n := 0
	for i := 0; i < p.Line; i++ {
		n += len(b.lines[i]) + 1
	}
	return n + p.Col
}

// PosOf is the position of a rune offset into Text().
func (b *Buffer) PosOf(off int) Pos {
	for i, l := range b.lines {
		if off <= len(l) {
			return Pos{i, max(off, 0)}
		}
		off -= len(l) + 1
	}
	return b.End()
}
