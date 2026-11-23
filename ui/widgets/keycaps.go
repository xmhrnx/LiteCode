package widgets

import (
	"strings"

	"github.com/xmhrnx/giode"

	"litecode/ui/theme"
)

// Keycaps draws a shortcut ("Ctrl+Shift+N", "Double ⇧") as key caps.
func Keycaps(shortcut string) giode.Element {
	c := theme.Current()
	var keys []giode.Element
	parts := strings.Split(shortcut, "+")
	if strings.HasPrefix(shortcut, "Double ") {
		parts = []string{"⇧", "⇧"}
	}
	for _, k := range parts {
		keys = append(keys, giode.Box(giode.Styles{
			MinWidth: 22, Height: 22, Padding: giode.SymmetricInset(6, 0), Align: giode.AlignCenter, Justify: giode.JustifyCenter,
			Background: c.Raised, BorderRadius: 5, BorderWidth: 1, BorderColor: c.PopupLine,
		}, giode.Text(k, giode.Styles{Color: c.TextMuted, FontSize: 11, FontWeight: giode.FontWeightMedium})))
	}
	return giode.Row(giode.Styles{Gap: 4, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, keys...)
}
