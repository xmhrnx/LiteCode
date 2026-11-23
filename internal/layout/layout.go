// Package layout is the model of the tool window arrangement: which side each
// tool window docks to, which one is open on each side, and how big each side
// is. The user's arrangement is saved globally, so a terminal moved to the
// right stays on the right in every project.
package layout

import (
	"sort"

	"litecode/internal/config"
)

// Side is where a tool window docks.
type Side string

const (
	Left   Side = "left"
	Right  Side = "right"
	Bottom Side = "bottom"
)

// Sides in the order stripes are drawn.
var Sides = []Side{Left, Bottom, Right}

// WindowID names a tool window.
type WindowID string

// Placement is where one tool window lives.
type Placement struct {
	Side  Side `json:"side"`
	Order int  `json:"order"`
}

// Layout is the whole arrangement.
type Layout struct {
	Windows map[WindowID]Placement `json:"windows"`
	Open    map[Side]WindowID      `json:"open"`
	Sizes   map[Side]int           `json:"sizes"` // dp: width for left/right, height for bottom
	stash   map[Side]WindowID      // what Hide All Tool Windows closed
}

// Size limits in dp.
const (
	MinSize = 120
	MaxSize = 1400
)

// New returns the default arrangement for the given windows.
func New(defaults map[WindowID]Placement) *Layout {
	l := &Layout{Windows: map[WindowID]Placement{}, Open: map[Side]WindowID{}, Sizes: map[Side]int{Left: 300, Right: 340, Bottom: 260}}
	for id, p := range defaults {
		l.Windows[id] = p
	}
	return l
}

// Load reads the saved arrangement over the defaults. Windows the file doesn't
// know keep their default place; windows that no longer exist are dropped.
func Load(defaults map[WindowID]Placement) (*Layout, error) {
	l := New(defaults)
	path, err := config.Path("layout.json")
	if err != nil {
		return l, err
	}
	var saved Layout
	if err := config.LoadJSON(path, &saved); err != nil {
		return l, err
	}
	for id, p := range saved.Windows {
		if _, ok := defaults[id]; ok && validSide(p.Side) {
			l.Windows[id] = p
		}
	}
	for side, id := range saved.Open {
		if p, ok := l.Windows[id]; ok && p.Side == side {
			l.Open[side] = id
		}
	}
	for side, n := range saved.Sizes {
		if validSide(side) {
			l.Sizes[side] = clamp(n)
		}
	}
	return l, nil
}

// Save writes the arrangement.
func (l *Layout) Save() error {
	path, err := config.Path("layout.json")
	if err != nil {
		return err
	}
	return config.SaveJSON(path, l)
}

// On lists the windows docked to a side, in stripe order.
func (l *Layout) On(side Side) []WindowID {
	var ids []WindowID
	for id, p := range l.Windows {
		if p.Side == side {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool {
		a, b := l.Windows[ids[i]], l.Windows[ids[j]]
		if a.Order != b.Order {
			return a.Order < b.Order
		}
		return ids[i] < ids[j]
	})
	return ids
}

// IsOpen reports whether a window is showing.
func (l *Layout) IsOpen(id WindowID) bool {
	p, ok := l.Windows[id]
	return ok && l.Open[p.Side] == id
}

// Toggle opens a window (closing the other one on its side) or closes it.
func (l *Layout) Toggle(id WindowID) {
	if l.IsOpen(id) {
		l.Close(id)
		return
	}
	l.Show(id)
}

// Show opens a window.
func (l *Layout) Show(id WindowID) {
	if p, ok := l.Windows[id]; ok {
		l.Open[p.Side] = id
		l.stash = nil
	}
}

// Close closes a window if it's open.
func (l *Layout) Close(id WindowID) {
	if p, ok := l.Windows[id]; ok && l.Open[p.Side] == id {
		delete(l.Open, p.Side)
	}
}

// Move docks a window to another side, at the end of its stripe, keeping it
// open if it was.
func (l *Layout) Move(id WindowID, side Side) {
	p, ok := l.Windows[id]
	if !ok || p.Side == side {
		return
	}
	wasOpen := l.IsOpen(id)
	l.Close(id)
	order := 0
	for _, other := range l.On(side) {
		order = max(order, l.Windows[other].Order+1)
	}
	l.Windows[id] = Placement{Side: side, Order: order}
	if wasOpen {
		l.Open[side] = id
	}
}

// HideAll closes every window, or brings back the ones it closed last time
// (Ctrl+Shift+F12).
func (l *Layout) HideAll() {
	if len(l.Open) == 0 && l.stash != nil {
		l.Open, l.stash = l.stash, nil
		return
	}
	l.stash, l.Open = l.Open, map[Side]WindowID{}
}

// Resize sets a side's size in dp, within MinSize and MaxSize.
func (l *Layout) Resize(side Side, dp int) {
	l.Sizes[side] = clamp(dp)
}

func clamp(n int) int {
	return min(max(n, MinSize), MaxSize)
}

func validSide(s Side) bool {
	return s == Left || s == Right || s == Bottom
}
