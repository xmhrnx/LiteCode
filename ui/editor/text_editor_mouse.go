package editor

import (
	"gioui.org/io/key"
	"github.com/xmhrnx/giode"

	"litecode/internal/textbuf"
)

// posAt is the text position under a point of the scrolled lines, in dp.
func (e *TextEditor) posAt(x, y float32) textbuf.Pos {
	line := int(y / e.m.lineH)
	if y < 0 {
		line = -1
	}
	if line < 0 {
		return textbuf.Pos{}
	}
	if line >= e.buf.Lines() {
		return e.buf.End()
	}
	v := (x-e.m.gutterW-textPad)/e.m.charW + float32(e.scrollCol)
	return textbuf.Pos{Line: line, Col: runeCol(e.buf.Line(line), v, e.tabSize())}
}

// unitAt is the stretch of text a press selects at p: the rune boundary,
// the word or the line.
func (e *TextEditor) unitAt(p textbuf.Pos, unit int) textbuf.Range {
	switch unit {
	case 2:
		return e.buf.WordAt(p)
	case 3:
		end := textbuf.Pos{Line: p.Line + 1}
		if p.Line+1 >= e.buf.Lines() {
			end = e.buf.End()
		}
		return textbuf.Range{Start: textbuf.Pos{Line: p.Line}, End: end}
	}
	return textbuf.Range{Start: p, End: p}
}

// clicked places the caret, selects a word on a double click and a line
// on a triple one or a click on its number; Shift extends.
func (e *TextEditor) clicked(ev giode.ClickEvent) {
	if ev.Button != giode.MousePrimary {
		return
	}
	e.Focus()
	p := e.posAt(ev.PositionDp.X, ev.PositionDp.Y)
	unit := min(max(ev.Clicks, 1), 3)
	if ev.PositionDp.X < e.m.gutterW {
		unit = 3
	}
	r := e.unitAt(p, unit)
	if ev.Shift() {
		e.sel.Caret = r.End
		if r.End.Less(e.sel.Anchor) {
			e.sel.Caret = r.Start
		}
	} else {
		e.sel = textbuf.Selection{Anchor: r.Start, Caret: r.End}
	}
	e.wantX = -1
	e.hist.Seal()
}

// dragStarted begins selecting from where the press was.
func (e *TextEditor) dragStarted(ev giode.DragEvent) {
	if ev.Button != giode.MousePrimary {
		return
	}
	e.Focus()
	e.dragging = true
	e.dragUnit = 1
	if ev.PositionDp.X < e.m.gutterW {
		e.dragUnit = 3
	}
	p := e.posAt(ev.PositionDp.X, ev.PositionDp.Y)
	e.dragStart = e.unitAt(p, e.dragUnit)
	if ev.Modifiers.Contain(key.ModShift) {
		e.dragStart = textbuf.Range{Start: e.sel.Anchor, End: e.sel.Anchor}
	}
	e.dragMoved(ev)
	e.hist.Seal()
}

// dragMoved stretches the selection to the pointer; following the caret
// scrolls when it's past the top or bottom.
func (e *TextEditor) dragMoved(ev giode.DragEvent) {
	if !e.dragging {
		return
	}
	r := e.unitAt(e.posAt(ev.PositionDp.X, ev.PositionDp.Y), e.dragUnit)
	if r.Start.Less(e.dragStart.Start) {
		e.sel = textbuf.Selection{Anchor: e.dragStart.End, Caret: r.Start}
	} else {
		e.sel = textbuf.Selection{Anchor: e.dragStart.Start, Caret: r.End}
	}
	e.wantX = -1
	e.follow = true
}
