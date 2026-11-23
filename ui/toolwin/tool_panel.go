package toolwin

import (
	"github.com/xmhrnx/giode"

	"litecode/internal/layout"
	"litecode/ui/core"
	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// ToolPanel is an open panel: a header with its title, its actions, an
// options menu and a hide button, over its content.
type ToolPanel struct {
	ide     *core.IDE
	w       core.ToolWindow
	menu    *giode.MenuElement
	options *giode.PressableElement
	hide    *widgets.ToolButton
}

// NewToolPanel creates the panel of a tool window.
func NewToolPanel(ide *core.IDE, w core.ToolWindow) *ToolPanel {
	p := &ToolPanel{ide: ide, w: w, menu: newWindowMenu(ide, w.ID())}
	p.options = giode.Pressable(func(s giode.PressState) giode.Element {
		c := theme.Current()
		bg := ""
		if s.Hovered {
			bg = c.Hover
		}
		return giode.Box(giode.Styles{Width: 26, Height: 26, BorderRadius: 6, Background: bg, Align: giode.AlignCenter, Justify: giode.JustifyCenter},
			theme.Icon("more", c.TextFaint))
	}).Label("Options").OnClick(p.menu.OpenAtClick)
	p.hide = widgets.NewToolButton("hide", "Hide", func() {
		ide.Layout.Close(w.ID())
		ide.SaveLayout()
	})
	return p
}

// ID is the window's ID.
func (p *ToolPanel) ID() layout.WindowID { return p.w.ID() }

func (p *ToolPanel) header() giode.Element {
	c := theme.Current()
	kids := []giode.Element{
		theme.Icon(p.w.Icon(), c.TextAccent),
		giode.Box(giode.Styles{Width: 6}),
		giode.Text(p.w.Title(), giode.Styles{Color: c.Text, FontSize: theme.UISize, FontWeight: giode.FontWeightSemiBold}),
		giode.Box(giode.Styles{Width: 10}),
	}
	kids = append(kids, p.w.HeaderActions()...)
	kids = append(kids, giode.Box(giode.Styles{FlexGrow: 1}), p.options, p.hide, p.menu)
	return giode.Row(giode.Styles{
		Width: -1, Height: theme.ToolHeaderHeight, Align: giode.AlignCenter, Gap: 2, Wrap: giode.WrapNoWrap,
		Padding: giode.Inset{Left: 12, Right: 6},
	}, kids...)
}

// Layout draws the panel.
func (p *ToolPanel) Layout(gtx giode.Context) giode.Dimensions {
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Background: theme.Current().Panel},
		p.header(),
		theme.Separator(),
		giode.Box(giode.Styles{Width: -1, FlexGrow: 1, Overflow: giode.OverflowHidden}, p.w.View()),
	).Layout(gtx)
}
