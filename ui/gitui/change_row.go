package gitui

import (
	"path"

	"github.com/xmhrnx/giode"

	"litecode/internal/git"
	"litecode/ui/theme"
)

// changeRow is one file of the Commit window's change list: a checkbox, the
// file's icon and name in its status color, and its folder.
type changeRow struct {
	change   git.Change
	selected bool
	check    *giode.CheckboxElement
	press    *giode.PressableElement
}

// newChangeRow creates a row; the callbacks get the row's current change.
func newChangeRow(onCheck func(c git.Change, on bool), onSelect, onOpen func(c git.Change), onMenu func(c git.Change, e giode.ClickEvent)) *changeRow {
	r := &changeRow{}
	r.check = giode.Input.Checkbox("", giode.Styles{Background: theme.Current().Accent}).
		OnChange(func(on bool) { onCheck(r.change, on) })
	r.press = giode.Pressable(r.view).
		OnClick(func(giode.ClickEvent) { onSelect(r.change) }).
		OnDoubleClick(func(giode.ClickEvent) { onOpen(r.change) }).
		OnSecondaryClick(func(e giode.ClickEvent) {
			onSelect(r.change)
			onMenu(r.change, e)
		})
	return r
}

func (r *changeRow) view(s giode.PressState) giode.Element {
	c := theme.Current()
	bg := ""
	switch {
	case r.selected:
		bg = c.Selection
	case s.Hovered:
		bg = c.Hover
	}
	dir := path.Dir(r.change.Path)
	if dir == "." {
		dir = ""
	}
	return giode.Row(giode.Styles{
		Width: -1, Height: theme.RowHeight, Background: bg, BackgroundGradient: theme.RowWash(r.selected), BorderRadius: 7, Margin: giode.SymmetricInset(6, 0), Align: giode.AlignCenter, Gap: 6, Wrap: giode.WrapNoWrap,
		Padding: giode.Inset{Left: 30, Right: 8},
	},
		r.check,
		theme.FileIcon(path.Base(r.change.Path), false),
		giode.Text(path.Base(r.change.Path), giode.Styles{Color: theme.StatusColor(r.change.Kind), FontSize: theme.UISize}),
		giode.Text(dir, theme.Label(c.TextFaint)),
	)
}

// Layout draws the row.
func (r *changeRow) Layout(gtx giode.Context) giode.Dimensions {
	return r.press.Layout(gtx)
}
