package editor

import (
	"fmt"
	"strings"

	"github.com/xmhrnx/giode"

	"litecode/ui/core"
	"litecode/ui/theme"
)

// StatusBar is the line along the bottom of the window: the selected file's
// path as breadcrumbs, the caret position and file properties, and any git
// operation in progress.
type StatusBar struct {
	ide    *core.IDE
	editor *EditorArea
}

// NewStatusBar creates the status bar for an editor area.
func NewStatusBar(ide *core.IDE, editor *EditorArea) *StatusBar {
	return &StatusBar{ide: ide, editor: editor}
}

func (s *StatusBar) crumbs() []giode.Element {
	c := theme.Current()
	p := s.ide.Project
	if p == nil {
		return nil
	}
	parts := []string{p.Name}
	if v := s.editor.Current(); v != nil {
		parts = append(parts, strings.Split(p.Rel(v.Path()), "/")...)
	}
	var kids []giode.Element
	for i, part := range parts {
		if i > 0 {
			kids = append(kids, giode.Text("/", giode.Styles{Color: c.TextFaint, FontSize: theme.SmallSize}))
		}
		color := c.TextMuted
		if i == len(parts)-1 {
			color = c.Text
		}
		kids = append(kids, giode.Text(part, giode.Styles{Color: color, FontSize: theme.SmallSize}))
	}
	return kids
}

func (s *StatusBar) properties() []giode.Element {
	c := theme.Current()
	st := giode.Styles{Color: c.TextMuted, FontSize: theme.SmallSize}
	var kids []giode.Element
	if g := s.ide.Project; g != nil && g.Git != nil && g.Git.Busy != "" {
		kids = append(kids, giode.Spinner(giode.Styles{Width: 12, Height: 12, Color: c.TextMuted}), giode.Text(g.Git.Busy, st))
	}
	v := s.editor.Current()
	if v == nil {
		return kids
	}
	line, col := v.Caret()
	indent := "Tab"
	if u := v.editor.indentUnit(); u != "\t" {
		indent = fmt.Sprintf("%d spaces", len(u))
	}
	eol := "LF"
	if v.doc.CRLF {
		eol = "CRLF"
	}
	for i, t := range []string{fmt.Sprintf("Ln %d, Col %d", line, col), indent, eol, "UTF-8"} {
		if i > 0 {
			kids = append(kids, giode.Box(giode.Styles{Width: 3, Height: 3, BorderRadius: 2, Background: c.TextFaint}))
		}
		kids = append(kids, giode.Text(t, st))
	}
	kids = append(kids, giode.Box(giode.Styles{Background: c.AccentSoft, BorderRadius: 10, Padding: giode.SymmetricInset(9, 2)},
		giode.Text(v.Language().Name, giode.Styles{Color: c.TextAccent, FontSize: 11, FontWeight: giode.FontWeightMedium})))
	return kids
}

// Layout draws the bar.
func (s *StatusBar) Layout(gtx giode.Context) giode.Dimensions {
	left := giode.Row(giode.Styles{Gap: 4, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, s.crumbs()...)
	right := giode.Row(giode.Styles{Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, s.properties()...)
	return giode.Row(giode.Styles{
		Width: -1, Height: theme.StatusBarHeight, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		Padding: giode.SymmetricInset(16, 0),
	}, left, giode.Box(giode.Styles{FlexGrow: 1}), right).Layout(gtx)
}
