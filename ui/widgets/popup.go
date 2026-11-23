package widgets

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/elements"
	"github.com/xmhrnx/giode/styles/properties"
)

// Popup is a floating panel near the top of the window, centered, as Search
// Everywhere and Find in Files are. Escape or a click outside closes it.
// Create it once and put Element() among the window root's children.
type Popup struct {
	top     int
	content func() giode.Element
	open    bool
	onClose func()
	scrim   *giode.PressableElement
	panel   *giode.PressableElement
}

// NewPopup creates a closed popup whose panel sits top dp below the window's top.
func NewPopup(top int, content func() giode.Element) *Popup {
	p := &Popup{top: top, content: content}
	p.scrim = giode.Pressable(func(giode.PressState) giode.Element {
		return giode.Box(giode.Styles{Width: -1, Height: -1, Background: "#0000004D"})
	}, giode.Styles{Width: -1, Height: -1, Cursor: giode.CursorDefault}).Focusable(false).
		OnClick(func(giode.ClickEvent) { p.Close() })
	// The panel takes presses on its own background so they don't reach the scrim.
	p.panel = giode.Pressable(func(giode.PressState) giode.Element { return p.content() },
		giode.Styles{Cursor: giode.CursorDefault}).Focusable(false)
	return p
}

// OnClose runs fn whenever the popup closes.
func (p *Popup) OnClose(fn func()) *Popup {
	p.onClose = fn
	return p
}

// Open shows the popup.
func (p *Popup) Open() { p.open = true }

// IsOpen reports whether the popup shows.
func (p *Popup) IsOpen() bool { return p.open }

// Close hides the popup.
func (p *Popup) Close() {
	if !p.open {
		return
	}
	p.open = false
	if p.onClose != nil {
		p.onClose()
	}
}

// Element is the popup layer: an absolutely placed box covering the window
// while open, nothing while closed.
func (p *Popup) Element() giode.Element {
	if !p.open {
		return giode.Box(giode.Styles{Display: giode.DisplayNone})
	}
	return giode.Box(giode.Styles{
		Position: giode.PositionAbsolute, Top: giode.Flush, Left: giode.Flush, Right: giode.Flush, Bottom: giode.Flush,
		ZIndex: 50,
	},
		giode.Raw(func(gtx giode.Context) giode.Dimensions {
			elements.HandleDismiss(p.Close)
			return giode.Dimensions{}
		}),
		giode.Box(giode.Styles{Position: giode.PositionAbsolute, Top: giode.Flush, Left: giode.Flush, Right: giode.Flush, Bottom: giode.Flush}, p.scrim),
		giode.Box(giode.Styles{
			Position: giode.PositionAbsolute, Top: properties.Top(p.top), Left: giode.Flush, Right: giode.Flush,
			Justify: giode.JustifyCenter, ZIndex: 1,
		}, p.panel),
	)
}
