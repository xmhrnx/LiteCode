package welcome

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/xmhrnx/giode"

	"litecode/internal/git"
	"litecode/ui/core"
	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// CloneDialog is Clone Repository: a repository URL and the folder to clone into.
type CloneDialog struct {
	ide    *core.IDE
	open   bool
	url    *giode.InputElement
	dir    *giode.InputElement
	clone  *giode.ButtonElement
	cancel *giode.ButtonElement
	dlg    *giode.DialogElement
}

// NewCloneDialog creates the dialog.
func NewCloneDialog(ide *core.IDE) *CloneDialog {
	d := &CloneDialog{ide: ide}
	d.url = widgets.TextField("https://github.com/owner/repo.git", theme.InputStyles()).OnChange(widgets.Typing(d.suggestDir))
	d.dir = widgets.TextField("Directory", theme.InputStyles())
	d.clone = giode.Button("Clone", theme.PrimaryButton()).OnClick(d.doClone)
	d.cancel = giode.Button("Cancel", theme.SecondaryButton()).Variant(giode.VariantOutlined).OnClick(func() { d.open = false })
	d.dlg = giode.Dialog(&d.open, d.view, theme.DialogPanel(560))
	return d
}

// Open shows the dialog.
func (d *CloneDialog) Open() {
	d.open = true
	d.url.SetText("")
	d.dir.SetText("")
	d.url.Focus()
}

// suggestDir proposes ~/Projects/<repo name> as the URL is typed.
func (d *CloneDialog) suggestDir(url string) {
	name := strings.TrimSuffix(filepath.Base(strings.TrimRight(url, "/")), ".git")
	if name == "" || name == "." {
		return
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	d.dir.SetText(filepath.Join(home, "Projects", name))
}

func (d *CloneDialog) doClone() {
	url, dir := strings.TrimSpace(d.url.Text()), strings.TrimSpace(d.dir.Text())
	if url == "" || dir == "" {
		d.ide.Error(errors.New("enter the repository URL and a directory"))
		return
	}
	d.clone.Loading(true)
	go func() {
		err := os.MkdirAll(filepath.Dir(dir), 0o755)
		if err == nil {
			err = git.Clone(context.Background(), url, dir)
		}
		d.ide.App.Post(func() {
			d.clone.Loading(false)
			if err != nil {
				d.ide.Error(err)
				return
			}
			d.open = false
			d.ide.OpenProject(dir)
		})
	}()
}

func (d *CloneDialog) view() giode.Element {
	c := theme.Current()
	label := func(s string) giode.Element {
		return giode.Text(s, giode.Styles{Color: c.TextMuted, FontSize: theme.UISize})
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 10},
		giode.Text("Clone Repository", giode.Styles{Color: c.Text, FontSize: 14, FontWeight: giode.FontWeightSemiBold}),
		giode.Grid(giode.Styles{GridColumns: "auto 1fr", Gap: 10, RowGap: 8, Align: giode.AlignCenter, Width: -1},
			label("URL:"), d.url, label("Directory:"), d.dir),
		giode.Row(giode.Styles{Width: -1, Justify: giode.JustifyEnd, Gap: 8}, d.cancel, d.clone),
	)
}

// Layout lays the dialog out; it takes no space.
func (d *CloneDialog) Layout(gtx giode.Context) giode.Dimensions {
	return d.dlg.Layout(gtx)
}
