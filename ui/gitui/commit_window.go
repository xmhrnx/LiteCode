package gitui

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"github.com/xmhrnx/giode"

	"litecode/internal/git"
	"litecode/internal/keymap"
	"litecode/internal/layout"
	"litecode/ui/core"
	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// CommitWindowID is the Commit tool window's ID.
const CommitWindowID layout.WindowID = "changes"

// CommitWindow is the Commit tool window: the local changes with checkboxes,
// the commit message, Amend, and Commit / Commit and Push.
type CommitWindow struct {
	ide       *core.IDE
	checked   map[string]bool
	seen      map[string]bool // paths shown before, so new ones get a default check
	selected  string
	tracked   *changeGroup
	untracked *changeGroup
	list      *giode.ScrollElement
	message   *giode.InputElement
	amend     *giode.CheckboxElement
	commit    *giode.ButtonElement
	push      *giode.ButtonElement
	actions   []giode.Element
	confirm   *widgets.ConfirmDialog
	rows      *giode.KeyedCache[string, *changeRow]
}

// NewCommitWindow creates the Commit tool window.
func NewCommitWindow(ide *core.IDE) *CommitWindow {
	w := &CommitWindow{ide: ide, checked: map[string]bool{}, seen: map[string]bool{}, confirm: widgets.NewConfirmDialog()}
	w.rows = giode.NewKeyedCache[string, *changeRow]()
	w.tracked = newChangeGroup("Changes", func(on bool) { w.checkGroup(false, on) })
	w.untracked = newChangeGroup("Untracked", func(on bool) { w.checkGroup(true, on) })
	w.list = giode.Scroll(w.listView, giode.Styles{ScrollBar: theme.ScrollBar()})
	c := theme.Current()
	w.message = widgets.TextField("Commit message", giode.Styles{
		Width: -1, Height: -1, Background: c.Editor, Color: c.Text, BorderColor: c.PopupLine, BorderRadius: 8, FontSize: theme.UISize,
		Padding: giode.UniformInset(10),
	}).Multiline(true)
	w.amend = giode.Input.Checkbox("Amend", giode.Styles{Background: c.Accent, Color: c.Text}).OnChange(w.amendChanged)
	w.commit = giode.Button("Commit", theme.PrimaryButton()).OnClick(func() { w.doCommit(false) })
	w.push = giode.Button("Commit and Push…", theme.SecondaryButton()).Variant(giode.VariantOutlined).OnClick(func() { w.doCommit(true) })
	w.actions = []giode.Element{
		widgets.NewToolButton("refresh", "Refresh", func() { ide.Project.Git.Refresh(ide, false) }),
		widgets.NewToolButton("revert", "Discard Changes… ("+ide.Shortcut(keymap.Rollback)+")", w.Rollback),
		widgets.NewToolButton("diff", "Show Diff", func() {
			if c, ok := w.change(w.selected); ok {
				ide.ShowDiff(workingTreeDiff(ide, c))
			}
		}),
	}
	return w
}

func (w *CommitWindow) ID() layout.WindowID            { return CommitWindowID }
func (w *CommitWindow) Title() string                  { return "Changes" }
func (w *CommitWindow) Icon() string                   { return "panel-changes" }
func (w *CommitWindow) HeaderActions() []giode.Element { return w.actions }

// FocusMessage puts the keyboard in the message field (Ctrl+K).
func (w *CommitWindow) FocusMessage() { w.message.Focus() }

func (w *CommitWindow) changes() []git.Change {
	return w.ide.Project.Git.Changes
}

func (w *CommitWindow) change(path string) (git.Change, bool) {
	for _, c := range w.changes() {
		if c.Path == path {
			return c, true
		}
	}
	return git.Change{}, false
}

// syncChecks gives new files their default check: tracked changes are
// included, untracked files aren't.
func (w *CommitWindow) syncChecks() {
	for _, c := range w.changes() {
		if !w.seen[c.Path] {
			w.seen[c.Path] = true
			w.checked[c.Path] = c.Kind != git.Untracked
		}
	}
}

func (w *CommitWindow) checkGroup(untracked, on bool) {
	for _, c := range w.changes() {
		if (c.Kind == git.Untracked) == untracked {
			w.checked[c.Path] = on
		}
	}
}

// checkedChanges returns the files the commit will include.
func (w *CommitWindow) checkedChanges() []git.Change {
	var out []git.Change
	for _, c := range w.changes() {
		if w.checked[c.Path] {
			out = append(out, c)
		}
	}
	return out
}

func (w *CommitWindow) amendChanged(on bool) {
	if !on || strings.TrimSpace(w.message.Text()) != "" {
		return
	}
	repo := w.ide.Project.Git.Repo()
	var msg string
	w.ide.Background(func(ctx context.Context) error {
		var err error
		msg, err = repo.LastMessage(ctx)
		return err
	}, func() {
		if strings.TrimSpace(w.message.Text()) == "" {
			w.message.SetText(msg)
		}
	})
}

func (w *CommitWindow) doCommit(andPush bool) {
	msg := strings.TrimSpace(w.message.Text())
	files := w.checkedChanges()
	switch {
	case msg == "":
		w.ide.Error(errors.New("enter a commit message"))
		return
	case len(files) == 0 && !w.amend.Checked():
		w.ide.Error(errors.New("select the files to commit"))
		return
	}
	opts := git.CommitOptions{Message: msg, Files: files, Amend: w.amend.Checked()}
	w.ide.Project.Git.Do(w.ide, "Committing…", func(ctx context.Context, r *git.Repo) error {
		return r.Commit(ctx, opts)
	}, func() {
		w.message.SetText("")
		w.amend.SetChecked(false)
		w.ide.Info(committedMessage(len(files), msg))
		if andPush {
			w.ide.Run(keymap.PushAction)
		}
	})
}

func committedMessage(n int, msg string) string {
	subject := strings.SplitN(msg, "\n", 2)[0]
	if n == 1 {
		return "1 file committed: " + subject
	}
	return itoa(n) + " files committed: " + subject
}

// Rollback discards the checked files' changes after asking.
func (w *CommitWindow) Rollback() {
	files := w.checkedChanges()
	if len(files) == 0 {
		if c, ok := w.change(w.selected); ok {
			files = []git.Change{c}
		}
	}
	w.rollback(files)
}

func (w *CommitWindow) rollback(files []git.Change) {
	if len(files) == 0 {
		return
	}
	what := files[0].Path
	if len(files) > 1 {
		what = itoa(len(files)) + " files"
	}
	w.confirm.Ask("Discard Changes", "Discard the local changes to "+what+"? Untracked files are deleted.", "Discard", func() {
		w.ide.Project.Git.Do(w.ide, "Discarding…", func(ctx context.Context, r *git.Repo) error {
			return r.Rollback(ctx, files)
		}, nil)
	})
}

func (w *CommitWindow) rowMenu(c git.Change, e giode.ClickEvent) {
	path := c.Path
	theme.NewMenu().
		Item("Show Diff", func() {
			if ch, ok := w.change(path); ok {
				w.ide.ShowDiff(workingTreeDiff(w.ide, ch))
			}
		}).
		Item("Open File", func() { w.ide.OpenFile(w.ide.Project.Abs(path), 0, 0) }).
		Separator().
		Item("Discard Changes…", func() {
			if ch, ok := w.change(path); ok {
				w.rollback([]git.Change{ch})
			}
		}).
		Item("Copy Path", func() { w.ide.App.SetClipboard(filepath.FromSlash(path)) }).OpenAtClick(e)
}

func (w *CommitWindow) row(c git.Change) giode.Element {
	r := w.rows.Get(c.Path, func() *changeRow {
		return newChangeRow(
			func(c git.Change, on bool) { w.checked[c.Path] = on },
			func(c git.Change) { w.selected = c.Path },
			func(c git.Change) { w.ide.ShowDiff(workingTreeDiff(w.ide, c)) },
			w.rowMenu)
	})
	r.change, r.selected = c, c.Path == w.selected
	r.check.SetChecked(w.checked[c.Path])
	return r
}

func (w *CommitWindow) listView() giode.Element {
	w.syncChecks()
	var tracked, untracked []giode.Element
	allT, allU := true, true
	for _, c := range w.changes() {
		if c.Kind == git.Untracked {
			untracked = append(untracked, w.row(c))
			allU = allU && w.checked[c.Path]
		} else {
			tracked = append(tracked, w.row(c))
			allT = allT && w.checked[c.Path]
		}
	}
	var kids []giode.Element
	if len(tracked) > 0 {
		w.tracked.set(len(tracked), allT)
		kids = append(append(kids, w.tracked), tracked...)
	}
	if len(untracked) > 0 {
		w.untracked.set(len(untracked), allU)
		kids = append(append(kids, w.untracked), untracked...)
	}
	if len(kids) == 0 {
		c := theme.Current()
		return giode.Box(giode.Styles{Width: -1, Padding: giode.UniformInset(20), Justify: giode.JustifyCenter},
			giode.Text("No changes", giode.Styles{Color: c.TextFaint, FontSize: theme.UISize}))
	}
	return giode.Stack(giode.Styles{Width: -1, Padding: giode.Inset{Top: 2, Bottom: 2}}, kids...)
}

// View draws the window.
func (w *CommitWindow) View() giode.Element {
	c := theme.Current()
	if w.ide.Project.Git == nil {
		return giode.Box(giode.Styles{Padding: giode.UniformInset(16)},
			giode.Text("This project isn't a Git repository.", giode.Styles{Color: c.TextMuted, FontSize: theme.UISize}))
	}
	return giode.Stack(giode.Styles{Width: -1, Height: -1},
		giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, w.list),
		theme.Separator(),
		giode.Stack(giode.Styles{Width: -1, Gap: 10, Padding: giode.UniformInset(12)},
			giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
				giode.Text("Message", giode.Styles{Color: c.TextFaint, FontSize: theme.SmallSize, FontWeight: giode.FontWeightMedium}),
				giode.Box(giode.Styles{FlexGrow: 1}), w.amend),
			giode.Box(giode.Styles{Width: -1, Height: 110}, w.message),
			giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, w.commit, w.push)),
		w.confirm,
	)
}
