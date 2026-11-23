package gitui

import (
	"context"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/checkbox"

	"litecode/internal/git"
	"litecode/ui/core"
	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// NewBranchDialog asks for a branch name and whether to check it out.
type NewBranchDialog struct {
	ide      *core.IDE
	checkout *checkbox.Checkbox
	prompt   *widgets.PromptDialog
}

// NewNewBranchDialog creates the dialog.
func NewNewBranchDialog(ide *core.IDE) *NewBranchDialog {
	d := &NewBranchDialog{ide: ide}
	d.checkout = giode.Input.Checkbox("Checkout branch", giode.Styles{Background: theme.Current().Accent, Color: theme.Current().Text})
	d.checkout.SetChecked(true)
	d.prompt = widgets.NewPromptDialog(d.checkout)
	return d
}

// Ask opens the dialog; the branch starts at from (HEAD when empty).
func (d *NewBranchDialog) Ask(from string) {
	title := "Create New Branch"
	if from != "" {
		title += " from " + from
	}
	d.prompt.Ask(title, "", func(name string) {
		checkout := d.checkout.Checked()
		d.ide.Project.Git.Do(d.ide, "Creating "+name+"…", func(ctx context.Context, r *git.Repo) error {
			return r.NewBranch(ctx, name, from, checkout)
		}, nil)
	})
}

// Layout lays the dialog out; it takes no space.
func (d *NewBranchDialog) Layout(gtx giode.Context) giode.Dimensions {
	return d.prompt.Layout(gtx)
}
