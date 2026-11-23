package widgets

import (
	"github.com/xmhrnx/giode"

	"litecode/ui/theme"
)

// SectionTitle is the small gray heading of a group in a popup list.
func SectionTitle(title string) giode.Element {
	return giode.Box(giode.Styles{Width: -1, Padding: giode.Inset{Left: 20, Top: 8, Bottom: 4}},
		giode.Text(title, giode.Styles{Color: theme.Current().TextFaint, FontSize: theme.SmallSize}))
}
