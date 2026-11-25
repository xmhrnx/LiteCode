package editor

import (
	"slices"
	"strings"
	"unicode"
	"unicode/utf16"

	"gioui.org/io/key"

	"litecode/internal/complete"
	"litecode/internal/highlight"
	"litecode/internal/lsp"
	"litecode/internal/textbuf"
)

// popupRows is how many completions show at once.
const popupRows = 10

// completionSource is where completions come from besides the text's
// own words: the file's language server.
type completionSource interface {
	ServerComplete(line, char16 int, trigger string, done func(items []lsp.Completion, ok bool))
	ServerResolve(c lsp.Completion, done func(lsp.Completion))
	ServerTriggers() []string
	HasServer() bool
}

// completion is an open completion list.
type completion struct {
	start   textbuf.Pos      // where the word being completed starts
	items   []lsp.Completion // the candidates: the server's, or words
	served  bool             // items came from the server
	shown   []complete.Item  // the candidates matching what's typed, best first
	sel     int              // the selected one in shown
	first   int              // the first one in view
	req     int              // the request whose answer is wanted
	waiting bool
	trigger string
	member  bool // after "." or the like: only the server knows what fits
}

// col16 is a rune column as UTF-16 units, the language servers' count.
func col16(line []rune, col int) int {
	n := 0
	for _, r := range line[:min(col, len(line))] {
		n += max(utf16.RuneLen(r), 1)
	}
	return n
}

// posFrom16 is a server position as a buffer position.
func (e *TextEditor) posFrom16(p lsp.Position) textbuf.Pos {
	if p.Line >= e.buf.Lines() {
		return e.buf.End()
	}
	line := e.buf.Line(max(p.Line, 0))
	n, col := 0, 0
	for col < len(line) && n < p.Character {
		n += max(utf16.RuneLen(line[col]), 1)
		col++
	}
	return textbuf.Pos{Line: max(p.Line, 0), Col: col}
}

// wordStart is where the identifier ending at p starts.
func (e *TextEditor) wordStart(p textbuf.Pos) textbuf.Pos {
	line := e.buf.Line(p.Line)
	c := p.Col
	for c > 0 && complete.IsWordRune(line[c-1]) {
		c--
	}
	return textbuf.Pos{Line: p.Line, Col: c}
}

// inCommentOrString reports whether the caret is in a comment or a string,
// where completions don't open by themselves.
func (e *TextEditor) inCommentOrString() bool {
	c := e.sel.Caret
	for _, t := range e.hl.Tokens(e.buf, c.Line) {
		if t.Start < c.Col && c.Col <= t.End {
			return t.Kind == highlight.Comment || t.Kind == highlight.String
		}
	}
	return false
}

// isTrigger reports whether typed text asks the server for completions
// ("." after a value, "::" in Rust).
func (e *TextEditor) isTrigger(s string) bool {
	if e.source == nil {
		return false
	}
	for _, t := range e.source.ServerTriggers() {
		if t == s || len(t) > 1 && strings.HasSuffix(string(e.buf.Line(e.sel.Caret.Line)[:e.sel.Caret.Col]), t) && strings.HasSuffix(t, s) {
			return true
		}
	}
	return false
}

// afterTyping opens, narrows or closes the list after text was typed.
func (e *TextEditor) afterTyping(s string) {
	r := []rune(s)
	if len(r) != 1 {
		e.closeCompletion()
		return
	}
	switch {
	case e.isTrigger(s):
		e.openCompletion(s)
	case complete.IsWordRune(r[0]) && e.comp != nil:
		e.refilter()
		e.requestCompletions()
	case complete.IsWordRune(r[0]) && !unicode.IsDigit(e.buf.Line(e.sel.Caret.Line)[e.wordStart(e.sel.Caret).Col]) && !e.inCommentOrString():
		e.openCompletion("")
	default:
		e.closeCompletion()
	}
}

// afterMember reports whether the word at start follows a member access
// ("json.", "self.", "a->", "std::"), where the text's words would be
// guesses.
func (e *TextEditor) afterMember(start textbuf.Pos) bool {
	before := strings.TrimRight(string(e.buf.Line(start.Line)[:start.Col]), " \t")
	return strings.HasSuffix(before, ".") || strings.HasSuffix(before, "->") || strings.HasSuffix(before, "::")
}

// openCompletion opens the list at the caret. With a language server it
// shows the server's completions, waiting for them; without one, the
// text's words and the language's keywords, but nothing after a member
// access, where those would be wrong.
func (e *TextEditor) openCompletion(trigger string) {
	start := e.wordStart(e.sel.Caret)
	if trigger != "" {
		start = e.sel.Caret
	}
	server := e.source != nil && e.source.HasServer()
	member := trigger != "" || e.afterMember(start)
	if member && !server {
		e.closeCompletion()
		return
	}
	e.comp = &completion{start: start, trigger: trigger, member: member, waiting: server}
	if !server {
		e.comp.items = e.wordItems(start)
	}
	e.refilter()
	e.requestCompletions()
}

// Complete opens the completion list from the keyboard (Ctrl+Space).
func (e *TextEditor) Complete() {
	e.openCompletion("")
}

// wordItems are the text's words and the language's keywords.
func (e *TextEditor) wordItems(start textbuf.Pos) []lsp.Completion {
	typing := string(e.buf.Line(start.Line)[start.Col:e.sel.Caret.Col])
	end := start.Col
	line := e.buf.Line(start.Line)
	for end < len(line) && complete.IsWordRune(line[end]) {
		end++
	}
	current := string(line[start.Col:end])
	var items []lsp.Completion
	seen := map[string]bool{current: true, typing: true}
	for _, w := range highlight.For(e.lang.ID, e.lang.LineComment, e.lang.BlockComment, e.lang.Keywords).Words() {
		if !seen[w] {
			seen[w] = true
			items = append(items, lsp.Completion{Label: w, Insert: w, Filter: w, Kind: lsp.KindKeyword, SortText: "1"})
		}
	}
	for _, w := range complete.Words(e.buf.Line, e.buf.Lines(), start.Line, current) {
		if !seen[w] {
			seen[w] = true
			items = append(items, lsp.Completion{Label: w, Insert: w, Filter: w, Kind: lsp.KindText, SortText: "2"})
		}
	}
	return items
}

// requestCompletions asks the server for the list at the caret; its
// answer replaces the words when it has any.
func (e *TextEditor) requestCompletions() {
	if e.source == nil || e.comp == nil || !e.source.HasServer() {
		return
	}
	e.reqSeq++
	id, c := e.reqSeq, e.sel.Caret
	e.comp.req, e.comp.waiting = id, true
	e.source.ServerComplete(c.Line, col16(e.buf.Line(c.Line), c.Col), e.comp.trigger, func(items []lsp.Completion, ok bool) {
		if e.comp == nil || e.comp.req != id {
			return
		}
		e.comp.waiting = false
		switch {
		case ok && len(items) > 0:
			e.comp.items, e.comp.served = items, true
		case !e.comp.member && !e.comp.served:
			// The server had nothing: the text's words, as without one.
			e.comp.items = e.wordItems(e.comp.start)
		}
		e.refilter()
		e.ide.App.Invalidate()
	})
}

// refilter ranks the candidates against what's typed since the start,
// closing the list when the caret left the word or nothing matches.
func (e *TextEditor) refilter() {
	c := e.comp
	caret := e.sel.Caret
	if caret.Line != c.start.Line || caret.Col < c.start.Col {
		e.closeCompletion()
		return
	}
	typed := string(e.buf.Line(caret.Line)[c.start.Col:caret.Col])
	items := make([]complete.Item, len(c.items))
	for i, it := range c.items {
		items[i] = complete.Item{Label: it.Label, Filter: it.Filter, Sort: it.SortText, Index: i}
	}
	c.shown = complete.Rank(items, typed)
	c.sel, c.first = 0, 0
	if len(c.shown) == 0 && !c.waiting {
		e.closeCompletion()
	}
}

// closeCompletion closes the list.
func (e *TextEditor) closeCompletion() { e.comp = nil }

// moveSelection moves the list's selection by n rows, keeping it in view.
func (e *TextEditor) moveSelection(n int) {
	c := e.comp
	if len(c.shown) == 0 {
		return
	}
	switch {
	case n == 1 && c.sel == len(c.shown)-1:
		c.sel = 0
	case n == -1 && c.sel == 0:
		c.sel = len(c.shown) - 1
	default:
		c.sel = max(0, min(len(c.shown)-1, c.sel+n))
	}
	if c.sel < c.first {
		c.first = c.sel
	}
	if c.sel >= c.first+popupRows {
		c.first = c.sel - popupRows + 1
	}
}

// completionKey handles keys while the list is open: moving in it,
// taking a completion, closing it. It reports whether it used the key.
func (e *TextEditor) completionKey(name key.Name) bool {
	switch name {
	case key.NameUpArrow:
		e.moveSelection(-1)
	case key.NameDownArrow:
		e.moveSelection(1)
	case key.NamePageUp:
		e.moveSelection(-popupRows + 1)
	case key.NamePageDown:
		e.moveSelection(popupRows - 1)
	case key.NameReturn, key.NameEnter, key.NameTab:
		if len(e.comp.shown) == 0 {
			e.closeCompletion()
			return false
		}
		e.accept(e.comp.sel)
	case key.NameEscape:
		e.closeCompletion()
	default:
		return false
	}
	return true
}

// accept puts the shown completion i in: it replaces the word typed so
// far, and makes its extra edits (an import) in the same undo step. When
// the server sends those only on request, they follow when it answers.
func (e *TextEditor) accept(i int) {
	c := e.comp
	if c == nil || i >= len(c.shown) {
		return
	}
	it := c.items[c.shown[i].Index]
	e.closeCompletion()
	caret := e.sel.Caret
	start, text := c.start, it.Insert
	if it.Edit != nil {
		start, text = e.posFrom16(it.Edit.Range.Start), it.Edit.NewText
		if caret.Less(start) {
			start = c.start
		}
	}
	at := len([]rune(text))
	if it.Snippet {
		text, at = plainSnippet(text)
	}
	e.applyEdits(textbuf.Range{Start: start, End: caret}, text, at, it.Extra)
	if len(it.Extra) == 0 && c.served && e.source != nil {
		rev := e.buf.Rev()
		e.source.ServerResolve(it, func(r lsp.Completion) {
			if len(r.Extra) == 0 || e.buf.Rev() != rev {
				return
			}
			// Positions in the text before the completion went in:
			// only the ones before it (imports at the top) still hold.
			var before []lsp.TextEdit
			for _, x := range r.Extra {
				if p := e.posFrom16(x.Range.End); !start.Less(p) {
					before = append(before, x)
				}
			}
			if len(before) > 0 {
				e.hist.Seal()
				e.applyEdits(textbuf.Range{Start: e.sel.Caret, End: e.sel.Caret}, "", 0, before)
			}
		})
	}
}

// applyEdits replaces main with text, puts the caret at rune at of it,
// and makes the extra edits too, all as one undo step.
func (e *TextEditor) applyEdits(main textbuf.Range, text string, at int, extra []lsp.TextEdit) {
	type change struct {
		r    textbuf.Range
		text string
	}
	changes := []change{{main, text}}
	caret := e.buf.Offset(main.Start) + at
	for _, x := range extra {
		r := textbuf.Range{Start: e.posFrom16(x.Range.Start), End: e.posFrom16(x.Range.End)}
		if !main.Start.Less(r.End) {
			// Before the completion: it moves the caret.
			caret += len([]rune(x.NewText)) - (e.buf.Offset(r.End) - e.buf.Offset(r.Start))
		}
		changes = append(changes, change{r, x.NewText})
	}
	// From the end back, so earlier positions stay right.
	slices.SortStableFunc(changes, func(a, b change) int {
		switch {
		case b.r.Start.Less(a.r.Start):
			return -1
		case a.r.Start.Less(b.r.Start):
			return 1
		}
		return 0
	})
	t := e.begin()
	for _, ch := range changes {
		t.replace(ch.r, ch.text)
	}
	e.sel = caretAt(e.buf.PosOf(caret))
	t.commit(textbuf.KindOther)
}
