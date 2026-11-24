package settings

import (
	"github.com/xmhrnx/giode"

	"litecode/internal/config"
	"litecode/ui/core"
	"litecode/ui/theme"
)

// mainColors are the interface colors listed before "Show all colors".
var mainColors = []string{"Accent", "Editor", "Panel", "Raised", "Text", "TextMuted", "Success", "Warning", "Danger", "Selection"}

// editorColors are the editor colors besides the token kinds.
var editorColors = []string{"Caret", "Selection", "CurrentLine", "LineNumber", "LineActive", "Guide"}

// AppearancePage picks the UI and syntax themes, changes single colors,
// and resets the panel layout.
type AppearancePage struct {
	ide         *core.IDE
	scroll      *giode.ScrollElement
	uiCards     []*themeCard
	syntaxCards []*syntaxCard
	uiRows      map[string]*colorRow
	codeRows    []*colorRow
	showAll     *giode.CheckboxElement
	resetColors *giode.ButtonElement
	resetLayout *giode.ButtonElement
}

// NewAppearancePage creates the page; resetLayout puts the panels back
// where they start.
func NewAppearancePage(ide *core.IDE, resetLayout func()) *AppearancePage {
	p := &AppearancePage{ide: ide, uiRows: map[string]*colorRow{}}
	for i := range theme.UIThemes {
		p.uiCards = append(p.uiCards, newThemeCard(&theme.UIThemes[i], p.pickUI))
	}
	p.syntaxCards = append(p.syntaxCards, newSyntaxCard(nil, p.pickSyntax))
	for i := range theme.SyntaxThemes {
		p.syntaxCards = append(p.syntaxCards, newSyntaxCard(&theme.SyntaxThemes[i], p.pickSyntax))
	}
	for _, name := range theme.SchemeFields() {
		p.uiRows[name] = newColorRow(spaced(name),
			func() string { return theme.SchemeValue(theme.Values(), name) },
			func() bool { _, ok := ide.Settings.UIColors[name]; return ok },
			func(c string) { p.setColor(&ide.Settings.UIColors, name, c) })
	}
	for k, name := range theme.TokenNames {
		p.codeRows = append(p.codeRows, newColorRow(name,
			func() string { return theme.CodeValues().Tokens[k] },
			func() bool { _, ok := ide.Settings.CodeColors[name]; return ok },
			func(c string) { p.setColor(&ide.Settings.CodeColors, name, c) }))
	}
	for _, name := range editorColors {
		p.codeRows = append(p.codeRows, newColorRow(spaced(name),
			func() string { return codeValue(name) },
			func() bool { _, ok := ide.Settings.CodeColors[name]; return ok },
			func(c string) { p.setColor(&ide.Settings.CodeColors, name, c) }))
	}
	p.showAll = giode.Input.Checkbox("Show all interface colors")
	p.resetColors = giode.Button("Reset All Colors", theme.SecondaryButton()).Variant(giode.VariantOutlined).OnClick(func() {
		ide.Settings.UIColors, ide.Settings.CodeColors = nil, nil
		p.save()
	})
	p.resetLayout = giode.Button("Restore Default Layout", theme.SecondaryButton()).Variant(giode.VariantOutlined).OnClick(resetLayout)
	p.scroll = giode.Scroll(p.view, giode.Styles{ScrollBar: theme.ScrollBar()})
	return p
}

// codeValue is an editor color by field name.
func codeValue(name string) string {
	v := theme.CodeValues()
	switch name {
	case "Caret":
		return v.Caret
	case "Selection":
		return v.Selection
	case "CurrentLine":
		return v.CurrentLine
	case "LineNumber":
		return v.LineNumber
	case "LineActive":
		return v.LineActive
	case "Guide":
		return v.Guide
	}
	return ""
}

// spaced turns a field name into words: "TextMuted" is "Text muted".
func spaced(name string) string {
	out := []rune{}
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			out = append(out, ' ', r+('a'-'A'))
			continue
		}
		out = append(out, r)
	}
	return string(out)
}

// pickUI switches the UI theme. Colors changed by hand were made for the
// old one, so they go.
func (p *AppearancePage) pickUI(id string) {
	s := &p.ide.Settings
	s.Theme, s.UIColors = id, nil
	if s.SyntaxTheme != "" && theme.SyntaxByID(s.SyntaxTheme).Dark != theme.UIThemeByID(id).Dark {
		// A dark syntax theme on a light window can't be read: follow the theme.
		s.SyntaxTheme, s.CodeColors = "", nil
	}
	p.save()
}

// pickSyntax switches the syntax theme; "" follows the UI theme.
func (p *AppearancePage) pickSyntax(id string) {
	p.ide.Settings.SyntaxTheme, p.ide.Settings.CodeColors = id, nil
	p.save()
}

// setColor changes one color by hand; "" resets it.
func (p *AppearancePage) setColor(m *map[string]string, name, c string) {
	if c == "" {
		delete(*m, name)
	} else {
		if *m == nil {
			*m = map[string]string{}
		}
		(*m)[name] = c
	}
	p.save()
}

// save applies the settings' theme and writes them.
func (p *AppearancePage) save() {
	p.ide.ApplyTheme()
	p.ide.Error(config.SaveSettings(p.ide.Settings))
}

// section is a heading with a line of help under it.
func section(title, help string) giode.Element {
	c := theme.Current()
	kids := []giode.Element{giode.Text(title, giode.Styles{Color: c.Text, FontSize: theme.UISize, FontWeight: giode.FontWeightSemiBold})}
	if help != "" {
		kids = append(kids, giode.Text(help, giode.Styles{Color: c.TextFaint, FontSize: theme.SmallSize}))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 3, Padding: giode.Inset{Top: 8}}, kids...)
}

func (p *AppearancePage) view() giode.Element {
	var dark, light, syntax []giode.Element
	for _, k := range p.uiCards {
		if k.t.Dark {
			dark = append(dark, k)
		} else {
			light = append(light, k)
		}
	}
	uiDark := theme.IsDark()
	for _, k := range p.syntaxCards {
		if k.t == nil || k.t.Dark == uiDark {
			syntax = append(syntax, k)
		}
	}
	grid := func(kids []giode.Element) giode.Element {
		return giode.Row(giode.Styles{Width: -1, Gap: 10}, kids...)
	}
	var uiRows []giode.Element
	names := mainColors
	if p.showAll.Checked() {
		names = theme.SchemeFields()
	}
	for _, n := range names {
		uiRows = append(uiRows, p.uiRows[n])
	}
	codeRows := make([]giode.Element, len(p.codeRows))
	for i, r := range p.codeRows {
		codeRows[i] = r
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 14, Padding: giode.Inset{Right: 14, Bottom: 16}},
		pageTitle("Appearance"),
		section("Dark themes", ""),
		grid(dark),
		section("Light themes", ""),
		grid(light),
		section("Code colors", "Syntax themes for the kind of theme in use. Picking a theme clears the colors changed by hand below."),
		grid(syntax),
		section("Interface colors", "Click a color to change it. The accent, backgrounds and text also recolor what is made from them."),
		giode.Stack(giode.Styles{Width: -1, Gap: 2}, uiRows...),
		p.showAll,
		section("Editor colors", ""),
		giode.Stack(giode.Styles{Width: -1, Gap: 2}, codeRows...),
		p.resetColors,
		section("Fonts and panels", ""),
		settingRow("UI font", giode.Text(theme.UIFont+" 13", giode.Styles{Color: theme.Current().TextMuted, FontSize: theme.UISize}), ""),
		settingRow("Panels", p.resetLayout, "Moves every panel back to its default side and size."),
	)
}

// Layout draws the page.
func (p *AppearancePage) Layout(gtx giode.Context) giode.Dimensions {
	return giode.Box(giode.Styles{Width: -1, Height: -1}, p.scroll).Layout(gtx)
}
