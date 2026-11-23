package settings

import (
	"github.com/xmhrnx/giode"

	"litecode/ui/core"
	"litecode/ui/theme"
)

// AppearancePage shows the theme and fonts and resets the tool window layout.
type AppearancePage struct {
	ide         *core.IDE
	resetLayout *giode.ButtonElement
}

// NewAppearancePage creates the page; resetLayout puts the tool windows back
// where they start.
func NewAppearancePage(ide *core.IDE, resetLayout func()) *AppearancePage {
	return &AppearancePage{
		ide:         ide,
		resetLayout: giode.Button("Restore Default Layout", theme.SecondaryButton()).Variant(giode.VariantOutlined).OnClick(resetLayout),
	}
}

// Layout draws the page.
func (p *AppearancePage) Layout(gtx giode.Context) giode.Dimensions {
	c := theme.Current()
	value := func(s string) giode.Element {
		return giode.Text(s, giode.Styles{Color: c.TextMuted, FontSize: theme.UISize})
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 16},
		pageTitle("Appearance"),
		settingRow("Theme", value("LiteCode Dark"), ""),
		settingRow("UI font", value(theme.UIFont+" 13"), ""),
		settingRow("Panels", p.resetLayout, "Moves every panel back to its default side and size."),
	).Layout(gtx)
}
