package widgets

import (
	"sync/atomic"

	"github.com/xmhrnx/giode"
)

// typed counts edits in every text field. Giode doesn't show the app the
// keys a focused field takes, so this is how double Shift tells two Shift
// taps from Shift held for capitals: a real double tap types nothing.
var typed atomic.Uint64

// TextField is giode.Input.Text that counts its edits. Use it for every
// text field; give a field its own change handler with OnChange(Typing(fn)).
func TextField(placeholder string, st ...giode.Styles) *giode.InputElement {
	return giode.Input.Text(placeholder, st...).OnChange(func(string) { typed.Add(1) })
}

// Typing wraps a field's change handler so the edit is still counted.
func Typing(fn func(string)) func(string) {
	return func(s string) {
		typed.Add(1)
		fn(s)
	}
}

// TypedCount is the number of edits made in any text field so far.
func TypedCount() uint64 { return typed.Load() }

// NoteTyping counts text typed somewhere other than a text field (the
// code editor), so double Shift isn't taken for two capitals.
func NoteTyping() { typed.Add(1) }
