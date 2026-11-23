package settings

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/checkbox"

	"litecode/internal/config"
	"litecode/internal/format"
	"litecode/ui/core"
	"litecode/ui/theme"
)

// SaveActionsPage is Settings > On Save. These settings are global:
// Prettier runs in every project without per-project setup.
type SaveActionsPage struct {
	ide      *core.IDE
	reformat *checkbox.Checkbox
	prettier *checkbox.Checkbox
}

// NewSaveActionsPage creates the page.
func NewSaveActionsPage(ide *core.IDE) *SaveActionsPage {
	p := &SaveActionsPage{ide: ide}
	st := giode.Styles{Background: theme.Current().Accent, Color: theme.Current().Text}
	p.reformat = giode.Input.Checkbox("Reformat code on save (Ctrl+S)", st).OnChange(func(on bool) {
		ide.Settings.FormatOnSave = on
		p.save()
	})
	p.reformat.SetChecked(ide.Settings.FormatOnSave)
	p.prettier = giode.Input.Checkbox("Use Prettier for JavaScript, TypeScript, CSS, HTML, JSON, Markdown and YAML", st).OnChange(func(on bool) {
		ide.Settings.UsePrettier = on
		p.save()
	})
	p.prettier.SetChecked(ide.Settings.UsePrettier)
	return p
}

func (p *SaveActionsPage) save() {
	p.ide.Error(config.SaveSettings(p.ide.Settings))
}

func (p *SaveActionsPage) prettierStatus() string {
	if p.ide.Project == nil {
		return "Prettier is looked up in each project's node_modules, then on PATH."
	}
	if bin := format.FindPrettier(p.ide.Project.Root); bin != "" {
		return "This project uses " + bin
	}
	return "Prettier wasn't found in this project's node_modules or on PATH; install it with npm i -g prettier."
}

// Layout draws the page.
func (p *SaveActionsPage) Layout(gtx giode.Context) giode.Dimensions {
	c := theme.Current()
	note := func(s string) giode.Element {
		return giode.Box(giode.Styles{Padding: giode.Inset{Left: 28}}, giode.Text(s, giode.Styles{Color: c.TextFaint, FontSize: theme.SmallSize}))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 12},
		pageTitle("On Save"),
		giode.Text("These apply to every project.", giode.Styles{Color: c.TextMuted, FontSize: theme.UISize}),
		p.reformat,
		note("Languages without Prettier use their own formatter: gofmt for Go, rustfmt for Rust, or what a plugin declares."),
		p.prettier,
		note(p.prettierStatus()),
	).Layout(gtx)
}
