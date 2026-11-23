package search

import (
	"github.com/xmhrnx/giode"

	"litecode/ui/theme"
)

// toggleChip is one of Find in Files' option toggles: "Cc", "W", ".*".
type toggleChip struct {
	on  bool
	tip *giode.TooltipElement
}

func newToggleChip(label, tooltipText string, onChange func()) *toggleChip {
	t := &toggleChip{}
	press := giode.Pressable(func(s giode.PressState) giode.Element {
		c := theme.Current()
		bg, fg := "", c.TextMuted
		switch {
		case t.on:
			bg, fg = c.Selection, c.Text
		case s.Hovered:
			bg = c.Hover
		}
		return giode.Box(giode.Styles{Background: bg, BorderRadius: 4, Padding: giode.SymmetricInset(6, 2)},
			giode.Text(label, giode.Styles{Color: fg, FontSize: theme.SmallSize, FontFamily: theme.CodeFont}))
	}).Focusable(false).OnClick(func(giode.ClickEvent) {
		t.on = !t.on
		onChange()
	})
	t.tip = giode.Tooltip(press, tooltipText)
	return t
}

// Layout draws the chip.
func (t *toggleChip) Layout(gtx giode.Context) giode.Dimensions {
	return t.tip.Layout(gtx)
}
