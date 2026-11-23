package settings

import (
	"github.com/xmhrnx/giode"

	"litecode/internal/llm"
	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// aiProviderRow is one provider in the AI page's list: its icon, name, and
// whether it's active or ready to use.
type aiProviderRow struct {
	p        llm.Provider
	selected bool
	active   bool
	ready    bool // a key is saved, or none is needed
	press    *giode.PressableElement
}

func newAIProviderRow(onClick func(id string)) *aiProviderRow {
	r := &aiProviderRow{}
	r.press = giode.Pressable(r.view).Focusable(false).OnClick(widgets.PrimaryClick(func(giode.ClickEvent) { onClick(r.p.ID) }))
	return r
}

func kindIcon(k llm.Kind) string {
	switch k {
	case llm.Local:
		return "server"
	case llm.Custom:
		return "plug"
	}
	return "cloud"
}

func (r *aiProviderRow) view(s giode.PressState) giode.Element {
	c := theme.Current()
	bg := ""
	if s.Hovered {
		bg = c.Hover
	}
	kids := []giode.Element{
		theme.Icon(kindIcon(r.p.Kind), c.TextMuted),
		giode.Text(r.p.Name, giode.Styles{Color: c.Text, FontSize: theme.UISize, FlexGrow: 1, MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis}),
	}
	switch {
	case r.active:
		kids = append(kids, giode.Box(giode.Styles{BackgroundGradient: theme.AccentGradient(), BorderRadius: 8, Padding: giode.SymmetricInset(7, 1)},
			giode.Text("Active", giode.Styles{Color: c.AccentText, FontSize: 10, FontWeight: giode.FontWeightSemiBold})))
	case r.ready && r.p.NeedsKey:
		kids = append(kids, theme.IconSized("key", c.Success, 13))
	}
	return giode.Row(giode.Styles{
		Width: -1, Height: 30, Background: bg, BackgroundGradient: theme.RowWash(r.selected), BorderRadius: 7,
		Align: giode.AlignCenter, Gap: 8, Padding: giode.SymmetricInset(10, 0), Wrap: giode.WrapNoWrap,
	}, kids...)
}

// Layout draws the row.
func (r *aiProviderRow) Layout(gtx giode.Context) giode.Dimensions {
	return r.press.Layout(gtx)
}
