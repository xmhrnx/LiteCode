package editor

import (
	"image"
	"math"
	"strconv"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/scroll"
	"github.com/xmhrnx/giode/styles/properties"

	"litecode/internal/highlight"
	"litecode/internal/lang"
	"litecode/internal/textbuf"
	"litecode/ui/core"
	"litecode/ui/theme"
)

// gutterPad is the room around the line numbers, and textPad the room
// between the gutter and the text.
const (
	gutterPad = 14
	textPad   = 10
)

// metrics are the sizes the editor lays out with, in dp.
type metrics struct {
	fontSize  int
	lineH     float32
	charW     float32
	gutterW   float32
	pxPerDp   float32
	viewH     float32 // the visible height, as of the last frame
	textW     float32 // the text area's width, as of the last frame
	measuredF int     // the font size charW was measured at
}

// TextEditor is the code editor: a buffer drawn as numbered, colored
// lines, a caret and a selection, edited from the keyboard and the
// mouse. It draws everything on canvases and takes the keyboard through
// a KeyboardInput.
type TextEditor struct {
	ide   *core.IDE
	lang  *lang.Language
	buf   *textbuf.Buffer
	hist  textbuf.History
	hl    *highlight.Highlighter
	sel   textbuf.Selection
	wantX int // the screen column Up and Down aim for; -1 takes the caret's

	scrollCol int  // the first screen column shown
	follow    bool // scroll the caret into view on the next frame
	center    bool // and put it mid-screen (after Go to Line)

	dragging  bool
	dragUnit  int // 1 by rune, 2 by word, 3 by line
	dragStart textbuf.Range

	m        metrics
	keys     *giode.KeyboardInputElement
	scroll   *scroll.Scroll
	press    *giode.PressableElement
	onChange func()
}

// NewTextEditor makes an editor holding text in a language.
func NewTextEditor(ide *core.IDE, l *lang.Language, text string, onChange func()) *TextEditor {
	e := &TextEditor{ide: ide, lang: l, buf: textbuf.New(text), wantX: -1, onChange: onChange}
	e.hl = highlight.New(highlight.For(l.ID, l.LineComment, l.BlockComment, l.Keywords))
	c := theme.Current()
	e.press = giode.Pressable(e.lines, giode.Styles{Cursor: giode.CursorText}).
		Focusable(false).
		Label("Code").
		OnClick(redrawn(e, e.clicked)).
		OnDragStart(redrawn(e, e.dragStarted)).
		OnDragMove(redrawn(e, e.dragMoved)).
		OnDragEnd(func(giode.DragEvent) { e.dragging = false })
	e.scroll = giode.Scroll(func() giode.Element { return e.press }, giode.Styles{Background: c.Editor, ScrollBar: theme.ScrollBar()})
	e.keys = giode.KeyboardInput(e.scroll).
		OnText(redrawn(e, e.typed)).
		OnKey(func(ev giode.KeyEvent) bool {
			took := e.key(ev)
			if took {
				e.ide.App.Invalidate()
			}
			return took
		}).
		Label("Code editor")
	return e
}

// redrawn wraps an input handler to ask for another frame after it: what
// laid out before the editor this frame (the status bar) shows the
// change then.
func redrawn[T any](e *TextEditor, fn func(T)) func(T) {
	return func(v T) {
		fn(v)
		e.ide.App.Invalidate()
	}
}

// Text is the whole text.
func (e *TextEditor) Text() string { return e.buf.Text() }

// Focus gives the editor the keyboard.
func (e *TextEditor) Focus() { e.keys.Focus() }

// Caret is the caret's position.
func (e *TextEditor) Caret() textbuf.Pos { return e.sel.Caret }

// SelectionOffsets is the selection as rune offsets into Text().
func (e *TextEditor) SelectionOffsets() (start, end int) {
	r := e.sel.Range()
	return e.buf.Offset(r.Start), e.buf.Offset(r.End)
}

// SelectedText is the text the selection holds.
func (e *TextEditor) SelectedText() string { return e.buf.Slice(e.sel.Range()) }

// GoTo puts the caret at p and scrolls it to the middle of the screen.
func (e *TextEditor) GoTo(p textbuf.Pos) {
	p = e.buf.Clamp(p)
	e.sel = textbuf.Selection{Anchor: p, Caret: p}
	e.wantX, e.follow, e.center = -1, true, true
	e.hist.Seal()
}

// Select selects the rune offsets [start, end).
func (e *TextEditor) Select(start, end int) {
	e.sel = textbuf.Selection{Anchor: e.buf.PosOf(start), Caret: e.buf.PosOf(end)}
	e.wantX, e.follow = -1, true
}

// tabSize is the width of a tab in columns.
func (e *TextEditor) tabSize() int { return max(1, e.ide.Settings.TabSize) }

// indentUnit is one level of indentation.
func (e *TextEditor) indentUnit() string { return e.lang.Indent(e.tabSize()) }

// codeStyle is the text style of a token color.
func (e *TextEditor) codeStyle(color string) giode.Styles {
	return giode.Styles{FontFamily: theme.CodeFont, FontSize: properties.FontSize(e.m.fontSize), Color: color}
}

// offsetDp is the vertical scroll offset in dp.
func (e *TextEditor) offsetDp() float32 {
	if e.m.pxPerDp == 0 {
		return 0
	}
	return e.scroll.Offset() / e.m.pxPerDp
}

// lines is what scrolls: the gutter and the text, as tall as the lines.
func (e *TextEditor) lines(giode.PressState) giode.Element {
	h := float32(e.buf.Lines())*e.m.lineH + max(e.m.viewH-3*e.m.lineH, 0)
	return giode.Row(giode.Styles{Wrap: giode.WrapNoWrap},
		giode.Canvas(giode.Styles{Width: properties.Width(e.m.gutterW), Height: properties.Height(h)}, e.paintGutter),
		giode.Canvas(giode.Styles{FlexGrow: 1, Height: properties.Height(h)}, e.paintText),
	)
}

// Layout follows the settings and the caret, then draws the editor.
func (e *TextEditor) Layout(gtx giode.Context) giode.Dimensions {
	s := e.ide.Settings
	e.m.fontSize, e.m.lineH, e.m.pxPerDp = s.EditorFontSize, float32(s.EditorLineHeight), gtx.Metric.PxPerDp
	e.m.viewH = float32(e.scroll.ViewportHeight()) / e.m.pxPerDp
	if e.m.charW == 0 {
		// Until the first canvas measures it.
		e.m.charW = float32(e.m.fontSize) * 0.6
	}
	digits := len(strconv.Itoa(e.buf.Lines()))
	e.m.gutterW = float32(max(digits, 3))*e.m.charW + 2*gutterPad
	if e.follow {
		e.scrollToCaret()
	}
	e.keys.Caret(e.caretRect())
	return e.keys.Layout(gtx)
}

// scrollToCaret scrolls just enough to show the caret, or centers it.
func (e *TextEditor) scrollToCaret() {
	if e.m.viewH <= 0 {
		return // not laid out yet: try next frame
	}
	e.follow = false
	top := float32(e.sel.Caret.Line) * e.m.lineH
	off := e.offsetDp()
	switch {
	case e.center:
		off = top - e.m.viewH/2 + e.m.lineH/2
	case top < off:
		off = top
	case top+e.m.lineH > off+e.m.viewH:
		off = top + e.m.lineH - e.m.viewH
	}
	e.center = false
	e.scroll.ScrollTo(max(off, 0) * e.m.pxPerDp)
	if e.m.textW > 0 {
		cols := int((e.m.textW - textPad - e.m.charW) / e.m.charW)
		v := screenCol(e.buf.Line(e.sel.Caret.Line), e.sel.Caret.Col, e.tabSize())
		switch {
		case v < e.scrollCol:
			e.scrollCol = max(0, v-8)
		case v > e.scrollCol+cols:
			e.scrollCol = v - cols + 8
		}
	}
}

// caretRect is where the caret is drawn, in pixels from the editor's
// corner, for the input method's candidate window.
func (e *TextEditor) caretRect() image.Rectangle {
	v := screenCol(e.buf.Line(e.sel.Caret.Line), e.sel.Caret.Col, e.tabSize())
	x := e.m.gutterW + textPad + float32(v-e.scrollCol)*e.m.charW
	y := float32(e.sel.Caret.Line)*e.m.lineH - e.offsetDp()
	px := func(f float32) int { return int(math.Round(float64(f * e.m.pxPerDp))) }
	return image.Rect(px(x), px(y), px(x+2), px(y+e.m.lineH))
}

// changed is called after every edit to the text, from line on.
func (e *TextEditor) changed(line int) {
	e.hl.Changed(line)
	e.follow, e.wantX = true, -1
	if e.onChange != nil {
		e.onChange()
	}
}
