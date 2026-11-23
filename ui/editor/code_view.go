// Package editor is the editor area: the tabs of open files, the text of the
// selected one, and the status bar that describes it.
package editor

import (
	"context"
	"path/filepath"

	"github.com/xmhrnx/giode"

	"litecode/internal/format"
	"litecode/internal/lang"
	"litecode/internal/textbuf"
	"litecode/internal/textops"
	"litecode/internal/workspace"
	"litecode/ui/core"
)

// CodeView is one open file: its document and the editor showing it.
type CodeView struct {
	ide      *core.IDE
	doc      *workspace.Document
	lang     *lang.Language
	editor   *TextEditor
	modified bool
	onChange func()
}

// NewCodeView opens a document in a new view; onChange runs after every edit.
func NewCodeView(ide *core.IDE, doc *workspace.Document, onChange func()) *CodeView {
	v := &CodeView{ide: ide, doc: doc, lang: ide.Langs.ForPath(doc.Path), onChange: onChange}
	v.editor = NewTextEditor(ide, v.lang, doc.Saved, v.changed)
	return v
}

// Path is the file's path.
func (v *CodeView) Path() string { return v.doc.Path }

// Language is the file's language.
func (v *CodeView) Language() *lang.Language { return v.lang }

// Modified reports unsaved changes.
func (v *CodeView) Modified() bool { return v.modified }

// Text is the current text.
func (v *CodeView) Text() string { return v.editor.Text() }

// Focus gives the view the keyboard.
func (v *CodeView) Focus() { v.editor.Focus() }

// Caret returns the caret's 1-based line and column.
func (v *CodeView) Caret() (line, col int) {
	c := v.editor.Caret()
	return c.Line + 1, c.Col + 1
}

// GoTo puts the caret at a 1-based line and column.
func (v *CodeView) GoTo(line, col int) {
	v.editor.GoTo(textbuf.Pos{Line: line - 1, Col: col - 1})
	v.editor.Focus()
}

func (v *CodeView) changed() {
	v.modified = v.doc.Modified(v.editor.Text())
	if v.onChange != nil {
		v.onChange()
	}
}

// line runs a line action on the text and the selection.
func (v *CodeView) line(action func(text string, start, end int) textops.Edit) {
	s, e := v.editor.SelectionOffsets()
	v.editor.applyWhole(action(v.editor.Text(), s, e))
}

// DuplicateLine is Ctrl+D.
func (v *CodeView) DuplicateLine() { v.line(textops.DuplicateLine) }

// DeleteLine is Ctrl+Y.
func (v *CodeView) DeleteLine() { v.line(textops.DeleteLine) }

// CommentLine is Ctrl+/.
func (v *CodeView) CommentLine() {
	v.line(func(text string, s, e int) textops.Edit {
		return textops.ToggleLineComment(text, s, e, v.lang.LineComment)
	})
}

// MoveLine is Alt+Shift+Up/Down.
func (v *CodeView) MoveLine(dir int) {
	v.line(func(text string, s, e int) textops.Edit { return textops.MoveLines(text, s, e, dir) })
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
	if err := v.doc.Save(v.editor.Text()); err != nil {
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
	root, path, before, l, o := v.ide.Project.Root, v.doc.Path, v.editor.Text(), v.lang, v.formatOptions()
	go func() {
		out, err := format.Format(context.Background(), root, path, before, l, o)
		v.ide.App.Post(func() {
			switch {
			case err == format.ErrNoFormatter:
			case err != nil:
				v.ide.Error(err)
			case v.editor.Text() == before && out != before:
				v.editor.SetText(out)
			}
			if done != nil {
				done()
			}
		})
	}()
}

// Title is the tab title.
func (v *CodeView) Title() string { return filepath.Base(v.doc.Path) }

// Layout draws the editor.
func (v *CodeView) Layout(gtx giode.Context) giode.Dimensions {
	return v.editor.Layout(gtx)
}
