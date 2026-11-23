package toolwin

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"

	"litecode/internal/layout"
	"litecode/ui/core"
	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// Dock arranges the stripes, the open tool windows and the center (the
// editor) in the window.
type Dock struct {
	ide       *core.IDE
	center    giode.Element
	left      *Stripe
	right     *Stripe
	panels    map[layout.WindowID]*ToolPanel
	splitters map[layout.Side]*widgets.Splitter
}

// NewDock creates the dock around center.
func NewDock(ide *core.IDE, center giode.Element) *Dock {
	d := &Dock{
		ide: ide, center: center,
		left:      NewStripe(ide, layout.Left, layout.Bottom),
		right:     NewStripe(ide, layout.Right, ""),
		panels:    map[layout.WindowID]*ToolPanel{},
		splitters: map[layout.Side]*widgets.Splitter{},
	}
	for id, w := range ide.Windows {
		d.panels[id] = NewToolPanel(ide, w)
	}
	for side, sign := range map[layout.Side]int{layout.Left: 1, layout.Right: -1, layout.Bottom: -1} {
		side := side
		d.splitters[side] = widgets.NewSplitter(side != layout.Bottom, sign,
			func() int { return ide.Layout.Sizes[side] },
			func(n int) { ide.Layout.Resize(side, n) },
			ide.SaveLayout)
	}
	return d
}

// panel returns the open panel on a side, or nil.
func (d *Dock) panel(side layout.Side) *ToolPanel {
	id, ok := d.ide.Layout.Open[side]
	if !ok {
		return nil
	}
	return d.panels[id]
}

func (d *Dock) sidePanel(side layout.Side) []giode.Element {
	p := d.panel(side)
	if p == nil {
		return nil
	}
	box := giode.Box(theme.Card(theme.Current().Panel).With(giode.Styles{Width: properties.Width(d.ide.Layout.Sizes[side]), Height: -1}), p)
	if side == layout.Left {
		return []giode.Element{box, d.splitters[side]}
	}
	return []giode.Element{d.splitters[side], box}
}

func (d *Dock) middle() giode.Element {
	kids := []giode.Element{giode.Box(theme.Card(theme.Current().Editor).With(giode.Styles{Width: -1, FlexGrow: 1}), d.center)}
	if p := d.panel(layout.Bottom); p != nil {
		kids = append(kids, d.splitters[layout.Bottom],
			giode.Box(theme.Card(theme.Current().Panel).With(giode.Styles{Width: -1, Height: properties.Height(d.ide.Layout.Sizes[layout.Bottom])}), p))
	}
	return giode.Stack(giode.Styles{FlexGrow: 1, Height: -1}, kids...)
}

// Layout draws the dock.
func (d *Dock) Layout(gtx giode.Context) giode.Dimensions {
	gap := giode.Box(giode.Styles{Width: theme.CardGap})
	kids := []giode.Element{d.left, gap}
	left := d.sidePanel(layout.Left)
	kids = append(kids, left...)
	kids = append(kids, d.middle())
	right := d.sidePanel(layout.Right)
	kids = append(kids, right...)
	if !d.right.Empty() {
		kids = append(kids, gap, d.right)
	}
	return giode.Row(giode.Styles{Width: -1, Height: -1, Wrap: giode.WrapNoWrap, Padding: giode.SymmetricInset(theme.CardGap, 0)}, kids...).Layout(gtx)
}
