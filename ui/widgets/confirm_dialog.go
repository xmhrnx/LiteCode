package widgets

import (
	"github.com/xmhrnx/giode"

	"litecode/ui/theme"
)

// ConfirmDialog asks a yes/no question. Create it once; Ask opens it.
type ConfirmDialog struct {
	open    bool
	title   string
	message string
	yes     *giode.ButtonElement
	no      *giode.ButtonElement
	dlg     *giode.DialogElement
	onYes   func()
}

// NewConfirmDialog creates a closed dialog.
func NewConfirmDialog() *ConfirmDialog {
	d := &ConfirmDialog{}
	d.yes = giode.Button("OK", theme.PrimaryButton()).OnClick(func() {
		d.open = false
		if d.onYes != nil {
			d.onYes()
		}
	})
	d.no = giode.Button("Cancel", theme.SecondaryButton()).Variant(giode.VariantOutlined).OnClick(func() { d.open = false })
	d.dlg = giode.Dialog(&d.open, d.view, theme.DialogPanel(440))
	return d
}

// Ask opens the dialog; yesLabel names the confirming button ("Delete").
func (d *ConfirmDialog) Ask(title, message, yesLabel string, onYes func()) {
	d.title, d.message, d.onYes, d.open = title, message, onYes, true
	d.yes.Label(yesLabel)
}

func (d *ConfirmDialog) view() giode.Element {
	c := theme.Current()
	return giode.Stack(giode.Styles{Gap: 14, Width: -1},
		giode.Text(d.title, giode.Styles{Color: c.Text, FontSize: 14, FontWeight: giode.FontWeightSemiBold}),
		giode.Text(d.message, giode.Styles{Color: c.TextMuted, FontSize: theme.UISize}),
		giode.Row(giode.Styles{Width: -1, Justify: giode.JustifyEnd, Gap: 8}, d.no, d.yes),
	)
}

// Layout lays the dialog out; it takes no space.
func (d *ConfirmDialog) Layout(gtx giode.Context) giode.Dimensions {
	return d.dlg.Layout(gtx)
}
