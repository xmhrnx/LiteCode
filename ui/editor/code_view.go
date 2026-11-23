// Package editor is the editor area: the tabs of open files, the text of the
// selected one, and the status bar that describes it.
package editor

import (
	"context"
	"path/filepath"
	"unicode/utf8"

	"gioui.org/io/key"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"
	"github.com/xmhrnx/giode/styles/properties"

	"litecode/internal/format"
	"litecode/internal/lang"
	"litecode/internal/textops"
	"litecode/internal/workspace"
	"litecode/ui/core"
	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// CodeView edits one open file. It's built on Giode's multi-line text field,
// which has no gutter, highlighting or folding yet (see docs/GIODE_GAPS.md).
type CodeView struct {
	ide      *core.IDE
	doc      *workspace.Document
	lang     *lang.Language
	field    *input.Input
	text     string // the field's text as of the last change
	modified bool
	onChange func()
	applied  [2]int // font size and line height the field was styled with
}

// NewCodeView opens a document in a new view; onChange runs after every edit.
func NewCodeView(ide *core.IDE, doc *workspace.Document, onChange func()) *CodeView {
	v := &CodeView{ide: ide, doc: doc, lang: ide.Langs.ForPath(doc.Path), text: doc.Saved, onChange: onChange}
	v.field = widgets.TextField("", v.styles()).Multiline(true).Autocorrect(false).
		OnChange(widgets.Typing(v.changed)).
		OnKeys(v.onKey, key.NameTab, key.NameReturn, key.NameEnter)
	v.field.SetText(doc.Saved)
	v.field.SetCaret(0)
	return v
}

// styles is the field's style for the current editor settings.
func (v *CodeView) styles() giode.Styles {
	c := theme.Current()
	s := v.ide.Settings
	v.applied = [2]int{s.EditorFontSize, s.EditorLineHeight}
	return giode.Styles{
		Width: -1, Height: -1, Background: c.Editor, Color: c.TextCode, BorderColor: c.Editor, BorderRadius: 1,
		FontFamily: theme.CodeFont, FontSize: properties.FontSize(s.EditorFontSize), LineHeight: properties.LineHeight(s.EditorLineHeight),
		TabWidth: giode.TabWidth(s.TabSize), Padding: giode.Inset{Left: 16, Top: 8, Right: 8, Bottom: 8},
	}
}

// Path is the file's path.
func (v *CodeView) Path() string { return v.doc.Path }

// Language is the file's language.
func (v *CodeView) Language() *lang.Language { return v.lang }

// Modified reports unsaved changes.
func (v *CodeView) Modified() bool { return v.modified }

// Text is the current text.
func (v *CodeView) Text() string { return v.text }

// Focus gives the view the keyboard.
func (v *CodeView) Focus() { v.field.Focus() }

// Caret returns the caret's 1-based line and column.
func (v *CodeView) Caret() (line, col int) {
	return textops.LineCol(v.text, v.field.Caret())
}

// GoTo puts the caret at a 1-based line and column.
func (v *CodeView) GoTo(line, col int) {
	pos := textops.Offset(v.text, line, col)
	v.field.SetCaret(pos)
	v.field.Focus()
}

func (v *CodeView) changed(text string) {
	v.text = text
	v.modified = v.doc.Modified(text)
	if v.onChange != nil {
		v.onChange()
	}
}

func (v *CodeView) indentUnit() string {
	return v.lang.Indent(v.ide.Settings.TabSize)
}

func (v *CodeView) onKey(e key.Event) {
	if e.State != key.Press {
		return
	}
	start, end := v.field.Selection()
	switch e.Name {
	case key.NameTab:
		if e.Modifiers.Contain(key.ModShift) {
			v.apply(textops.Unindent(v.text, start, end, v.indentUnit()))
		} else {
			v.apply(textops.Indent(v.text, start, end, v.indentUnit()))
		}
	case key.NameReturn, key.NameEnter:
		v.apply(textops.NewLine(v.text, start, end, v.indentUnit()))
	}
}

// apply replaces only what an edit changed, so undo stays fine-grained, and
// sets the selection it asks for.
func (v *CodeView) apply(e textops.Edit) {
	s, en, ins := textops.MinimalReplace(v.text, e.Text)
	if s != en || ins != "" {
		v.field.SetSelection(s, en)
		v.field.InsertText(ins)
	}
	if e.Start == e.End {
		v.field.SetCaret(e.Start)
	} else {
		v.field.SetSelection(e.Start, e.End)
	}
}

// DuplicateLine is Ctrl+D.
func (v *CodeView) DuplicateLine() {
	s, e := v.field.Selection()
	v.apply(textops.DuplicateLine(v.text, s, e))
}

// DeleteLine is Ctrl+Y.
func (v *CodeView) DeleteLine() {
	s, e := v.field.Selection()
	v.apply(textops.DeleteLine(v.text, s, e))
}

// CommentLine is Ctrl+/.
func (v *CodeView) CommentLine() {
	s, e := v.field.Selection()
	v.apply(textops.ToggleLineComment(v.text, s, e, v.lang.LineComment))
}

// MoveLine is Alt+Shift+Up/Down.
func (v *CodeView) MoveLine(dir int) {
	s, e := v.field.Selection()
	v.apply(textops.MoveLines(v.text, s, e, dir))
}

// Save writes the file, formatting it first when format is set and a
// formatter applies. done runs on the UI goroutine once it's written.
func (v *CodeView) Save(formatFirst bool, done func()) {
	if !formatFirst || !v.formatAvailable() {
		v.write(done)
		return
	}
	v.Format(func() { v.write(done) })
}

func (v *CodeView) write(done func()) {
	if !v.modified {
		if done != nil {
			done()
		}
		return
	}
	if err := v.doc.Save(v.text); err != nil {
		v.ide.Error(err)
		return
	}
	v.modified = false
	if v.onChange != nil {
		v.onChange()
	}
	if done != nil {
		done()
	}
}

func (v *CodeView) formatOptions() format.Options {
	return format.Options{Prettier: v.ide.Settings.UsePrettier}
}

func (v *CodeView) formatAvailable() bool {
	return v.ide.Project != nil && format.Available(v.ide.Project.Root, v.lang, v.formatOptions())
}

// Format reformats the text (Ctrl+Alt+L), then runs done, also when the
// formatter fails (its error shows in a balloon), so a save still happens.
// Edits made while the formatter ran win: its result is dropped.
func (v *CodeView) Format(done func()) {
	if v.ide.Project == nil {
		return
	}
	root, path, before, l, o := v.ide.Project.Root, v.doc.Path, v.text, v.lang, v.formatOptions()
	go func() {
		out, err := format.Format(context.Background(), root, path, before, l, o)
		v.ide.App.Post(func() {
			switch {
			case err == format.ErrNoFormatter:
			case err != nil:
				v.ide.Error(err)
			case v.text == before && out != before:
				caret := min(v.field.Caret(), utf8.RuneCountInString(out))
				v.apply(textops.Edit{Text: out, Start: caret, End: caret})
			}
			if done != nil {
				done()
			}
		})
	}()
}

// Title is the tab title.
func (v *CodeView) Title() string { return filepath.Base(v.doc.Path) }

// Layout draws the text.
func (v *CodeView) Layout(gtx giode.Context) giode.Dimensions {
	if s := v.ide.Settings; v.applied != [2]int{s.EditorFontSize, s.EditorLineHeight} {
		v.field.Styles(v.styles())
	}
	return v.field.Layout(gtx)
}
