package gitui

import (
	"github.com/xmhrnx/giode"

	"litecode/ui/theme"
)

// commitRail draws a commit's place on the history line: a ring on a
// vertical rail, open at the ends of the list. Head commits glow.
func commitRail(first, last, head bool) giode.Element {
	c := theme.Current()
	line := giode.Styles{Position: giode.PositionAbsolute, Left: 6, Width: 2, Background: "#3A3466"}
	if first {
		line.Top = 12
	} else {
		line.Top = giode.Flush
	}
	if last {
		line.Bottom = 12
	} else {
		line.Bottom = giode.Flush
	}
	ring := giode.Styles{ZIndex: 2, Width: 10, Height: 10, BorderRadius: 5, BorderWidth: 2, BorderColor: c.TextAccent, Background: c.Editor}
	if head {
		ring.BackgroundGradient, ring.BorderColor, ring.BoxShadow = theme.AccentGradient(), c.AccentFrom, theme.GlowShadow()
	}
	return giode.Box(giode.Styles{Width: 14, Height: -1, Align: giode.AlignCenter, Justify: giode.JustifyCenter},
		giode.Box(line), giode.Box(ring))
}
