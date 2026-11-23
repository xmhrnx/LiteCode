package textbuf

import "unicode"

// class sorts runes for word moves: words, spaces and punctuation.
func class(r rune) int {
	switch {
	case r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r):
		return 2
	case unicode.IsSpace(r):
		return 0
	}
	return 1
}

// WordLeft is where Ctrl+Left goes from p: the start of the word before it.
func (b *Buffer) WordLeft(p Pos) Pos {
	p = b.Clamp(p)
	if p.Col == 0 {
		if p.Line == 0 {
			return p
		}
		return Pos{p.Line - 1, len(b.lines[p.Line-1])}
	}
	l := b.lines[p.Line]
	c := p.Col
	for c > 0 && class(l[c-1]) == 0 {
		c--
	}
	if c > 0 {
		k := class(l[c-1])
		for c > 0 && class(l[c-1]) == k {
			c--
		}
	}
	return Pos{p.Line, c}
}

// WordRight is where Ctrl+Right goes from p: the end of the next word.
func (b *Buffer) WordRight(p Pos) Pos {
	p = b.Clamp(p)
	l := b.lines[p.Line]
	if p.Col == len(l) {
		if p.Line == len(b.lines)-1 {
			return p
		}
		return Pos{p.Line + 1, 0}
	}
	c := p.Col
	for c < len(l) && class(l[c]) == 0 {
		c++
	}
	if c < len(l) {
		k := class(l[c])
		for c < len(l) && class(l[c]) == k {
			c++
		}
	}
	return Pos{p.Line, c}
}

// WordAt is the word (or run of spaces or punctuation) around p, for a
// double click.
func (b *Buffer) WordAt(p Pos) Range {
	p = b.Clamp(p)
	l := b.lines[p.Line]
	if len(l) == 0 {
		return Range{p, p}
	}
	at := min(p.Col, len(l)-1)
	if p.Col > 0 && (p.Col == len(l) || class(l[p.Col]) != 2 && class(l[p.Col-1]) == 2) {
		at = p.Col - 1
	}
	k := class(l[at])
	s, e := at, at+1
	for s > 0 && class(l[s-1]) == k {
		s--
	}
	for e < len(l) && class(l[e]) == k {
		e++
	}
	return Range{Pos{p.Line, s}, Pos{p.Line, e}}
}

// Indentation is line i's leading spaces and tabs.
func (b *Buffer) Indentation(i int) string {
	l := b.lines[i]
	n := 0
	for n < len(l) && (l[n] == ' ' || l[n] == '\t') {
		n++
	}
	return string(l[:n])
}
