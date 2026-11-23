package gitui

import (
	"fmt"

	"github.com/xmhrnx/giode"

	"litecode/ui/theme"
)

// changeGroup is a node of the change list ("Changes", "Untracked")
// whose checkbox checks or unchecks every file in it.
type changeGroup struct {
	title string
	count int
	all   bool
	check *giode.CheckboxElement
}

func newChangeGroup(title string, onCheck func(on bool)) *changeGroup {
	g := &changeGroup{title: title}
	g.check = giode.Input.Checkbox("", giode.Styles{Background: theme.Current().Accent}).OnChange(onCheck)
	return g
}

// set updates the file count and whether all of them are checked.
func (g *changeGroup) set(count int, all bool) {
	g.count, g.all = count, all
	g.check.SetChecked(all)
}

// Layout draws the node.
func (g *changeGroup) Layout(gtx giode.Context) giode.Dimensions {
	c := theme.Current()
	files := "files"
	if g.count == 1 {
		files = "file"
	}
	return giode.Row(giode.Styles{Width: -1, Height: theme.RowHeight, Align: giode.AlignCenter, Gap: 6, Wrap: giode.WrapNoWrap, Padding: giode.Inset{Left: 10, Right: 8}},
		theme.IconSized("chevron-down", c.TextMuted, 14),
		g.check,
		giode.Text(g.title, giode.Styles{Color: c.Text, FontSize: theme.UISize}),
		giode.Text(fmt.Sprintf("%d %s", g.count, files), theme.Label(c.TextFaint)),
	).Layout(gtx)
}
