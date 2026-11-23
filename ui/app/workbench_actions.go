package app

import (
	"context"

	"litecode/internal/git"
	"litecode/internal/keymap"
	"litecode/internal/layout"
	"litecode/ui/assistant"
	"litecode/ui/editor"
	"litecode/ui/gitui"
	"litecode/ui/projectview"
	"litecode/ui/search"
	"litecode/ui/terminal"
)

// registerActions binds every keymap action to what it does here.
func (w *Workbench) registerActions() {
	ide := w.ide
	view := func(fn func(v *editor.CodeView)) func() {
		return func() {
			if v := w.editor.Current(); v != nil {
				fn(v)
			}
		}
	}
	diffView := func(fn func(d *gitui.DiffView)) func() {
		return func() {
			if d, ok := w.editor.CurrentExtra().(*gitui.DiffView); ok {
				fn(d)
			}
		}
	}
	showWindow := func(id layout.WindowID) func() {
		return func() {
			ide.Layout.Toggle(id)
			ide.SaveLayout()
			if g := ide.Project.Git; g != nil && ide.Layout.IsOpen(id) && (id == gitui.CommitWindowID || id == gitui.LogWindowID) {
				g.Refresh(ide, id == gitui.LogWindowID)
			}
		}
	}
	gitOp := func(label string, op func(ctx context.Context, r *git.Repo) error) func() {
		return func() {
			if g := ide.Project.Git; g != nil {
				g.Do(ide, label, op, nil)
			}
		}
	}
	for a, fn := range map[keymap.Action]func(){
		keymap.SearchEverywhere: func() { w.toggleSearch(search.TabAll) },
		keymap.GotoFile:         func() { w.toggleSearch(search.TabFiles) },
		keymap.GotoAction:       func() { w.toggleSearch(search.TabActions) },
		keymap.FindInPath:       func() { w.find.Open("") },
		keymap.RecentFiles:      w.recent.Open,
		keymap.GotoLine: view(func(v *editor.CodeView) {
			line, col := v.Caret()
			w.gotoLine.Ask(line, col, v.GoTo)
		}),
		keymap.Switcher:         func() { w.editor.SelectNext(1) },
		keymap.SwitcherBackward: func() { w.editor.SelectNext(-1) },
		keymap.NextTab:          func() { w.editor.SelectNext(1) },
		keymap.PreviousTab:      func() { w.editor.SelectNext(-1) },
		keymap.CloseEditor:      w.editor.CloseCurrent,
		keymap.SaveAll:          w.saveAll,
		keymap.ReformatCode:     view(func(v *editor.CodeView) { v.Format(nil) }),
		keymap.DuplicateLine:    view(func(v *editor.CodeView) { v.DuplicateLine() }),
		keymap.DeleteLine:       view(func(v *editor.CodeView) { v.DeleteLine() }),
		keymap.CommentLine:      view(func(v *editor.CodeView) { v.CommentLine() }),
		keymap.MoveLineUp:       view(func(v *editor.CodeView) { v.MoveLine(-1) }),
		keymap.MoveLineDown:     view(func(v *editor.CodeView) { v.MoveLine(1) }),
		keymap.ShowSettings:     w.settings.Open,
		keymap.OpenProject:      ide.ChooseProject,
		keymap.ProjectWindow:    showWindow(projectview.WindowID),
		keymap.CommitWindow:     showWindow(gitui.CommitWindowID),
		keymap.GitWindow:        showWindow(gitui.LogWindowID),
		keymap.TerminalWindow:   showWindow(terminal.WindowID),
		keymap.AssistantWindow: func() {
			showWindow(assistant.WindowID)()
			if ide.Layout.IsOpen(assistant.WindowID) {
				w.assistant.Focus()
			}
		},
		keymap.HideToolWindow:     w.hideLastWindow,
		keymap.HideAllWindows:     func() { ide.Layout.HideAll(); ide.SaveLayout() },
		keymap.CommitAction:       w.showCommit,
		keymap.PushAction:         w.push.Open,
		keymap.UpdateProject:      gitOp("Updating…", func(ctx context.Context, r *git.Repo) error { return r.Update(ctx) }),
		keymap.Branches:           w.branches.Open,
		keymap.Rollback:           w.commit.Rollback,
		keymap.NextDifference:     diffView(func(d *gitui.DiffView) { d.Next() }),
		keymap.PreviousDifference: diffView(func(d *gitui.DiffView) { d.Prev() }),
	} {
		ide.Register(a, fn)
	}
}

func (w *Workbench) toggleSearch(t search.Tab) {
	w.everyw.Open(t)
}

func (w *Workbench) saveAll() {
	w.editor.SaveAll(true, func() {
		if g := w.ide.Project.Git; g != nil {
			g.Refresh(w.ide, false)
		}
	})
}

func (w *Workbench) showCommit() {
	if w.ide.Project.Git == nil {
		w.ide.Info("This project isn't a Git repository")
		return
	}
	w.editor.SaveAll(false, nil)
	w.ide.Layout.Show(gitui.CommitWindowID)
	w.ide.SaveLayout()
	w.ide.Project.Git.Refresh(w.ide, false)
	w.commit.FocusMessage()
}

// hideLastWindow closes the bottom window, else the right, else the left
// (Shift+Esc).
func (w *Workbench) hideLastWindow() {
	for _, side := range []layout.Side{layout.Bottom, layout.Right, layout.Left} {
		if id, ok := w.ide.Layout.Open[side]; ok {
			w.ide.Layout.Close(id)
			w.ide.SaveLayout()
			return
		}
	}
}

func (w *Workbench) resetLayout() {
	l := layout.New(defaultPlacements)
	l.Show(projectview.WindowID)
	w.ide.Layout.Windows, w.ide.Layout.Open, w.ide.Layout.Sizes = l.Windows, l.Open, l.Sizes
	w.ide.SaveLayout()
}
