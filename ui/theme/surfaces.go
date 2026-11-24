package theme

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"
)

// Card radius and the gap between cards, in dp.
const (
	CardRadius = 12
	CardGap    = 8
)

// Backdrop is the gradient behind everything.
func Backdrop() *properties.LinearGradient {
	return giode.Linear(155, Current().Backdrop...)
}

// AccentGradient is the gradient of primary buttons and active items.
func AccentGradient() *properties.LinearGradient {
	c := Current()
	return giode.Linear(135, c.AccentFrom, c.AccentTo)
}

// Card is a floating surface: rounded, edged, with a soft shadow, clipping
// its content to its corners.
func Card(bg string) giode.Styles {
	return giode.Styles{
		Background: bg, BorderRadius: CardRadius, BorderWidth: 1, BorderColor: Current().CardLine,
		BoxShadow: giode.BoxShadow{Y: 6, Blur: 18, Color: Current().Shadow + "/0.55"}, Overflow: giode.OverflowHidden,
	}
}

// GlowShadow is the accent glow under highlighted controls.
func GlowShadow() giode.BoxShadow {
	return giode.BoxShadow{Y: 4, Blur: 14, Color: Current().Glow}
}

// RowWash is the background of a selected list row: an accent wash that
// fades to the right. It's nil for other rows, so their flat Background shows.
func RowWash(selected bool) *properties.LinearGradient {
	if !selected {
		return nil
	}
	c := Current()
	return giode.Linear(90, c.WashFrom, c.WashMid, c.WashTo)
}

// ScrollBar is the thin, auto-hiding scrollbar of every list.
func ScrollBar() giode.ScrollBar {
	return giode.ScrollBar{Width: 6, Radius: 3, TrackColor: "transparent", ThumbColor: Current().ScrollThumb, AutoHide: true, Inset: 2}
}
