package editor

import (
	"github.com/xmhrnx/giode"

	"litecode/ui/theme"
)

// installButton sits in the status bar while the selected file's language
// has no server, and installs it.
type installButton struct {
	editor *EditorArea
	press  *giode.PressableElement
}

func newInstallButton(editor *EditorArea) *installButton {
	b := &installButton{editor: editor}
	b.press = giode.Pressable(b.view).Focusable(false).Label("Install language server").
		OnPrimaryClick(func(giode.ClickEvent) {
			if v := editor.Current(); v != nil {
				v.InstallServer()
			}
		})
	return b
}

func (b *installButton) view(st giode.PressState) giode.Element {
	c := theme.Current()
	v := b.editor.Current()
	label := "Install language server"
	if in, ok := v.Installer(); ok {
		label = "Install " + in.Server
	}
	if v.Installing() {
		label = "Installing…"
	}
	bg := c.Accent
	if st.Hovered {
		bg = c.FocusLine
	}
	return giode.Box(giode.Styles{Background: bg, BorderRadius: 10, Padding: giode.SymmetricInset(9, 2), Cursor: giode.CursorPointer},
		giode.Text(label, giode.Styles{Color: c.AccentText, FontSize: 11, FontWeight: giode.FontWeightMedium}))
}

// Layout draws the button.
func (b *installButton) Layout(gtx giode.Context) giode.Dimensions { return b.press.Layout(gtx) }
