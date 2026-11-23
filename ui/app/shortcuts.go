package app

import (
	"time"

	"gioui.org/io/key"
	"github.com/xmhrnx/giode"

	"litecode/internal/keymap"
	"litecode/ui/core"
	"litecode/ui/widgets"
)

// The second Shift of a double Shift comes between these after the first is
// released. Faster than a person taps is a capital letter's Shift.
const (
	doubleShiftWindow = 350 * time.Millisecond
	doubleShiftMinGap = 40 * time.Millisecond
)

// Shortcuts registers the keymap's strokes every frame, rebuilding its
// table only when the keymap changes, and detects double Shift.
type Shortcuts struct {
	ide       *core.IDE
	table     []giode.KeyShortcut
	rev       int
	built     bool
	lastShift time.Time // when Shift was last released with nothing else pressed
	typingRev uint64    // widgets.TypedCount when that tap began
	shiftDown bool
	pressRev  uint64 // widgets.TypedCount when Shift last went down
	dirty     bool   // another key was pressed while Shift was down
}

// NewShortcuts creates the shortcut table for the IDE's keymap.
func NewShortcuts(ide *core.IDE) *Shortcuts {
	return &Shortcuts{ide: ide}
}

func (s *Shortcuts) rebuild() {
	s.table = s.table[:0]
	for _, a := range keymap.Actions() {
		a := a
		for _, stroke := range s.ide.Keymap.Bindings[a] {
			if stroke == keymap.DoubleShift {
				continue
			}
			s.table = append(s.table, giode.Shortcut(stroke, func() { s.ide.Run(a) }))
		}
	}
	s.rev, s.built = s.ide.Keymap.Rev, true
}

// doubleShiftAction is the action bound to double Shift, if any.
func (s *Shortcuts) doubleShiftAction() keymap.Action {
	for a, strokes := range s.ide.Keymap.Bindings {
		for _, st := range strokes {
			if st == keymap.DoubleShift {
				return a
			}
		}
	}
	return ""
}

// watchShift sees every key, whatever took it, and fires on Shift, Shift.
func (s *Shortcuts) watchShift(e key.Event) {
	if e.Name != key.NameShift {
		if e.State == key.Press {
			s.lastShift, s.dirty = time.Time{}, s.shiftDown
		}
		return
	}
	now := time.Now()
	if e.State == key.Press {
		s.shiftDown, s.pressRev = true, widgets.TypedCount()
		gap := now.Sub(s.lastShift)
		if !s.lastShift.IsZero() && gap >= doubleShiftMinGap && gap < doubleShiftWindow && s.typingRev == widgets.TypedCount() {
			s.lastShift = time.Time{}
			if a := s.doubleShiftAction(); a != "" {
				s.ide.Run(a)
			}
			return
		}
		return
	}
	s.shiftDown = false
	if !s.dirty {
		// Compared at the next press: anything typed from this press on
		// (a capital letter) means it wasn't a bare tap.
		s.lastShift, s.typingRev = now, s.pressRev
	} else {
		s.lastShift = time.Time{}
	}
	s.dirty = false
}

// Register installs the shortcuts for this frame. Call it from the view.
func (s *Shortcuts) Register() {
	if s.ide.CapturingKeys {
		return
	}
	if !s.built || s.rev != s.ide.Keymap.Rev {
		s.rebuild()
	}
	giode.Shortcuts(s.table...)
	giode.WatchKeys(s.watchShift)
}
