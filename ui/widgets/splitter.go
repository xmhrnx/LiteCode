package widgets

import (
	"github.com/xmhrnx/giode"

	"litecode/ui/theme"
)

// Splitter is the gap between two cards; dragging it resizes them. It reports the size the dragged area should take, from the size
// it had when the drag started plus the pointer's movement (sign flips the
// direction, for areas on the right or at the bottom).
type Splitter struct {
	vertical bool // a vertical line, dragged sideways
	press    *giode.PressableElement
	start    int
}

// NewSplitter creates a splitter. size reads the area's current size, resize
// sets it, and sign is +1 when dragging right/down grows the area.
func NewSplitter(vertical bool, sign int, size func() int, resize func(int), done func()) *Splitter {
	s := &Splitter{vertical: vertical}
	s.press = giode.Pressable(func(st giode.PressState) giode.Element {
		c := theme.Current()
		line := "transparent"
		if st.Hovered || st.Dragging {
			line = c.FocusLine
		}
		// The gap between two cards is the handle; a pill lights up in it
		// while it's hovered or dragged.
		hit := giode.Styles{Align: giode.AlignCenter, Justify: giode.JustifyCenter, Cursor: giode.CursorRowResize}
		bar := giode.Styles{Background: line, BorderRadius: 1}
		if vertical {
			hit.Width, hit.Height, hit.Cursor = theme.CardGap, -1, giode.CursorColResize
			bar.Width, bar.Height = 2, 48
		} else {
			hit.Width, hit.Height = -1, theme.CardGap
			bar.Width, bar.Height = 48, 2
		}
		return giode.Box(hit, giode.Box(bar))
	}).Focusable(false).
		OnDragStart(func(giode.DragEvent) { s.start = size() }).
		OnDragMove(func(e giode.DragEvent) {
			d := e.DeltaDp.Y
			if vertical {
				d = e.DeltaDp.X
			}
			resize(s.start + sign*int(d))
		}).
		OnDragEnd(func(giode.DragEvent) {
			if done != nil {
				done()
			}
		})
	return s
}

// Layout draws the splitter.
func (s *Splitter) Layout(gtx giode.Context) giode.Dimensions {
	return s.press.Layout(gtx)
}
