package editor

import (
	"strconv"
	"strings"

	"github.com/xmhrnx/giode"

	"litecode/internal/highlight"
	"litecode/ui/theme"
)

// overscan is how many lines past each edge get drawn: the scroll
// applies this frame's wheel after the content is drawn.
const overscan = 12

// visibleLines is the range of lines to draw.
func (e *TextEditor) visibleLines() (first, last int) {
	off := e.offsetDp()
	first = max(0, int(off/e.m.lineH)-overscan)
	last = min(e.buf.Lines()-1, int((off+max(e.m.viewH, 600))/e.m.lineH)+overscan)
	return first, last
}

// measure finds the width of one column of the code font.
func (e *TextEditor) measure(p *giode.Painter) {
	if e.m.measuredF == e.m.fontSize {
		return
	}
	const n = 64
	w, _ := p.MeasureText(strings.Repeat("0", n), e.codeStyle(""))
	if w > 0 {
		e.m.charW, e.m.measuredF = w/n, e.m.fontSize
	}
}

// paintGutter draws the line numbers.
func (e *TextEditor) paintGutter(p *giode.Painter) {
	e.measure(p)
	cc := theme.CodeScheme()
	first, last := e.visibleLines()
	caret := e.sel.Caret.Line
	for i := first; i <= last; i++ {
		y := float32(i) * e.m.lineH
		color := cc.LineNumber
		if i == caret {
			p.Rect(0, y, p.Width(), e.m.lineH, 0, cc.CurrentLine)
			color = cc.LineActive
		}
		p.Text(p.Width()-gutterPad, y+e.m.lineH/2, strconv.Itoa(i+1), giode.TextAnchorRight, e.codeStyle(color))
	}
}

// paintText draws the current line, the selection, the indent guides,
// the text and the caret.
func (e *TextEditor) paintText(p *giode.Painter) {
	e.m.textW = p.Width()
	cc := theme.CodeScheme()
	first, last := e.visibleLines()
	tab := e.tabSize()
	cols := int(p.Width()/e.m.charW) + 2
	sel := e.sel.Range()
	focused := e.keys.Focused()
	x := func(v int) float32 { return textPad + float32(v-e.scrollCol)*e.m.charW }
	guide := e.guideWidth()
	for i := first; i <= last; i++ {
		y := float32(i) * e.m.lineH
		line := e.buf.Line(i)
		if i == e.sel.Caret.Line && sel.Empty() {
			p.Rect(0, y, p.Width(), e.m.lineH, 0, cc.CurrentLine)
		}
		if !sel.Empty() && i >= sel.Start.Line && i <= sel.End.Line {
			a, b := 0, screenCol(line, len(line), tab)+1
			if i == sel.Start.Line {
				a = screenCol(line, sel.Start.Col, tab)
			}
			if i == sel.End.Line {
				b = screenCol(line, sel.End.Col, tab)
			}
			if b > a {
				l := max(x(a), 0)
				p.Rect(l, y, x(b)-l, e.m.lineH, 3, cc.Selection)
			}
		}
		if guide > 0 {
			for v := 0; v < e.indentCols(i); v += guide {
				if gx := x(v); gx >= textPad {
					p.Rect(gx, y, 1, e.m.lineH, 0, cc.Guide)
				}
			}
		}
		e.paintLine(p, line, e.hl.Tokens(e.buf, i), y, cols)
	}
	if focused && e.sel.Caret.Line >= first && e.sel.Caret.Line <= last {
		e.paintCaret(p, x)
	}
}

// paintLine draws a line's runes in runs of one color, from the first
// shown column to the last.
func (e *TextEditor) paintLine(p *giode.Painter, line []rune, toks []highlight.Token, y float32, cols int) {
	cc := theme.CodeScheme()
	tab := e.tabSize()
	var run strings.Builder
	runKind, runAt := highlight.Plain, 0
	flush := func() {
		if s := run.String(); strings.TrimSpace(s) != "" {
			px := textPad + float32(runAt-e.scrollCol)*e.m.charW
			p.Text(px, y+e.m.lineH/2, s, giode.TextAnchorLeft, e.codeStyle(cc.Tokens[runKind]))
		}
		run.Reset()
	}
	v, t := 0, 0
	for i, r := range line {
		next := advance(v, r, tab)
		if v-e.scrollCol > cols {
			break
		}
		for t < len(toks) && toks[t].End <= i {
			t++
		}
		k := highlight.Plain
		if t < len(toks) && toks[t].Start <= i {
			k = toks[t].Kind
		}
		if next <= e.scrollCol {
			v = next
			continue
		}
		if k != runKind || run.Len() == 0 || wide(r) {
			flush()
			runKind, runAt = k, max(v, e.scrollCol)
		}
		if r == '\t' {
			run.WriteString(strings.Repeat(" ", next-max(v, e.scrollCol)))
		} else {
			run.WriteRune(r)
		}
		if wide(r) {
			flush()
			runAt = next
		}
		v = next
	}
	flush()
}

// paintCaret draws the caret, and the text an input method is composing
// in front of it.
func (e *TextEditor) paintCaret(p *giode.Painter, x func(int) float32) {
	cc := theme.CodeScheme()
	c := e.sel.Caret
	cx := x(screenCol(e.buf.Line(c.Line), c.Col, e.tabSize()))
	y := float32(c.Line) * e.m.lineH
	if comp := e.keys.Composition(); comp != "" {
		w, _ := p.MeasureText(comp, e.codeStyle(""))
		p.Rect(cx, y, w, e.m.lineH, 0, theme.Current().Editor)
		p.Text(cx, y+e.m.lineH/2, comp, giode.TextAnchorLeft, e.codeStyle(cc.Tokens[highlight.Plain]))
		p.Rect(cx, y+e.m.lineH-3, w, 1, 0, cc.Composition)
		cx += w
	}
	if cx >= textPad-1 {
		p.Rect(cx-1, y+2, 2, e.m.lineH-4, 1, cc.Caret)
	}
}

// guideWidth is the columns between indent guides; 0 draws none.
func (e *TextEditor) guideWidth() int {
	if e.lang.IndentTabs {
		return e.tabSize()
	}
	return len(e.indentUnit())
}

// indentCols is how far line i is indented, in columns; a blank line
// takes the indentation of the lines around it.
func (e *TextEditor) indentCols(i int) int {
	if c, ok := e.ownIndent(i); ok {
		return c
	}
	up, down := 0, 0
	for j := i - 1; j >= max(0, i-50); j-- {
		if c, ok := e.ownIndent(j); ok {
			up = c
			break
		}
	}
	for j := i + 1; j < min(e.buf.Lines(), i+50); j++ {
		if c, ok := e.ownIndent(j); ok {
			down = c
			break
		}
	}
	return min(up, down)
}

// ownIndent is line i's indentation, or false for a blank line.
func (e *TextEditor) ownIndent(i int) (int, bool) {
	line := e.buf.Line(i)
	n := len(e.buf.Indentation(i))
	if n == len(line) {
		return 0, false
	}
	return screenCol(line, n, e.tabSize()), true
}
