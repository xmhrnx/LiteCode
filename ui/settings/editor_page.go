package settings

import (
	"strconv"

	"github.com/xmhrnx/giode"

	"litecode/internal/config"
	"litecode/ui/core"
	"litecode/ui/theme"
)

var (
	fontSizes   = []string{"10", "11", "12", "13", "14", "15", "16", "18", "20", "22", "24"}
	lineHeights = []string{"16", "18", "20", "22", "24", "26", "28"}
	tabSizes    = []string{"2", "4", "8"}
)

// EditorPage holds the editor font and indentation settings.
type EditorPage struct {
	ide        *core.IDE
	size       *giode.SelectElement
	lineHeight *giode.SelectElement
	tabSize    *giode.SelectElement
}

// NewEditorPage creates the page.
func NewEditorPage(ide *core.IDE) *EditorPage {
	p := &EditorPage{ide: ide}
	st := theme.InputStyles()
	st.Width = 90
	p.size = selectOf(st, fontSizes, ide.Settings.EditorFontSize, func(n int) { ide.Settings.EditorFontSize = n; p.save() })
	p.lineHeight = selectOf(st, lineHeights, ide.Settings.EditorLineHeight, func(n int) { ide.Settings.EditorLineHeight = n; p.save() })
	p.tabSize = selectOf(st, tabSizes, ide.Settings.TabSize, func(n int) { ide.Settings.TabSize = n; p.save() })
	return p
}

// selectOf makes a dropdown of numbers that reports the chosen one.
func selectOf(st giode.Styles, options []string, current int, set func(int)) *giode.SelectElement {
	s := giode.Input.Select(st, options...)
	for i, o := range options {
		if o == strconv.Itoa(current) {
			s.SetSelected(i)
		}
	}
	return s.OnChange(func(i int) {
		n, _ := strconv.Atoi(options[i])
		set(n)
	})
}

func (p *EditorPage) save() {
	p.ide.Error(config.SaveSettings(p.ide.Settings))
}

// Layout draws the page.
func (p *EditorPage) Layout(gtx giode.Context) giode.Dimensions {
	return giode.Stack(giode.Styles{Width: -1, Gap: 16},
		pageTitle("Editor"),
		settingRow("Font", giode.Text(theme.CodeFont, giode.Styles{Color: theme.Current().TextMuted, FontSize: theme.UISize}), ""),
		settingRow("Size", p.size, ""),
		settingRow("Line height", p.lineHeight, ""),
		settingRow("Tab size", p.tabSize, "Used for languages that don't set their own indentation."),
	).Layout(gtx)
}
