// Package toolwin docks panels to the left, right and bottom of the window,
// with the activity bars that open them, the headers with their options, and
// the splitters that resize them.
package toolwin

import (
	"github.com/xmhrnx/giode"

	"litecode/internal/keymap"
	"litecode/internal/layout"
	"litecode/ui/core"
	"litecode/ui/theme"
)

// newWindowMenu builds a panel's options menu: Move To and Hide.
func newWindowMenu(ide *core.IDE, id layout.WindowID) *giode.MenuElement {
	moveTo := theme.NewMenu()
	for _, side := range []struct {
		side  layout.Side
		label string
	}{{layout.Left, "Left"}, {layout.Bottom, "Bottom"}, {layout.Right, "Right"}} {
		side := side
		moveTo.RadioItem(side.label,
			func() bool { return ide.Layout.Windows[id].Side == side.side },
			func() {
				ide.Layout.Move(id, side.side)
				ide.SaveLayout()
			})
	}
	return theme.NewMenu().
		SubMenu("Move To", moveTo).
		Separator().
		Item("Hide", func() {
			ide.Layout.Close(id)
			ide.SaveLayout()
		}).Shortcut(ide.Keymap.First(keymap.HideToolWindow))
}
