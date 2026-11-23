package frame

import (
	"github.com/xmhrnx/giode"

	"litecode/internal/keymap"
	"litecode/ui/core"
	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// TitleBar is the window's frame: main menu, project and branch widgets on the
// left; search and settings on the right; then the window buttons.
type TitleBar struct {
	bar *giode.TitleBarElement
}

// NewTitleBar creates the bar for an open project.
func NewTitleBar(ide *core.IDE, onClose func() bool) *TitleBar {
	c := theme.Current()
	bar := giode.TitleBar(ide.App, "", giode.Styles{Background: "transparent", Color: c.TextMuted, Height: theme.TitleBarHeight, FontSize: theme.UISize}).
		Leading(NewMainMenu(ide), giode.Box(giode.Styles{Width: 4}), NewProjectWidget(ide), giode.Box(giode.Styles{Width: 6}), NewBranchWidget(ide)).
		Trailing(
			NewSearchPill(ide),
			giode.Box(giode.Styles{Width: 8}),
			widgets.NewToolButton("settings", "Settings ("+ide.Shortcut(keymap.ShowSettings)+")", func() { ide.Run(keymap.ShowSettings) }),
			giode.Box(giode.Styles{Width: 8}),
		).
		OnClose(onClose)
	if giode.IsPlatform(giode.PlatformMacOS) {
		bar.Variant(giode.TitleBarMac)
	}
	return &TitleBar{bar: bar}
}

// Layout draws the bar.
func (t *TitleBar) Layout(gtx giode.Context) giode.Dimensions {
	return t.bar.Layout(gtx)
}
