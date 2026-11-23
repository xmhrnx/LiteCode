package frame

import (
	"github.com/xmhrnx/giode"

	"litecode/internal/keymap"
	"litecode/ui/core"
	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// SearchPill is the command bar in the title bar: it looks like a search
// field and opens Search Anything.
type SearchPill struct {
	press *giode.PressableElement
}

// NewSearchPill creates the pill.
func NewSearchPill(ide *core.IDE) *SearchPill {
	p := &SearchPill{}
	p.press = giode.Pressable(func(s giode.PressState) giode.Element {
		c := theme.Current()
		st := giode.Styles{
			Width: 320, Height: 30, BorderRadius: 15, BorderWidth: 1, BorderColor: c.PopupLine, Background: c.Card,
			Align: giode.AlignCenter, Gap: 8, Padding: giode.Inset{Left: 12, Right: 6}, Wrap: giode.WrapNoWrap,
		}
		if s.Hovered {
			st.BorderColor, st.Background = c.FocusLine, c.Raised
		}
		return giode.Row(st,
			theme.IconSized("search", c.TextMuted, 15),
			giode.Text("Search files and actions", giode.Styles{Color: c.TextFaint, FontSize: theme.UISize, FlexGrow: 1, MaxLines: 1}),
			widgets.Keycaps(ide.Shortcut(keymap.SearchEverywhere)),
		)
	}).Label("Search Anything").Focusable(false).OnPrimaryClick(func(giode.ClickEvent) { ide.Run(keymap.SearchEverywhere) })
	return p
}

// Layout draws the pill.
func (p *SearchPill) Layout(gtx giode.Context) giode.Dimensions {
	return p.press.Layout(gtx)
}
