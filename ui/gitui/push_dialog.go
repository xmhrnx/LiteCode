package gitui

import (
	"context"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/dialog"
	"github.com/xmhrnx/giode/components/scroll"

	"litecode/internal/git"
	"litecode/ui/core"
	"litecode/ui/theme"
)

// PushDialog shows the commits a push will send and pushes them (Ctrl+Shift+K).
type PushDialog struct {
	ide     *core.IDE
	open    bool
	commits []git.Commit
	target  string
	loading bool
	list    *scroll.Scroll
	push    *giode.ButtonElement
	cancel  *giode.ButtonElement
	dlg     *dialog.Dialog
}

// NewPushDialog creates the dialog.
func NewPushDialog(ide *core.IDE) *PushDialog {
	d := &PushDialog{ide: ide}
	d.list = giode.Scroll(d.listView, giode.Styles{ScrollBar: theme.ScrollBar(), MaxHeight: 320}).FitContent(true)
	d.push = giode.Button("Push", theme.PrimaryButton()).OnClick(d.doPush)
	d.cancel = giode.Button("Cancel", theme.SecondaryButton()).Variant(giode.VariantOutlined).OnClick(func() { d.open = false })
	d.dlg = giode.Dialog(&d.open, d.view, theme.DialogPanel(620))
	return d
}

// Open loads the outgoing commits and shows the dialog.
func (d *PushDialog) Open() {
	g := d.ide.Project.Git
	if g == nil {
		return
	}
	d.open, d.loading, d.commits = true, true, nil
	d.target = g.Branch
	for _, b := range g.Branches {
		if b.Current && b.Upstream != "" {
			d.target = b.Upstream
		}
	}
	repo := g.Repo()
	go func() {
		commits, err := repo.Outgoing(context.Background())
		d.ide.App.Post(func() {
			d.loading = false
			d.commits = commits
			d.ide.Error(err)
		})
	}()
}

func (d *PushDialog) doPush() {
	d.open = false
	g := d.ide.Project.Git
	g.Do(d.ide, "Pushing…", func(ctx context.Context, r *git.Repo) error { return r.Push(ctx) },
		func() { d.ide.Info("Pushed " + g.Branch + " to " + d.target) })
}

func (d *PushDialog) listView() giode.Element {
	c := theme.Current()
	if d.loading {
		return giode.Box(giode.Styles{Padding: giode.UniformInset(12)}, giode.Spinner(giode.Styles{Width: 16, Height: 16}))
	}
	if len(d.commits) == 0 {
		return giode.Box(giode.Styles{Padding: giode.UniformInset(12)},
			giode.Text("Nothing to push", giode.Styles{Color: c.TextFaint, FontSize: theme.UISize}))
	}
	var rows []giode.Element
	for _, cm := range d.commits {
		rows = append(rows, giode.Row(giode.Styles{Width: -1, Height: theme.RowHeight, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Padding: giode.SymmetricInset(10, 0)},
			giode.Text(cm.Short, giode.Styles{Color: c.TextFaint, FontFamily: theme.CodeFont, FontSize: 12}),
			giode.Text(cm.Subject, theme.Label(c.Text)),
			giode.Box(giode.Styles{FlexGrow: 1}),
			giode.Text(cm.Author, theme.Label(c.TextMuted))))
	}
	return giode.Stack(giode.Styles{Width: -1}, rows...)
}

func (d *PushDialog) view() giode.Element {
	c := theme.Current()
	g := d.ide.Project.Git
	return giode.Stack(giode.Styles{Width: -1, Gap: 12},
		giode.Text("Push Commits", giode.Styles{Color: c.Text, FontSize: 14, FontWeight: giode.FontWeightSemiBold}),
		giode.Row(giode.Styles{Gap: 6, Align: giode.AlignCenter},
			theme.Icon("branch", c.TextMuted),
			giode.Text(g.Branch, giode.Styles{Color: c.Text, FontSize: theme.UISize}),
			giode.Text("→", giode.Styles{Color: c.TextFaint, FontSize: theme.UISize}),
			giode.Text(d.target, giode.Styles{Color: c.RefRemote, FontSize: theme.UISize})),
		giode.Box(giode.Styles{Width: -1, Background: c.Editor, BorderRadius: 6, BorderWidth: 1, BorderColor: c.PopupLine}, d.list),
		giode.Row(giode.Styles{Width: -1, Justify: giode.JustifyEnd, Gap: 8}, d.cancel, d.push),
	)
}

// Layout lays the dialog out; it takes no space.
func (d *PushDialog) Layout(gtx giode.Context) giode.Dimensions {
	return d.dlg.Layout(gtx)
}
