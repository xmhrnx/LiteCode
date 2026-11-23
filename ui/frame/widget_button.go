// Package frame is LiteCode's custom window frame: the title bar with the main
// menu, the project and branch pills, and the search and settings buttons.
package frame

import (
	"github.com/xmhrnx/giode"

	"litecode/ui/theme"
)

// widgetButton is a title bar pill: something on the left (a badge or an
// icon), a label and a chevron, outlined, filling on hover.
func widgetButton(lead func() giode.Element, label func() string, onClick func(giode.ClickEvent)) *giode.PressableElement {
	return giode.Pressable(func(s giode.PressState) giode.Element {
		c := theme.Current()
		bg := ""
		if s.Hovered || s.Pressed {
			bg = c.Hover
		}
		return giode.Row(giode.Styles{
			Height: 30, Background: bg, BorderRadius: 8, BorderWidth: 1, BorderColor: c.Separator,
			Align: giode.AlignCenter, Gap: 7, Padding: giode.Inset{Left: 6, Right: 8}, Wrap: giode.WrapNoWrap,
		},
			lead(),
			giode.Text(label(), giode.Styles{Color: c.Text, FontSize: theme.UISize, FontWeight: giode.FontWeightMedium, MaxLines: 1}),
			theme.IconSized("chevron-down", c.TextFaint, 14),
		)
	}).Focusable(false).OnClick(onClick)
}
