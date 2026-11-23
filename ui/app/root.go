package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/xmhrnx/giode"

	"litecode/internal/config"
	"litecode/internal/layout"
	"litecode/ui/core"
	"litecode/ui/projectview"
	"litecode/ui/theme"
	"litecode/ui/welcome"
)

// Root is the window's content: the welcome screen, or the workbench of the
// open project.
type Root struct {
	ide       *core.IDE
	pluginDir string
	welcome   *welcome.Screen
	bench     *Workbench
	switching bool
}

// NewRoot loads the user's settings and creates the root. Problems loading
// them show as balloons once the window is up.
func NewRoot(a *giode.App, pluginDir string) *Root {
	ide, errs := core.New(a, pluginDir)
	r := &Root{ide: ide, pluginDir: pluginDir}
	ide.Toasts = giode.Toaster().Placement(giode.ToastBottomRight)
	var err error
	ide.Layout, err = layout.Load(defaultPlacements)
	errs = append(errs, err)
	if path, err := config.Path("layout.json"); err == nil {
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			ide.Layout.Show(projectview.WindowID)
		}
	}
	errs = append(errs, ide.LoadAI())
	ide.OpenProject = r.OpenProject
	ide.CloseProject = r.closeProject
	r.welcome = welcome.NewScreen(ide)
	for _, err := range errs {
		ide.Error(err)
	}
	return r
}

// Start opens the project named on the command line, or else the last one,
// or else stays on the welcome screen.
func (r *Root) Start(args []string) {
	if len(args) > 0 {
		r.OpenProject(args[0])
		return
	}
	if len(r.ide.Recent.Projects) > 0 {
		last := r.ide.Recent.Projects[0].Path
		if info, err := os.Stat(last); err == nil && info.IsDir() {
			r.OpenProject(last)
		}
	}
}

// OpenProject closes the open project (saving it) and opens the folder at path.
func (r *Root) OpenProject(path string) {
	abs, err := filepath.Abs(path)
	if err != nil {
		r.ide.Error(err)
		return
	}
	if info, err := os.Stat(abs); err != nil || !info.IsDir() {
		r.ide.Error(errors.New(path + " is not a folder"))
		return
	}
	if r.ide.Project != nil && r.ide.Project.Root == abs || r.switching {
		return
	}
	r.switching = true
	open := func() {
		go func() {
			p := core.OpenProject(context.Background(), abs)
			r.ide.App.Post(func() { r.show(p) })
		}()
	}
	if r.bench != nil {
		r.bench.Close(open)
		return
	}
	open()
}

func (r *Root) show(p *core.Project) {
	r.switching = false
	r.ide.Project = p
	r.ide.Error(r.ide.Recent.Touch(p.Root, time.Now()))
	r.bench = NewWorkbench(r.ide, r.pluginDir, r.onWindowClose)
	r.ide.App.SetTitle(p.Name + " – LiteCode")
}

func (r *Root) closeProject() {
	if r.bench == nil {
		return
	}
	r.bench.Close(func() {
		r.bench, r.ide.Project = nil, nil
		r.ide.App.SetTitle("LiteCode")
	})
}

// onWindowClose saves the project before the window closes.
func (r *Root) onWindowClose() bool {
	if r.bench == nil {
		return true
	}
	r.bench.Close(func() {
		r.bench = nil
		r.ide.App.Close()
	})
	return false
}

// FocusChanged passes the window gaining or losing focus to the workbench.
func (r *Root) FocusChanged(focused bool) {
	if r.bench != nil {
		r.bench.FocusChanged(focused)
	}
}

// HandleDrop opens dropped files in the editor, or a dropped folder as the project.
func (r *Root) HandleDrop(paths []string) {
	for _, p := range paths {
		info, err := os.Stat(p)
		switch {
		case err != nil:
			r.ide.Error(err)
		case info.IsDir():
			r.OpenProject(p)
			return
		case r.bench != nil:
			r.ide.OpenFile(p, 0, 0)
		}
	}
}

// View is the window's view function.
func (r *Root) View() giode.Element {
	var content giode.Element = r.welcome
	if r.bench != nil {
		content = r.bench
	}
	return giode.Box(giode.Styles{Width: -1, Height: -1, Background: theme.Current().Editor}, content, r.ide.Toasts)
}
