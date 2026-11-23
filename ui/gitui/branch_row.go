package gitui

import (
	"github.com/xmhrnx/giode"

	"litecode/internal/git"
	"litecode/ui/theme"
)

// branchRow is one line of the branches panel. Its fields are refreshed every
// frame; the pressable is created once per key.
type branchRow struct {
	label    string
	depth    int
	branch   *git.Branch // nil for the HEAD row
	selected bool
	press    *giode.PressableElement
}

func newBranchRow(onClick, onDouble func(r *branchRow), onMenu func(r *branchRow, e giode.ClickEvent)) *branchRow {
	r := &branchRow{}
	r.press = giode.Pressable(r.view).
		OnPrimaryClick(func(giode.ClickEvent) { onClick(r) }).
		OnDoubleClick(func(giode.ClickEvent) { onDouble(r) }).
		OnSecondaryClick(func(e giode.ClickEvent) { onMenu(r, e) })
	return r
}

func (r *branchRow) view(s giode.PressState) giode.Element {
	c := theme.Current()
	bg := ""
	switch {
	case r.selected:
		bg = c.Selection
	case s.Hovered:
		bg = c.Hover
	}
	color := c.Text
	if r.branch != nil && r.branch.Current {
		color = c.RefHead
	}
	return giode.Row(giode.Styles{
		Width: -1, Height: theme.RowHeight, Background: bg, BackgroundGradient: theme.RowWash(r.selected), BorderRadius: 7, Margin: giode.SymmetricInset(6, 0), Align: giode.AlignCenter, Gap: 6, Wrap: giode.WrapNoWrap,
		Padding: giode.Inset{Left: 10 + 16*r.depth, Right: 8},
	}, theme.Icon("branch", c.TextMuted), giode.Text(r.label, theme.Label(color)))
}

// Layout draws the row.
func (r *branchRow) Layout(gtx giode.Context) giode.Dimensions {
	return r.press.Layout(gtx)
}
