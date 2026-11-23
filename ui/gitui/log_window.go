package gitui

import (
	"context"
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"

	"litecode/internal/git"
	"litecode/internal/layout"
	"litecode/ui/core"
	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// LogWindowID is the Git tool window's ID.
const LogWindowID layout.WindowID = "history"

// LogWindow is the Git tool window's Log: branches, commits and the selected
// commit's details side by side.
type LogWindow struct {
	ide       *core.IDE
	branches  *BranchesPanel
	details   *CommitDetails
	list      *giode.VirtualListElement
	search    *input.Input
	newBranch *NewBranchDialog
	actions   []giode.Element
	rows      *giode.KeyedCache[string, *commitRow]
	menus     widgets.MenuSlot

	branch   string       // "" for all branches
	commits  []git.Commit // the branch's commits when filtered
	shown    []git.Commit // after the text filter
	shownFor [3]any       // what shown was computed from
	selected string
	locals   map[string]bool
	localsAt int // git state version locals was built from
}

// NewLogWindow creates the Git tool window; nb is the shared New Branch dialog.
func NewLogWindow(ide *core.IDE, nb *NewBranchDialog) *LogWindow {
	w := &LogWindow{ide: ide, rows: giode.NewKeyedCache[string, *commitRow](), newBranch: nb}
	w.branches = NewBranchesPanel(ide, w.filterBranch, w.newBranch.Ask)
	w.details = NewCommitDetails(ide)
	w.search = widgets.TextField("Text or hash", theme.InputStyles()).OnChange(widgets.Typing(func(string) { w.shownFor = [3]any{} }))
	w.list = giode.VirtualList(func() int { return len(w.visible()) }, theme.RowHeight, w.row,
		giode.Styles{ScrollBar: theme.ScrollBar(), Background: theme.Current().Editor})
	w.actions = []giode.Element{
		widgets.NewToolButton("refresh", "Refresh", func() { ide.Project.Git.Refresh(ide, true) }),
		widgets.NewToolButton("fetch", "Fetch", w.fetch),
	}
	return w
}

func (w *LogWindow) ID() layout.WindowID            { return LogWindowID }
func (w *LogWindow) Title() string                  { return "History" }
func (w *LogWindow) Icon() string                   { return "panel-history" }
func (w *LogWindow) HeaderActions() []giode.Element { return w.actions }

func (w *LogWindow) fetch() {
	w.ide.Project.Git.Do(w.ide, "Fetching…", func(ctx context.Context, r *git.Repo) error { return r.Fetch(ctx) },
		func() { w.ide.Info("Fetched all remotes") })
}

func (w *LogWindow) filterBranch(name string) {
	w.branch, w.commits = name, nil
	w.shownFor = [3]any{}
	if name == "" {
		return
	}
	repo := w.ide.Project.Git.Repo()
	var commits []git.Commit
	w.ide.Background(func(ctx context.Context) error {
		var err error
		commits, err = repo.Log(ctx, name, core.LogLimit)
		return err
	}, func() {
		if w.branch == name {
			w.commits = commits
			w.shownFor = [3]any{}
		}
	})
}

// visible returns the commits to list: the branch's, filtered by the search text.
func (w *LogWindow) visible() []git.Commit {
	g := w.ide.Project.Git
	query := strings.ToLower(strings.TrimSpace(w.search.Text()))
	key := [3]any{g.Version, w.branch + "\x00" + query, len(w.commits)}
	if key == w.shownFor {
		return w.shown
	}
	w.shownFor = key
	src := g.Log
	if w.branch != "" {
		src = w.commits
	}
	if query == "" {
		w.shown = src
		return w.shown
	}
	w.shown = nil
	for _, c := range src {
		if strings.Contains(strings.ToLower(c.Subject), query) || strings.HasPrefix(c.Hash, query) ||
			strings.Contains(strings.ToLower(c.Author), query) {
			w.shown = append(w.shown, c)
		}
	}
	return w.shown
}

func (w *LogWindow) row(i int) giode.Element {
	commits := w.visible()
	if i >= len(commits) {
		return giode.Box(giode.Styles{})
	}
	c := commits[i]
	r := w.rows.Get(c.Hash, func() *commitRow {
		return newCommitRow(w.selectCommit, w.commitMenu)
	})
	r.commit, r.selected, r.current, r.locals = c, c.Hash == w.selected, w.ide.Project.Git.Branch, w.localBranches()
	r.first, r.last = i == 0, i == len(commits)-1
	return r
}

// localBranches returns the local branch names, rebuilt when git changes.
func (w *LogWindow) localBranches() map[string]bool {
	g := w.ide.Project.Git
	if w.locals == nil || w.localsAt != g.Version {
		w.locals, w.localsAt = map[string]bool{}, g.Version
		for _, b := range g.Branches {
			if !b.Remote {
				w.locals[b.Name] = true
			}
		}
	}
	return w.locals
}

func (w *LogWindow) selectCommit(c git.Commit) {
	w.selected = c.Hash
	w.details.Show(c)
}

func (w *LogWindow) commitMenu(c git.Commit, e giode.ClickEvent) {
	g := w.ide.Project.Git
	w.menus.Open(theme.NewMenu().
		Item("Copy Hash", func() { w.ide.App.SetClipboard(c.Hash) }).
		Item("New Branch…", func() { w.newBranch.Ask(c.Hash) }).
		Item("Check Out Commit", func() {
			g.Do(w.ide, "Checking out "+c.Short+"…", func(ctx context.Context, r *git.Repo) error {
				return r.Checkout(ctx, git.Branch{Name: c.Hash})
			}, nil)
		}), e)
}

// commitList is the list, or a note when there's nothing in it.
func (w *LogWindow) commitList() giode.Element {
	if len(w.visible()) > 0 {
		return w.list
	}
	c := theme.Current()
	msg := "No commits yet"
	if strings.TrimSpace(w.search.Text()) != "" {
		msg = "No commits match"
	}
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Align: giode.AlignCenter, Justify: giode.JustifyCenter, Gap: 8},
		theme.IconSized("panel-history", c.TextFaint, 28),
		giode.Text(msg, giode.Styles{Color: c.TextMuted, FontSize: theme.UISize}))
}

func (w *LogWindow) toolbar() giode.Element {
	return giode.Row(giode.Styles{Width: -1, Height: 36, Align: giode.AlignCenter, Gap: 8, Padding: giode.SymmetricInset(8, 0), Wrap: giode.WrapNoWrap},
		giode.Box(giode.Styles{Width: 260}, w.search),
	)
}

// View draws the window.
func (w *LogWindow) View() giode.Element {
	c := theme.Current()
	if w.ide.Project.Git == nil {
		return giode.Box(giode.Styles{Padding: giode.UniformInset(16)},
			giode.Text("This project isn't a Git repository.", giode.Styles{Color: c.TextMuted, FontSize: theme.UISize}))
	}
	return giode.Row(giode.Styles{Width: -1, Height: -1, Wrap: giode.WrapNoWrap},
		giode.Box(giode.Styles{Width: 220, Height: -1}, w.branches),
		theme.VSeparator(),
		giode.Stack(giode.Styles{FlexGrow: 1, Height: -1}, w.toolbar(), theme.Separator(), giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, w.commitList())),
		theme.VSeparator(),
		giode.Box(giode.Styles{Width: 340, Height: -1}, w.details),
		&w.menus,
	)
}
