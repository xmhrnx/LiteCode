// Package terminal is the Terminal tool window. It docks and moves like any
// other tool window; the emulator inside needs a Giode terminal component
// (see docs/GIODE_GAPS.md), so until then it says so.
package terminal

import (
	"github.com/xmhrnx/giode"

	"litecode/internal/layout"
	"litecode/ui/theme"
)

// WindowID is the Terminal tool window's ID.
const WindowID layout.WindowID = "terminal"

// Window is the Terminal tool window.
type Window struct{}

// New creates the Terminal tool window.
func New() *Window { return &Window{} }

func (w *Window) ID() layout.WindowID            { return WindowID }
func (w *Window) Title() string                  { return "Terminal" }
func (w *Window) Icon() string                   { return "panel-terminal" }
func (w *Window) HeaderActions() []giode.Element { return nil }

// View draws the placeholder.
func (w *Window) View() giode.Element {
	c := theme.Current()
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Background: c.Editor, Align: giode.AlignCenter, Justify: giode.JustifyCenter, Gap: 6, Padding: giode.UniformInset(16)},
		giode.Text("Terminal emulator not available yet", giode.Styles{Color: c.TextMuted, FontSize: theme.UISize}),
		giode.Text("Needs a Giode Terminal component: PTY + VT parser + cell grid (docs/GIODE_GAPS.md)", giode.Styles{Color: c.TextFaint, FontSize: theme.SmallSize, TextAlign: giode.TextAlignCenter}),
	)
}
