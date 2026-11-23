package frame

import (
	"github.com/xmhrnx/giode"

	"litecode/internal/keymap"
	"litecode/ui/core"
	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// MainMenu is the ☰ button holding the main menu.
type MainMenu struct {
	ide    *core.IDE
	menu   *giode.MenuElement
	button *widgets.ToolButton
}

// NewMainMenu creates the menu.
func NewMainMenu(ide *core.IDE) *MainMenu {
	m := &MainMenu{ide: ide, menu: theme.NewMenu()}
	m.menu.
		SubMenu("File", m.sub(keymap.OpenProject, "", keymap.SaveAll, keymap.ShowSettings)).
		SubMenu("Edit", m.sub(keymap.DuplicateLine, keymap.DeleteLine, keymap.CommentLine, keymap.MoveLineUp, keymap.MoveLineDown, "", keymap.ReformatCode, "", keymap.FindInPath)).
		SubMenu("View", m.sub(keymap.ProjectWindow, keymap.CommitWindow, keymap.GitWindow, keymap.TerminalWindow, keymap.AssistantWindow, "", keymap.HideToolWindow, keymap.HideAllWindows)).
		SubMenu("Navigate", m.sub(keymap.SearchEverywhere, keymap.GotoFile, keymap.GotoAction, keymap.RecentFiles, keymap.GotoLine, "", keymap.NextTab, keymap.PreviousTab)).
		SubMenu("Git", m.sub(keymap.CommitAction, keymap.PushAction, keymap.UpdateProject, "", keymap.Branches, keymap.Rollback)).
		Separator().
		Item("Close Project", func() { ide.CloseProject() }).
		Item("Exit", func() { ide.App.RequestClose() })
	m.button = widgets.NewToolButtonEvent("menu", "Main Menu", m.menu.OpenAtClick)
	return m
}

// sub builds a submenu of actions; "" is a separator.
func (m *MainMenu) sub(actions ...keymap.Action) *giode.MenuElement {
	sub := theme.NewMenu()
	for _, a := range actions {
		if a == "" {
			sub.Separator()
			continue
		}
		a := a
		sub.Item(keymap.Names[a], func() { m.ide.Run(a) })
		if s := m.ide.Keymap.First(a); s != "" && s != keymap.DoubleShift {
			sub.Shortcut(s)
		}
	}
	return sub
}

// Layout draws the button and its menu.
func (m *MainMenu) Layout(gtx giode.Context) giode.Dimensions {
	return giode.Box(giode.Styles{}, m.button, m.menu).Layout(gtx)
}
