package widgets

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"

	"litecode/ui/theme"
)

// ListRow is a row of a popup list: an icon, a label, a dimmer detail and a
// hint on the right (a shortcut, a path). Its fields are set every frame;
// create the row once per item (with a KeyedCache).
type ListRow struct {
	Icon      giode.Element
	Label     string
	LabelTint string // label color; "" for the default
	Detail    string
	Hint      string
	Selected  bool
	Height    int
	press     *giode.PressableElement
}

// NewListRow creates a row; onClick gets the click (for menus at the pointer).
func NewListRow(onClick func(e giode.ClickEvent)) *ListRow {
	r := &ListRow{}
	r.press = giode.Pressable(r.view).Focusable(false).OnClick(onClick)
	return r
}

func (r *ListRow) view(s giode.PressState) giode.Element {
	c := theme.Current()
	bg := ""
	switch {
	case r.Selected:
		bg = c.Selection
	case s.Hovered:
		bg = c.Hover
	}
	color := r.LabelTint
	if color == "" {
		color = c.Text
	}
	h := r.Height
	if h == 0 {
		h = 26
	}
	kids := []giode.Element{}
	if r.Icon != nil {
		kids = append(kids, r.Icon)
	}
	kids = append(kids,
		giode.Text(r.Label, giode.Styles{Color: color, FontSize: theme.UISize, MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis}),
		giode.Text(r.Detail, theme.Label(c.TextFaint)),
		giode.Box(giode.Styles{FlexGrow: 1}),
		giode.Text(r.Hint, giode.Styles{Color: c.TextMuted, FontSize: theme.SmallSize, MaxLines: 1}),
	)
	return giode.Box(giode.Styles{Width: -1, Padding: giode.SymmetricInset(6, 0)},
		giode.Row(giode.Styles{
			Width: -1, Height: properties.Height(h), Background: bg, BackgroundGradient: theme.RowWash(r.Selected), BorderRadius: 7, Align: giode.AlignCenter, Gap: 8,
			Wrap: giode.WrapNoWrap, Padding: giode.SymmetricInset(8, 0), Overflow: giode.OverflowHidden,
		}, kids...))
}

// Layout draws the row.
func (r *ListRow) Layout(gtx giode.Context) giode.Dimensions {
	return r.press.Layout(gtx)
}
