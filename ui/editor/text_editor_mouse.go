package editor

import (
	"github.com/xmhrnx/giode"

	"litecode/internal/textbuf"
)

// posAt is the text position at a point of the scrolled lines, in dp.
func (e *TextEditor) posAt(x, y float32) textbuf.Pos {
	if y < 0 {
		return textbuf.Pos{}
	}
	line := int(y / e.m.lineH)
	if line >= e.buf.Lines() {
		return e.buf.End()
	}
	v := (x - e.m.gutterW - textPad) / e.m.charW
	return textbuf.Pos{Line: line, Col: runeCol(e.buf.Line(line), v, e.tabSize())}
}

// inGutter reports whether a point of the scrolled lines is on the line
// numbers, which stay at the left edge.
func (e *TextEditor) inGutter(x float32) bool { return x-e.offsetXDp() < e.m.gutterW }

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

// pressed places the caret where the button goes down: a second press
// selects a word, a third or a press on a line number the line, and
// Shift extends the selection. A drag from there selects by the same
// unit.
func (e *TextEditor) pressed(ev giode.ClickEvent) {
	if ev.Button != giode.MousePrimary {
		return
	}
	e.Focus()
	unit := min(max(ev.Clicks, 1), 3)
	if e.inGutter(ev.PositionDp.X) {
		unit = 3
	}
	r := e.unitAt(e.posAt(ev.PositionDp.X, ev.PositionDp.Y), unit)
	e.dragUnit = unit
	if ev.Shift() {
		e.dragStart = textbuf.Range{Start: e.sel.Anchor, End: e.sel.Anchor}
		e.extendTo(r)
	} else {
		e.dragStart = r
		e.sel = textbuf.Selection{Anchor: r.Start, Caret: r.End}
	}
	e.wantX = -1
	e.hist.Seal()
}

// dragMoved stretches the selection to the pointer; following the caret
// scrolls when it's past an edge.
func (e *TextEditor) dragMoved(ev giode.DragEvent) {
	if !e.dragging || ev.Button != giode.MousePrimary {
		return
	}
	e.extendTo(e.unitAt(e.posAt(ev.PositionDp.X, ev.PositionDp.Y), e.dragUnit))
	e.follow = true
}

// extendTo selects from where the press started to the unit r.
func (e *TextEditor) extendTo(r textbuf.Range) {
	if r.Start.Less(e.dragStart.Start) {
		e.sel = textbuf.Selection{Anchor: e.dragStart.End, Caret: r.Start}
	} else {
		e.sel = textbuf.Selection{Anchor: e.dragStart.Start, Caret: r.End}
	}
}
