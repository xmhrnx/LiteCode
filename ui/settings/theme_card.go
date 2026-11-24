package settings

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"

	"litecode/internal/highlight"
	"litecode/ui/theme"
)

// themeCard is a UI theme to pick: a small picture of the window in its
// colors, and its name.
type themeCard struct {
	t     *theme.UITheme
	press *giode.PressableElement
}

func newThemeCard(t *theme.UITheme, pick func(id string)) *themeCard {
	k := &themeCard{t: t}
	k.press = giode.Pressable(k.view).Focusable(false).Label(t.Name).
		OnPrimaryClick(func(giode.ClickEvent) { pick(t.ID) })
	return k
}

// preview draws the theme's window in miniature: a side panel, the
// editor with a few lines of code, and an accent button.
func previewOf(s *theme.Scheme, tokens [highlight.Heading + 1]string) giode.Element {
	bar := func(w int, color string) giode.Element {
		return giode.Box(giode.Styles{Width: properties.Width(w), Height: 4, BorderRadius: 2, Background: color})
	}
	line := func(kids ...giode.Element) giode.Element {
		return giode.Row(giode.Styles{Gap: 3, Wrap: giode.WrapNoWrap}, kids...)
	}
	side := giode.Stack(giode.Styles{Width: 30, Height: -1, Background: s.Panel, Gap: 5, Padding: giode.UniformInset(6)},
		bar(16, s.TextMuted), bar(12, s.TextFaint), bar(14, s.TextFaint))
	code := giode.Stack(giode.Styles{FlexGrow: 1, Height: -1, Background: s.Editor, Gap: 5, Padding: giode.UniformInset(7)},
		line(bar(12, tokens[highlight.Keyword]), bar(18, tokens[highlight.Function])),
		line(bar(8, tokens[highlight.Plain]), bar(22, tokens[highlight.String])),
		line(bar(26, tokens[highlight.Comment])),
		giode.Box(giode.Styles{Width: 20, Height: 7, BorderRadius: 4, Background: s.Accent}),
	)
	return giode.Row(giode.Styles{Width: -1, Height: 64, Wrap: giode.WrapNoWrap, BorderRadius: 8, Overflow: giode.OverflowHidden,
		BorderWidth: 1, BorderColor: s.CardLine}, side, code)
}

func (k *themeCard) view(st giode.PressState) giode.Element {
	c := theme.Current()
	selected := theme.Active().UI == k.t.ID || theme.Active().UI == "" && k.t.ID == theme.UIThemes[0].ID
	edge, width := c.Separator, 1
	if selected {
		edge, width = c.FocusLine, 2
	} else if st.Hovered {
		edge = c.PopupLine
	}
	syn := theme.SyntaxByID(k.t.Syntax)
	return giode.Stack(giode.Styles{Width: 132, Gap: 6, Padding: giode.UniformInset(6), BorderRadius: 12, BorderWidth: properties.BorderWidth(width), BorderColor: edge, Cursor: giode.CursorPointer},
		previewOf(&k.t.Scheme, syn.Tokens),
		giode.Text(k.t.Name, giode.Styles{Color: c.Text, FontSize: theme.SmallSize, MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis}),
	)
}

// Layout draws the card.
func (k *themeCard) Layout(gtx giode.Context) giode.Dimensions { return k.press.Layout(gtx) }
