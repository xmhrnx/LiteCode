// Package core is the state every part of the UI shares, and the hooks they
// use to ask for things another part shows (open a file, show a diff). The
// root view fills the hooks in; components only see this package, so they
// don't import each other.
package core

import (
	"context"
	"errors"
	"time"

	"github.com/xmhrnx/giode"

	"litecode/internal/config"
	"litecode/internal/keymap"
	"litecode/internal/lang"
	"litecode/internal/layout"
	"litecode/internal/recent"
	"litecode/ui/theme"
)

// ToolWindow is a panel that docks to a side: Project, Commit, Git, Terminal.
type ToolWindow interface {
	ID() layout.WindowID
	Title() string
	Icon() string // a registered icon name
	View() giode.Element
	// HeaderActions are the small buttons in the window's header.
	HeaderActions() []giode.Element
}

// DiffRequest asks the diff viewer to compare two versions of a file.
type DiffRequest struct {
	Title      string
	Path       string // relative to the project
	LeftLabel  string
	RightLabel string
	Left       func(ctx context.Context) (string, error)
	Right      func(ctx context.Context) (string, error)
}

// IDE is the shared state.
type IDE struct {
	App      *giode.App
	Settings config.Settings
	Keymap   *keymap.Keymap
	Langs    *lang.Registry
	Layout   *layout.Layout
	Recent   *recent.List
	Toasts   *giode.ToasterElement
	Project  *Project // nil on the welcome screen
	Windows  map[layout.WindowID]ToolWindow
	AI       *AI
	// CapturingKeys is set while Settings > Keymap records a stroke, so the
	// shortcuts don't fire.
	CapturingKeys bool

	// Hooks the root fills in.
	OpenFile     func(path string, line, col int)
	ShowDiff     func(DiffRequest)
	OpenProject  func(path string)
	CloseProject func()
	// CurrentFile returns the selected editor file's path and text ("" when none).
	CurrentFile func() (path, text string)
	// OpenSettings opens the Settings dialog on a page ("AI Assistant").
	OpenSettings func(page string)

	actions map[keymap.Action]func()
}

// New returns the shared state with the user's settings loaded. Problems
// loading them are returned together; the defaults stand in.
func New(app *giode.App, pluginDir string) (*IDE, []error) {
	var errs []error
	ide := &IDE{App: app, Windows: map[layout.WindowID]ToolWindow{}, actions: map[keymap.Action]func(){}}
	var err error
	if ide.Settings, err = config.LoadSettings(); err != nil {
		errs = append(errs, err)
	}
	ide.ApplyTheme()
	if ide.Keymap, err = keymap.Load(); err != nil {
		errs = append(errs, err)
	}
	if ide.Recent, err = recent.Load(); err != nil {
		errs = append(errs, err)
	}
	ide.Langs = lang.NewRegistry(pluginDir)
	errs = append(errs, ide.Langs.Errors...)
	return ide, errs
}

// ApplyTheme makes the theme in the settings the active colors.
func (ide *IDE) ApplyTheme() {
	s := ide.Settings
	theme.Apply(theme.Choice{UI: s.Theme, Syntax: s.SyntaxTheme, UIColors: s.UIColors, Code: s.CodeColors})
}

// Register binds the code that performs an action.
func (ide *IDE) Register(a keymap.Action, fn func()) {
	ide.actions[a] = fn
}

// Run performs an action, if anything registered it.
func (ide *IDE) Run(a keymap.Action) {
	if fn := ide.actions[a]; fn != nil {
		fn()
	}
}

// Has reports whether an action can run.
func (ide *IDE) Has(a keymap.Action) bool {
	return ide.actions[a] != nil
}

// Error shows an error in a balloon.
func (ide *IDE) Error(err error) {
	if err != nil {
		ide.Toasts.Show(giode.ToastError, err.Error())
	}
}

// Info shows a message in a balloon.
func (ide *IDE) Info(msg string) {
	ide.Toasts.Show(giode.ToastInfo, msg)
}

// Background runs work off the UI goroutine, then done (which may be nil) on
// it; an error goes to a balloon instead of done.
func (ide *IDE) Background(work func(ctx context.Context) error, done func()) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		err := work(ctx)
		ide.App.Post(func() {
			if err != nil {
				ide.Error(err)
				return
			}
			if done != nil {
				done()
			}
		})
	}()
}

// SaveLayout writes the tool window arrangement, reporting a failure.
func (ide *IDE) SaveLayout() {
	ide.Error(ide.Layout.Save())
}

// Shortcut returns an action's first key stroke, written the platform's way.
func (ide *IDE) Shortcut(a keymap.Action) string {
	s := ide.Keymap.First(a)
	if s == "" || s == keymap.DoubleShift {
		if s == keymap.DoubleShift {
			return "Double ⇧"
		}
		return ""
	}
	return giode.FormatShortcut(s)
}

// ChooseProject asks for a folder and opens it as the project.
func (ide *IDE) ChooseProject() {
	go func() {
		dir, err := ide.App.PickFolder(context.Background(), giode.FolderOptions{Title: "Open Project"})
		ide.App.Post(func() {
			switch {
			case errors.Is(err, giode.ErrCanceled):
			case err != nil:
				ide.Error(err)
			case dir != "":
				ide.OpenProject(dir)
			}
		})
	}()
}
