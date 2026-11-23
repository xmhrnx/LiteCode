package settings

import (
	"github.com/xmhrnx/giode"

	"litecode/ui/theme"
)

// settingRow lays out a label, a control and an optional comment under them.
func settingRow(label string, control giode.Element, comment string) giode.Element {
	c := theme.Current()
	kids := []giode.Element{
		giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{Width: 180}, giode.Text(label, giode.Styles{Color: c.Text, FontSize: theme.UISize})), control),
	}
	if comment != "" {
		kids = append(kids, giode.Box(giode.Styles{Padding: giode.Inset{Left: 192}},
			giode.Text(comment, giode.Styles{Color: c.TextFaint, FontSize: theme.SmallSize})))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 4}, kids...)
}

// pageTitle is a settings page's heading.
func pageTitle(title string) giode.Element {
	return giode.Text(title, giode.Styles{Color: theme.Current().Text, FontSize: 14, FontWeight: giode.FontWeightSemiBold})
}
