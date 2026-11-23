// Package terminal is the Terminal tool window: the user's shell, started
// in the project's folder the first time the window shows. It docks and
// moves like any other tool window.
package terminal

import (
	"os/exec"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/pty"

	"litecode/internal/layout"
	"litecode/ui/core"
	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// WindowID is the Terminal tool window's ID.
const WindowID layout.WindowID = "terminal"

// Window is the Terminal tool window.
type Window struct {
	ide     *core.IDE
	term    *giode.TerminalElement
	started bool
	exited  bool
	err     error
	restart *widgets.ToolButton
}

// New creates the Terminal tool window.
func New(ide *core.IDE) *Window {
	w := &Window{ide: ide}
	w.restart = widgets.NewToolButton("refresh", "Restart shell", w.restartShell)
	return w
}

func (w *Window) ID() layout.WindowID { return WindowID }
func (w *Window) Title() string       { return "Terminal" }
func (w *Window) Icon() string        { return "panel-terminal" }

// HeaderActions is the restart button.
func (w *Window) HeaderActions() []giode.Element { return []giode.Element{w.restart} }

// palette is the terminal's colors, matched to the editor's.
func palette() giode.TerminalPalette {
	c, cc := theme.Current(), theme.CodeScheme()
	return giode.TerminalPalette{
		Foreground: giode.Color(c.TextCode), Background: giode.Color(c.Editor),
		Cursor: giode.Color(cc.Caret), Selection: giode.Color(cc.Selection),
		ANSI: [16]giode.Color{
			"#2A2C36", "#F07A93", "#A5D88A", "#F5C451", "#82AAFF", "#C792EA", "#7FD6C2", "#D3D5DC",
			"#636776", "#FF8FA6", "#B9E79F", "#FFD675", "#9DBDFF", "#D7A8F2", "#97E3D2", "#FFFFFF",
		},
	}
}

// start runs the shell in the project's folder.
func (w *Window) start() {
	w.started, w.exited, w.err = true, false, nil
	w.term = giode.Terminal().
		FontFamily(theme.CodeFont).
		FontSize(float32(w.ide.Settings.EditorFontSize)).
		Padding(8).
		Palette(palette()).
		OnExit(func(int, error) { w.exited = true })
	var cmd *exec.Cmd
	if sh := pty.DefaultShell(); sh != "" {
		cmd = exec.Command(sh)
		if w.ide.Project != nil {
			cmd.Dir = w.ide.Project.Root
		}
	}
	if err := w.term.Start(cmd); err != nil {
		w.err = err
	}
}

// restartShell hangs up the running shell and starts a new one.
func (w *Window) restartShell() {
	if w.term != nil && !w.exited {
		w.term.Close()
	}
	w.start()
	w.term.Focus()
}

// View is the terminal, started on first view.
func (w *Window) View() giode.Element {
	if !w.started {
		w.start()
	}
	c := theme.Current()
	if w.err != nil || w.exited {
		msg := "The shell exited."
		if w.err != nil {
			msg = "The shell couldn't start: " + w.err.Error()
		}
		return giode.Stack(giode.Styles{Width: -1, Height: -1, Background: c.Editor, Align: giode.AlignCenter, Justify: giode.JustifyCenter, Gap: 6, Padding: giode.UniformInset(16)},
			giode.Text(msg, giode.Styles{Color: c.TextMuted, FontSize: theme.UISize}),
			giode.Text("Restart it with the button above.", giode.Styles{Color: c.TextFaint, FontSize: theme.SmallSize}),
		)
	}
	return giode.Box(giode.Styles{Width: -1, Height: -1, Background: c.Editor}, w.term)
}
