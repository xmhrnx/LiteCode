package widgets

import (
	"github.com/xmhrnx/giode"

	"litecode/ui/theme"
)

// PromptDialog asks for one line of text: a new file's name, a branch name.
// Create it once; Ask opens it.
type PromptDialog struct {
	open   bool
	title  string
	field  *giode.InputElement
	ok     *giode.ButtonElement
	cancel *giode.ButtonElement
	dlg    *giode.DialogElement
	submit func(string)
	extra  giode.Element
}

// NewPromptDialog creates a closed dialog; extra (which may be nil) is laid
// out under the field, for options such as "Checkout branch".
func NewPromptDialog(extra giode.Element) *PromptDialog {
	d := &PromptDialog{extra: extra}
	d.field = TextField("", theme.InputStyles()).Submit(func(string) { d.accept() })
	d.ok = giode.Button("OK", theme.PrimaryButton()).OnClick(d.accept)
	d.cancel = giode.Button("Cancel", theme.SecondaryButton()).Variant(giode.VariantOutlined).OnClick(func() { d.open = false })
	d.dlg = giode.Dialog(&d.open, d.view, theme.DialogPanel(420))
	return d
}

// Ask opens the dialog with a title and initial text; submit gets the text on OK.
func (d *PromptDialog) Ask(title, initial string, submit func(string)) {
	d.title, d.submit, d.open = title, submit, true
	d.field.SetText(initial)
	d.field.SelectAll()
	d.field.Focus()
}

func (d *PromptDialog) accept() {
	text := d.field.Text()
	if text == "" {
		return
	}
	d.open = false
	if d.submit != nil {
		d.submit(text)
	}
}

func (d *PromptDialog) view() giode.Element {
	c := theme.Current()
	kids := []giode.Element{
		giode.Text(d.title, giode.Styles{Color: c.Text, FontSize: 14, FontWeight: giode.FontWeightSemiBold}),
		d.field,
	}
	if d.extra != nil {
		kids = append(kids, d.extra)
	}
	kids = append(kids, giode.Row(giode.Styles{Width: -1, Justify: giode.JustifyEnd, Gap: 8}, d.cancel, d.ok))
	return giode.Stack(giode.Styles{Gap: 14, Width: -1}, kids...)
}

// Layout lays the dialog out; it takes no space.
func (d *PromptDialog) Layout(gtx giode.Context) giode.Dimensions {
	return d.dlg.Layout(gtx)
}
