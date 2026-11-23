// Package settings is the Settings dialog (Ctrl+Alt+S). Changes apply and
// are saved as they're made.
package settings

import (
	"github.com/xmhrnx/giode"

	"litecode/ui/core"
	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// page is one entry of the settings tree.
type page struct {
	title string
	view  giode.Element
}

// Dialog is the Settings dialog.
type Dialog struct {
	ide      *core.IDE
	open     bool
	pages    []page
	selected int
	nav      []*widgets.ListRow
	close    *giode.ButtonElement
	dlg      *giode.DialogElement
}

// NewDialog creates the dialog. pluginDir is where plugins live; resetLayout
// restores the default tool window layout.
func NewDialog(ide *core.IDE, pluginDir string, resetLayout func()) *Dialog {
	d := &Dialog{ide: ide}
	d.pages = []page{
		{"Appearance", NewAppearancePage(ide, resetLayout)},
		{"Keymap", NewKeymapPage(ide)},
		{"Editor", NewEditorPage(ide)},
		{"On Save", NewSaveActionsPage(ide)},
		{"AI Assistant", NewAIPage(ide)},
		{"Plugins", NewPluginsPage(ide, pluginDir)},
	}
	for i := range d.pages {
		i := i
		d.nav = append(d.nav, widgets.NewListRow(func(giode.ClickEvent) { d.selected = i }))
	}
	d.close = giode.Button("Close", theme.PrimaryButton()).OnClick(func() { d.open = false })
	d.dlg = giode.Dialog(&d.open, d.view, theme.DialogPanel(980)).
		Padding(giode.Inset{})
	return d
}

// Open shows the dialog.
func (d *Dialog) Open() { d.open = true }

// OpenPage shows the dialog on a page by title ("Keymap").
func (d *Dialog) OpenPage(title string) {
	for i, p := range d.pages {
		if p.title == title {
			d.selected = i
		}
	}
	d.open = true
}

func (d *Dialog) view() giode.Element {
	c := theme.Current()
	var nav []giode.Element
	for i, p := range d.pages {
		r := d.nav[i]
		r.Label, r.Selected = p.title, i == d.selected
		nav = append(nav, r)
	}
	return giode.Stack(giode.Styles{Width: -1, Height: 660},
		giode.Row(giode.Styles{Width: -1, FlexGrow: 1, Wrap: giode.WrapNoWrap},
			giode.Stack(giode.Styles{Width: 220, Height: -1, Padding: giode.SymmetricInset(4, 14), Gap: 2, Background: c.Panel}, append([]giode.Element{giode.Box(giode.Styles{Padding: giode.Inset{Left: 20, Bottom: 10}}, giode.Text("Settings", giode.Styles{Color: c.Text, FontSize: 15, FontWeight: giode.FontWeightSemiBold}))}, nav...)...),
			theme.VSeparator(),
			giode.Box(giode.Styles{FlexGrow: 1, Height: -1, Padding: giode.UniformInset(24)}, d.pages[d.selected].view),
		),
		theme.Separator(),
		giode.Row(giode.Styles{Width: -1, Justify: giode.JustifyEnd, Padding: giode.UniformInset(12)}, d.close),
	)
}

// Layout lays the dialog out; it takes no space.
func (d *Dialog) Layout(gtx giode.Context) giode.Dimensions {
	if !d.open {
		d.ide.CapturingKeys = false
	}
	return d.dlg.Layout(gtx)
}
