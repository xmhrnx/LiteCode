package textbuf

import "time"

// mergeWindow is how close typed edits must come to share an undo step.
const mergeWindow = 1500 * time.Millisecond

// Selection is an anchor and a caret; the caret is where the cursor shows.
type Selection struct {
	Anchor, Caret Pos
}

// Range is the selected text.
func (s Selection) Range() Range { return Ordered(s.Anchor, s.Caret) }

// Change is one replacement: Removed was at At, Inserted took its place.
type Change struct {
	At       Pos
	Removed  string
	Inserted string
}

// Step is what one undo reverts: changes in order, with the selection
// before and after.
type Step struct {
	Changes []Change
	Before  Selection
	After   Selection
	kind    Kind
	at      time.Time
}

// Kind says whether an edit may merge with the one before it.
type Kind int

// The kinds of edit. Typing merges with typing, deleting with deleting.
const (
	KindOther Kind = iota
	KindTyping
	KindDeleting
)

// History is the undo and redo stacks.
type History struct {
	undo, redo []Step
	// sealed stops the next edit merging into the last step (after the
	// caret moves on its own).
	sealed bool
}

// Seal makes the next edit a new undo step.
func (h *History) Seal() { h.sealed = true }

// Record adds a step, merging it into the last one when both are typing
// (or deleting) in one place in quick succession.
func (h *History) Record(s Step, kind Kind) {
	h.redo = h.redo[:0]
	s.kind, s.at = kind, time.Now()
	if n := len(h.undo); n > 0 && !h.sealed && kind != KindOther {
		last := &h.undo[n-1]
		if last.kind == kind && s.at.Sub(last.at) < mergeWindow && last.After == s.Before {
			last.Changes = append(last.Changes, s.Changes...)
			last.After, last.at = s.After, s.at
			return
		}
	}
	h.sealed = false
	h.undo = append(h.undo, s)
	if len(h.undo) > 2000 {
		h.undo = h.undo[len(h.undo)-2000:]
	}
}

// Undo pops the last step, for the buffer to revert.
func (h *History) Undo() (Step, bool) {
	n := len(h.undo)
	if n == 0 {
		return Step{}, false
	}
	s := h.undo[n-1]
	h.undo = h.undo[:n-1]
	h.redo = append(h.redo, s)
	h.sealed = true
	return s, true
}

// Redo pops the last undone step, for the buffer to apply again.
func (h *History) Redo() (Step, bool) {
	n := len(h.redo)
	if n == 0 {
		return Step{}, false
	}
	s := h.redo[n-1]
	h.redo = h.redo[:n-1]
	h.undo = append(h.undo, s)
	h.sealed = true
	return s, true
}

// end is where text inserted at p ends.
func end(p Pos, text string) Pos {
	for _, r := range text {
		if r == '\n' {
			p.Line++
			p.Col = 0
		} else {
			p.Col++
		}
	}
	return p
}

// Apply makes a change in the buffer.
func (b *Buffer) Apply(c Change) {
	b.Replace(Range{c.At, end(c.At, c.Removed)}, c.Inserted)
}

// Revert takes a change back out of the buffer.
func (b *Buffer) Revert(c Change) {
	b.Replace(Range{c.At, end(c.At, c.Inserted)}, c.Removed)
}
