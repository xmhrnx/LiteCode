// Package gitui is the Git user interface: the Commit and Git (log) tool
// windows, the diff viewer, the branches popup and the push dialog.
package gitui

import (
	"strconv"

	"github.com/xmhrnx/giode"

	"litecode/internal/diff"
	"litecode/ui/theme"
)

// diffRowHeight is the height of one line in the diff viewer, in dp.
const diffRowHeight = 20

// diffRow draws one row of a side-by-side diff: each side's line number and text.
func diffRow(r diff.Row) giode.Element {
	c := theme.Current()
	leftBg, rightBg := "", ""
	switch r.Kind {
	case diff.Deleted:
		leftBg = c.DiffDeleted
	case diff.Inserted:
		rightBg = c.DiffInserted
	case diff.Changed:
		leftBg, rightBg = c.DiffChanged, c.DiffChanged
	}
	return giode.Row(giode.Styles{Width: -1, Height: diffRowHeight, Wrap: giode.WrapNoWrap},
		diffSide(r.LeftNo, r.Left, leftBg),
		giode.Box(giode.Styles{Width: 1, Height: -1, Background: c.PopupLine}),
		diffSide(r.RightNo, r.Right, rightBg),
	)
}

func diffSide(no int, text, bg string) giode.Element {
	c := theme.Current()
	num := ""
	if no > 0 {
		num = strconv.Itoa(no)
	}
	return giode.Row(giode.Styles{FlexGrow: 1, Height: -1, Background: bg, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Overflow: giode.OverflowHidden},
		giode.Text(num, giode.Styles{Color: c.TextFaint, FontFamily: theme.CodeFont, FontSize: 12, Width: 44, TextAlign: giode.TextAlignEnd}),
		giode.Box(giode.Styles{Width: 12}),
		giode.Text(text, theme.Code(c.TextCode)),
	)
}
