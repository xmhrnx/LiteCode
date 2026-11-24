package editor

import (
	"strings"

	"gioui.org/io/key"
	"github.com/xmhrnx/giode"

	"litecode/internal/textbuf"
	"litecode/ui/widgets"
)

// pairs are the brackets and quotes typing one of closes.
var pairs = map[rune]rune{'(': ')', '[': ']', '{': '}', '"': '"', '\'': '\'', '`': '`'}

// typed inserts text typed on the keyboard, closing brackets and quotes,
// typing over a closer that's already there, and wrapping a selection in
// a pair.
func (e *TextEditor) typed(s string) {
	widgets.NoteTyping()
	runes := []rune(s)
	if len(runes) != 1 || !e.typedPair(runes[0]) {
		e.insert(s, textbuf.KindTyping)
	}
	e.afterTyping(s)
}

func (e *TextEditor) typedPair(r rune) bool {
	c := e.sel.Caret
	line := e.buf.Line(c.Line)
	next := rune(0)
	if c.Col < len(line) {
		next = line[c.Col]
	}
	sel := e.sel.Range()
	if closer, ok := pairs[r]; ok && !sel.Empty() {
		// Wrap the selection.
		t := e.begin()
		t.replace(textbuf.Range{Start: sel.End, End: sel.End}, string(closer))
		t.replace(textbuf.Range{Start: sel.Start, End: sel.Start}, string(r))
		e.sel = textbuf.Selection{Anchor: shiftCol(sel.Start, 1), Caret: shiftCol(sel.End, sel.Start.Line == sel.End.Line)}
		t.commit(textbuf.KindOther)
		return true
	}
	if sel.Empty() && next == r && strings.ContainsRune(")]}\"'`", r) {
		// Type over the closer.
		e.sel = caretAt(textbuf.Pos{Line: c.Line, Col: c.Col + 1})
		e.follow = true
		return true
	}
	closer, ok := pairs[r]
	if !ok || !sel.Empty() {
		return false
	}
	if next != 0 && !strings.ContainsRune(" \t)]};,:", next) {
		return false
	}
	if r == closer {
		// A quote: not after a word character (it's an apostrophe).
		if c.Col > 0 && isWordRune(line[c.Col-1]) {
			return false
		}
	}
	t := e.begin()
	t.replace(textbuf.Range{Start: c, End: c}, string(r)+string(closer))
	e.sel = caretAt(textbuf.Pos{Line: c.Line, Col: c.Col + 1})
	t.commit(textbuf.KindTyping)
	return true
}

// shiftCol moves p right by n columns, or by one when n is true.
func shiftCol[T int | bool](p textbuf.Pos, n T) textbuf.Pos {
	switch v := any(n).(type) {
	case int:
		p.Col += v
	case bool:
		if v {
			p.Col++
		}
	}
	return p
}

func isWordRune(r rune) bool {
	return r == '_' || r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r > 127
}

// key handles the keys that don't type: moving, selecting, deleting,
// new lines, indentation, the clipboard and undo. Strokes the keymap
// binds go to their actions instead.
func (e *TextEditor) key(ev giode.KeyEvent) bool {
	if ev.State != key.Press || widgets.IsModifier(ev.Name) {
		return false
	}
	if e.ide.CapturingKeys || e.ide.Keymap.Bound(widgets.StrokeOf(ev)) {
		return false
	}
	mod := ev.Modifiers.Contain(key.ModShortcut)
	shift := ev.Modifiers.Contain(key.ModShift)
	if e.comp != nil {
		if !mod && !shift && e.completionKey(ev.Name) {
			return true
		}
		if ev.Name != key.NameDeleteBackward || mod {
			e.closeCompletion()
		}
	}
	if ev.Modifiers.Contain(key.ModAlt) {
		return false
	}
	if mod {
		return e.shortcutKey(ev.Name, shift)
	}
	switch ev.Name {
	case key.NameLeftArrow, key.NameRightArrow, key.NameUpArrow, key.NameDownArrow,
		key.NameHome, key.NameEnd, key.NamePageUp, key.NamePageDown:
		e.moveKey(ev.Name, false, shift)
	case key.NameDeleteBackward:
		e.backspace()
		if e.comp != nil {
			e.refilter()
			e.requestCompletions()
		}
	case key.NameDeleteForward:
		e.deleteTo(e.right(e.sel.Caret))
	case key.NameReturn, key.NameEnter:
		e.newLine()
	case key.NameTab:
		e.tab(shift)
	case key.NameEscape:
		if e.sel.Range().Empty() {
			return false
		}
		e.sel.Anchor = e.sel.Caret
	default:
		return false
	}
	return true
}

// shortcutKey handles Ctrl (Command on macOS) strokes.
func (e *TextEditor) shortcutKey(name key.Name, shift bool) bool {
	switch name {
	case "A":
		e.sel = textbuf.Selection{Anchor: textbuf.Pos{}, Caret: e.buf.End()}
	case "C":
		e.copy()
	case "X":
		e.cut()
	case "V":
		e.paste()
	case "Z":
		if shift {
			e.redo()
		} else {
			e.undo()
		}
	case key.NameDeleteBackward:
		e.deleteTo(e.buf.WordLeft(e.sel.Caret))
	case key.NameDeleteForward:
		e.deleteTo(e.buf.WordRight(e.sel.Caret))
	case key.NameUpArrow, key.NameDownArrow:
		// Scroll a line, leaving the caret.
		d := e.m.lineH * e.m.pxPerDp
		if name == key.NameUpArrow {
			d = -d
		}
		e.scroll.ScrollBy(0, d)
	case key.NameLeftArrow, key.NameRightArrow, key.NameHome, key.NameEnd:
		e.moveKey(name, true, shift)
	default:
		return false
	}
	return true
}

// moveKey moves the caret; with shift it extends the selection, with
// word by words (Left, Right) or to the ends of the text (Home, End).
func (e *TextEditor) moveKey(name key.Name, word, shift bool) {
	c := e.sel.Caret
	sel := e.sel.Range()
	keepX := false
	switch name {
	case key.NameLeftArrow:
		switch {
		case !shift && !sel.Empty() && !word:
			c = sel.Start
		case word:
			c = e.buf.WordLeft(c)
		default:
			c = e.left(c)
		}
	case key.NameRightArrow:
		switch {
		case !shift && !sel.Empty() && !word:
			c = sel.End
		case word:
			c = e.buf.WordRight(c)
		default:
			c = e.right(c)
		}
	case key.NameUpArrow:
		c, keepX = e.vertical(c, -1), true
	case key.NameDownArrow:
		c, keepX = e.vertical(c, 1), true
	case key.NamePageUp, key.NamePageDown:
		n := max(1, int(e.m.viewH/e.m.lineH)-1)
		if name == key.NamePageUp {
			n = -n
		}
		e.scroll.ScrollBy(0, float32(n)*e.m.lineH*e.m.pxPerDp)
		c, keepX = e.vertical(c, n), true
	case key.NameHome:
		if word {
			c = textbuf.Pos{}
		} else {
			// Smart Home: the first non-blank, then the line's start.
			ind := len([]rune(e.buf.Indentation(c.Line)))
			if c.Col == ind {
				ind = 0
			}
			c.Col = ind
		}
	case key.NameEnd:
		if word {
			c = e.buf.End()
		} else {
			c.Col = len(e.buf.Line(c.Line))
		}
	}
	e.sel.Caret = c
	if !shift {
		e.sel.Anchor = c
	}
	if !keepX {
		e.wantX = -1
	}
	e.follow = true
	e.hist.Seal()
}

func (e *TextEditor) left(p textbuf.Pos) textbuf.Pos {
	switch {
	case p.Col > 0:
		p.Col--
	case p.Line > 0:
		p.Line--
		p.Col = len(e.buf.Line(p.Line))
	}
	return p
}

func (e *TextEditor) right(p textbuf.Pos) textbuf.Pos {
	switch {
	case p.Col < len(e.buf.Line(p.Line)):
		p.Col++
	case p.Line < e.buf.Lines()-1:
		p.Line++
		p.Col = 0
	}
	return p
}

// vertical moves p n lines, keeping to the screen column it started in.
func (e *TextEditor) vertical(p textbuf.Pos, n int) textbuf.Pos {
	tab := e.tabSize()
	if e.wantX < 0 {
		e.wantX = screenCol(e.buf.Line(p.Line), p.Col, tab)
	}
	to := p.Line + n
	switch {
	case to < 0:
		return textbuf.Pos{}
	case to >= e.buf.Lines():
		return e.buf.End()
	}
	line := e.buf.Line(to)
	return textbuf.Pos{Line: to, Col: runeCol(line, float32(e.wantX), tab)}
}

// backspace deletes the rune before the caret, both halves of an empty
// pair, or back to the previous indent stop in leading spaces.
func (e *TextEditor) backspace() {
	c := e.sel.Caret
	if !e.sel.Range().Empty() || c.Col == 0 {
		e.deleteTo(e.left(c))
		return
	}
	line := e.buf.Line(c.Line)
	if c.Col < len(line) && pairs[line[c.Col-1]] == line[c.Col] {
		t := e.begin()
		t.replace(textbuf.Range{Start: textbuf.Pos{Line: c.Line, Col: c.Col - 1}, End: textbuf.Pos{Line: c.Line, Col: c.Col + 1}}, "")
		e.sel = caretAt(textbuf.Pos{Line: c.Line, Col: c.Col - 1})
		t.commit(textbuf.KindDeleting)
		return
	}
	if ind := len([]rune(e.buf.Indentation(c.Line))); c.Col <= ind && line[c.Col-1] == ' ' {
		unit := max(1, len(e.indentUnit()))
		to := c.Col - 1
		for to > 0 && to%unit != 0 && line[to-1] == ' ' {
			to--
		}
		e.deleteTo(textbuf.Pos{Line: c.Line, Col: to})
		return
	}
	e.deleteTo(e.left(c))
}

// newLine breaks the line, carrying the indentation over and adding a
// level after an opening bracket or a colon; between a bracket pair the
// closer goes on a line of its own.
func (e *TextEditor) newLine() {
	r := e.sel.Range()
	line := e.buf.Line(r.Start.Line)
	ind := e.buf.Indentation(r.Start.Line)
	ind = string([]rune(ind)[:min(len([]rune(ind)), r.Start.Col)])
	before := strings.TrimRight(string(line[:r.Start.Col]), " \t")
	after := e.buf.Line(r.End.Line)[r.End.Col:]
	opens := strings.HasSuffix(before, "{") || strings.HasSuffix(before, "(") || strings.HasSuffix(before, "[") ||
		strings.HasSuffix(before, ":") && (e.lang.ID == "python" || e.lang.ID == "yaml")
	t := e.begin()
	if !opens {
		end := t.replace(r, "\n"+ind)
		e.sel = caretAt(end)
		t.commit(textbuf.KindOther)
		return
	}
	inner := ind + e.indentUnit()
	text := "\n" + inner
	closes := len(after) > 0 && strings.ContainsRune(")]}", after[0]) && len(before) > 0 && pairs[rune(before[len(before)-1])] == after[0]
	if closes {
		text += "\n" + ind
	}
	t.replace(r, text)
	e.sel = caretAt(textbuf.Pos{Line: r.Start.Line + 1, Col: len([]rune(inner))})
	t.commit(textbuf.KindOther)
}

// tab indents (or with shift, unindents) the selected lines, or inserts
// indentation at the caret.
func (e *TextEditor) tab(shift bool) {
	r := e.sel.Range()
	if shift || r.Start.Line != r.End.Line {
		e.indentLines(shift)
		return
	}
	unit := e.indentUnit()
	if unit != "\t" {
		// Spaces to the next stop.
		v := screenCol(e.buf.Line(r.Start.Line), r.Start.Col, e.tabSize())
		unit = strings.Repeat(" ", len(unit)-v%len(unit))
	}
	e.insert(unit, textbuf.KindTyping)
}
