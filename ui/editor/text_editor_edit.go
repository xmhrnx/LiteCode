package editor

import (
	"strings"

	"litecode/internal/textbuf"
	"litecode/internal/textops"
)

// edit is one undo step being made: changes applied to the buffer as
// they're added.
type edit struct {
	e     *TextEditor
	step  textbuf.Step
	first int // the first line touched
}

// begin starts an edit from the current selection.
func (e *TextEditor) begin() *edit {
	return &edit{e: e, step: textbuf.Step{Before: e.sel}, first: e.buf.Lines()}
}

// replace puts s in place of r and returns where s ends.
func (t *edit) replace(r textbuf.Range, s string) textbuf.Pos {
	b := t.e.buf
	r = textbuf.Range{Start: b.Clamp(r.Start), End: b.Clamp(r.End)}
	removed := b.Slice(r)
	if removed == s {
		return r.End
	}
	end := b.Replace(r, s)
	t.step.Changes = append(t.step.Changes, textbuf.Change{At: r.Start, Removed: removed, Inserted: s})
	t.first = min(t.first, r.Start.Line)
	return end
}

// commit records the step with the selection as it is now.
func (t *edit) commit(kind textbuf.Kind) {
	if len(t.step.Changes) == 0 {
		return
	}
	t.step.After = t.e.sel
	t.e.hist.Record(t.step, kind)
	t.e.changed(t.first)
}

// insert replaces the selection with s and puts the caret after it.
func (e *TextEditor) insert(s string, kind textbuf.Kind) {
	t := e.begin()
	end := t.replace(e.sel.Range(), s)
	e.sel = caretAt(end)
	t.commit(kind)
}

// caretAt is a bare caret at p.
func caretAt(p textbuf.Pos) textbuf.Selection { return textbuf.Selection{Anchor: p, Caret: p} }

// deleteTo deletes from the caret to p, or the selection when there is
// one.
func (e *TextEditor) deleteTo(p textbuf.Pos) {
	r := e.sel.Range()
	if r.Empty() {
		r = textbuf.Ordered(e.sel.Caret, p)
	}
	if r.Empty() {
		return
	}
	t := e.begin()
	t.replace(r, "")
	e.sel = caretAt(r.Start)
	t.commit(textbuf.KindDeleting)
}

// applyWhole applies an action that rewrote the whole text, changing
// only the part that differs so undo stays small.
func (e *TextEditor) applyWhole(ed textops.Edit) {
	s, en, ins := textops.MinimalReplace(e.buf.Text(), ed.Text)
	t := e.begin()
	if s != en || ins != "" {
		t.replace(textbuf.Range{Start: e.buf.PosOf(s), End: e.buf.PosOf(en)}, ins)
	}
	e.sel = textbuf.Selection{Anchor: e.buf.PosOf(ed.Start), Caret: e.buf.PosOf(ed.End)}
	t.commit(textbuf.KindOther)
	e.follow = true
}

// SetText replaces the whole text as one undo step, keeping the caret
// near where it was.
func (e *TextEditor) SetText(text string) {
	c := e.buf.Offset(e.sel.Caret)
	e.applyWhole(textops.Edit{Text: text, Start: c, End: c})
}

// undo reverts the last step; redo applies it again.
func (e *TextEditor) undo() {
	s, ok := e.hist.Undo()
	if !ok {
		return
	}
	first := e.buf.Lines()
	for i := len(s.Changes) - 1; i >= 0; i-- {
		e.buf.Revert(s.Changes[i])
		first = min(first, s.Changes[i].At.Line)
	}
	e.sel = s.Before
	e.changed(first)
}

func (e *TextEditor) redo() {
	s, ok := e.hist.Redo()
	if !ok {
		return
	}
	first := e.buf.Lines()
	for _, c := range s.Changes {
		e.buf.Apply(c)
		first = min(first, c.At.Line)
	}
	e.sel = s.After
	e.changed(first)
}

// lineRange is the whole lines the selection touches, with the newline
// after them when there is one.
func (e *TextEditor) lineRange() textbuf.Range {
	r := e.sel.Range()
	last := r.End.Line
	if last > r.Start.Line && r.End.Col == 0 {
		last--
	}
	if last+1 < e.buf.Lines() {
		return textbuf.Range{Start: textbuf.Pos{Line: r.Start.Line}, End: textbuf.Pos{Line: last + 1}}
	}
	return textbuf.Range{Start: textbuf.Pos{Line: r.Start.Line}, End: e.buf.End()}
}

// copy puts the selection on the clipboard; with none, the caret's line.
func (e *TextEditor) copy() {
	if r := e.sel.Range(); !r.Empty() {
		e.ide.App.SetClipboard(e.buf.Slice(r))
		return
	}
	s := e.buf.Slice(e.lineRange())
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	e.ide.App.SetClipboard(s)
}

// cut is copy, then deleting what was copied.
func (e *TextEditor) cut() {
	e.copy()
	if e.sel.Range().Empty() {
		r := e.lineRange()
		e.sel = textbuf.Selection{Anchor: r.Start, Caret: r.End}
	}
	e.deleteTo(e.sel.Caret)
}

// paste puts the clipboard's text in place of the selection.
func (e *TextEditor) paste() {
	e.ide.App.ReadClipboard(func(s string) {
		s = strings.ReplaceAll(s, "\r\n", "\n")
		if s == "" {
			return
		}
		e.hist.Seal()
		e.insert(s, textbuf.KindOther)
	})
}

// indentLines adds (or with out, removes) a level of indentation on the
// selected lines, keeping them selected.
func (e *TextEditor) indentLines(out bool) {
	r := e.sel.Range()
	last := r.End.Line
	if last > r.Start.Line && r.End.Col == 0 {
		last--
	}
	unit := e.indentUnit()
	t := e.begin()
	shift := map[int]int{}
	for i := r.Start.Line; i <= last; i++ {
		if !out {
			if len(e.buf.Line(i)) > 0 {
				t.replace(textbuf.Range{Start: textbuf.Pos{Line: i}, End: textbuf.Pos{Line: i}}, unit)
				shift[i] = len([]rune(unit))
			}
			continue
		}
		n := removable(e.buf.Line(i), unit, e.tabSize())
		if n > 0 {
			t.replace(textbuf.Range{Start: textbuf.Pos{Line: i}, End: textbuf.Pos{Line: i, Col: n}}, "")
			shift[i] = -n
		}
	}
	move := func(p textbuf.Pos) textbuf.Pos {
		if p.Col > 0 || shift[p.Line] < 0 {
			p.Col = max(0, p.Col+shift[p.Line])
		}
		return p
	}
	e.sel = textbuf.Selection{Anchor: move(e.sel.Anchor), Caret: move(e.sel.Caret)}
	t.commit(textbuf.KindOther)
}

// removable is how many leading runes of line one level of unindent
// takes: a tab, or up to a level's worth of spaces.
func removable(line []rune, unit string, tab int) int {
	if len(line) > 0 && line[0] == '\t' {
		return 1
	}
	n := 0
	for n < len(line) && n < max(len(unit), tab) && line[n] == ' ' {
		n++
	}
	if unit == "\t" {
		n = min(n, tab)
	}
	return n
}
