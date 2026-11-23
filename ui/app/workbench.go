package app

import (
	"github.com/xmhrnx/giode"

	"litecode/internal/layout"
	"litecode/internal/workspace"
	"litecode/ui/assistant"
	"litecode/ui/core"
	"litecode/ui/editor"
	"litecode/ui/frame"
	"litecode/ui/gitui"
	"litecode/ui/projectview"
	"litecode/ui/search"
	"litecode/ui/settings"
	"litecode/ui/terminal"
	"litecode/ui/theme"
	"litecode/ui/toolwin"
)

// Workbench is the window of an open project.
type Workbench struct {
	ide       *core.IDE
	title     *frame.TitleBar
	editor    *editor.EditorArea
	status    *editor.StatusBar
	dock      *toolwin.Dock
	project   *projectview.ProjectView
	commit    *gitui.CommitWindow
	log       *gitui.LogWindow
	assistant *assistant.Window
	everyw    *search.SearchEverywhere
	find      *search.FindInFiles
	recent    *search.RecentFiles
	gotoLine  *search.GotoLine
	branches  *gitui.BranchesPopup
	push      *gitui.PushDialog
	newBranch *gitui.NewBranchDialog
	settings  *settings.Dialog
	keys      *Shortcuts
	focused   bool
}

// NewWorkbench builds the window for ide.Project. pluginDir is shown in
// Settings > Plugins.
func NewWorkbench(ide *core.IDE, pluginDir string, onClose func() bool) *Workbench {
	w := &Workbench{ide: ide, keys: NewShortcuts(ide), focused: true}
	w.editor = editor.NewEditorArea(ide)
	w.editor.OnTabsChange = w.saveSession
	w.status = editor.NewStatusBar(ide, w.editor)
	w.project = projectview.New(ide, func() string {
		if v := w.editor.Current(); v != nil {
			return v.Path()
		}
		return ""
	})
	w.commit = gitui.NewCommitWindow(ide)
	w.newBranch = gitui.NewNewBranchDialog(ide)
	w.log = gitui.NewLogWindow(ide, w.newBranch)
	ide.Windows = map[layout.WindowID]core.ToolWindow{}
	w.assistant = assistant.New(ide)
	for _, tw := range []core.ToolWindow{w.project, w.commit, w.log, terminal.New(ide), w.assistant} {
		ide.Windows[tw.ID()] = tw
	}
	w.everyw = search.NewSearchEverywhere(ide)
	w.find = search.NewFindInFiles(ide)
	w.recent = search.NewRecentFiles(ide, w.editor.Recent)
	w.gotoLine = search.NewGotoLine()
	w.branches = gitui.NewBranchesPopup(ide, w.newBranch)
	w.push = gitui.NewPushDialog(ide)
	w.settings = settings.NewDialog(ide, pluginDir, w.resetLayout)
	w.dock = toolwin.NewDock(ide, w.editor)
	w.title = frame.NewTitleBar(ide, onClose)
	ide.OpenFile = w.editor.Open
	ide.CurrentFile = func() (string, string) {
		if v := w.editor.Current(); v != nil {
			return v.Path(), v.Text()
		}
		return "", ""
	}
	ide.OpenSettings = w.settings.OpenPage
	ide.ShowDiff = func(r core.DiffRequest) {
		w.editor.OpenTab("diff:"+r.Path+":"+r.LeftLabel+":"+r.RightLabel, r.Title, gitui.NewDiffView(ide, r))
	}
	w.registerActions()
	if session, err := workspace.LoadSession(ide.Project.Root); err == nil {
		w.editor.Restore(session)
	} else {
		ide.Error(err)
	}
	if g := ide.Project.Git; g != nil {
		g.Refresh(ide, true)
	}
	return w
}

func (w *Workbench) saveSession() {
	w.ide.Error(workspace.SaveSession(w.ide.Project.Root, w.editor.Session()))
}

// Close saves everything before the project closes; it reports whether the
// window may close.
func (w *Workbench) Close(done func()) {
	w.editor.SaveAll(false, func() {
		w.saveSession()
		if done != nil {
			done()
		}
	})
}

// watchFocus saves files when the window loses focus and refreshes git when
// it comes back.
func (w *Workbench) watchFocus() {
	focused := w.ide.App.Focused()
	if focused == w.focused {
		return
	}
	w.focused = focused
	if !focused {
		w.editor.SaveAll(false, nil)
		return
	}
	if g := w.ide.Project.Git; g != nil {
		g.Refresh(w.ide, false)
	}
}

// Layout draws the window.
func (w *Workbench) Layout(gtx giode.Context) giode.Dimensions {
	w.keys.Register()
	w.watchFocus()
	c := theme.Current()
	return giode.Box(giode.Styles{Width: -1, Height: -1, FlexDirection: giode.FlexDirectionColumn, Background: c.Editor, BackgroundGradient: theme.Backdrop()},
		w.title,
		giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, w.dock),
		w.status,
		w.everyw.Element(),
		w.find.Element(),
		w.recent.Element(),
		w.branches.Element(),
		w.gotoLine, w.push, w.newBranch, w.settings,
	).Layout(gtx)
}
