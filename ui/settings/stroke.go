package settings

import (
	"strings"

	"gioui.org/io/key"
	"github.com/xmhrnx/giode"
)

// keyNames maps Gio's key names to the shortcut syntax's.
var keyNames = map[key.Name]string{
	key.NameLeftArrow: "Left", key.NameRightArrow: "Right", key.NameUpArrow: "Up", key.NameDownArrow: "Down",
	key.NameEscape: "Esc", key.NameReturn: "Enter", key.NameEnter: "Enter", key.NameTab: "Tab", key.NameSpace: "Space",
	key.NameDeleteBackward: "Backspace", key.NameDeleteForward: "Delete", key.NamePageUp: "PageUp",
	key.NamePageDown: "PageDown", key.NameHome: "Home", key.NameEnd: "End",
}

// isModifier reports whether a key is a modifier pressed on its own.
func isModifier(n key.Name) bool {
	switch n {
	case key.NameShift, key.NameCtrl, key.NameAlt, key.NameCommand, key.NameSuper:
		return true
	}
	return false
}

// strokeOf writes a key press as a shortcut spec ("Mod+Shift+N"). The
// platform's shortcut modifier (Ctrl, or Command on macOS) is written "Mod",
// so a keymap carries over between systems.
func strokeOf(e key.Event) string {
	var parts []string
	m := e.Modifiers
	if m.Contain(key.ModShortcut) {
		parts = append(parts, "Mod")
		m &^= key.ModShortcut
	}
	if m.Contain(key.ModCtrl) {
		parts = append(parts, "Ctrl")
	}
	if m.Contain(key.ModAlt) {
		parts = append(parts, "Alt")
	}
	if m.Contain(key.ModShift) {
		parts = append(parts, "Shift")
	}
	if m.Contain(key.ModSuper) || m.Contain(key.ModCommand) {
		parts = append(parts, "Super")
	}
	name, ok := keyNames[e.Name]
	if !ok {
		name = strings.ToUpper(string(e.Name))
	}
	return strings.Join(append(parts, name), "+")
}

// formatStroke shows a stroke the platform's way.
func formatStroke(s string) string {
	if s == "DoubleShift" {
		return "Double ⇧"
	}
	return giode.FormatShortcut(s)
}
