// Package widgets holds small components several parts of the IDE share.
package widgets

import (
	"github.com/xmhrnx/giode"

	"litecode/ui/theme"
)

// ToolButton is a 24dp square icon button with a rounded hover, as in tool
// window headers and the title bar. Create it once.
type ToolButton struct {
	icon    string
	toggled func() bool
	press   *giode.PressableElement
	tip     *giode.TooltipElement
}

// NewToolButton makes a button showing a registered icon; tooltip may be "".
func NewToolButton(icon, tooltip string, onClick func()) *ToolButton {
	return NewToolButtonEvent(icon, tooltip, func(giode.ClickEvent) { onClick() })
}

// NewToolButtonEvent is NewToolButton with the click passed on, to open a
// menu where it happened.
func NewToolButtonEvent(icon, tooltip string, onClick func(e giode.ClickEvent)) *ToolButton {
	b := &ToolButton{icon: icon}
	b.press = giode.Pressable(func(s giode.PressState) giode.Element {
		c := theme.Current()
		bg := ""
		switch {
		case s.Pressed || (b.toggled != nil && b.toggled()):
			bg = c.Pressed
		case s.Hovered:
			bg = c.Hover
		}
		return giode.Box(giode.Styles{Width: 26, Height: 26, BorderRadius: 5, Background: bg, Align: giode.AlignCenter, Justify: giode.JustifyCenter},
			theme.Icon(b.icon, c.TextMuted))
	}).Label(tooltip).OnPrimaryClick(onClick)
	if tooltip != "" {
		b.tip = giode.Tooltip(b.press, tooltip)
	}
	return b
}

// Toggled shows the button pressed while fn reports true.
func (b *ToolButton) Toggled(fn func() bool) *ToolButton {
	b.toggled = fn
	return b
}

// Layout draws the button.
func (b *ToolButton) Layout(gtx giode.Context) giode.Dimensions {
	if b.tip != nil {
		return b.tip.Layout(gtx)
	}
	return b.press.Layout(gtx)
}
