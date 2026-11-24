package settings

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"

	"litecode/internal/highlight"
	"litecode/ui/theme"
)

// syntaxCard is a syntax theme to pick: a few lines of code in its
// colors on the editor's background, and its name. A nil theme is "match
// the UI theme".
type syntaxCard struct {
	t     *theme.SyntaxTheme
	press *giode.PressableElement
}

func newSyntaxCard(t *theme.SyntaxTheme, pick func(id string)) *syntaxCard {
	k := &syntaxCard{t: t}
	id, name := "", "Match theme"
	if t != nil {
		id, name = t.ID, t.Name
	}
	k.press = giode.Pressable(k.view).Focusable(false).Label(name).
		OnPrimaryClick(func(giode.ClickEvent) { pick(id) })
	return k
}

// sample is code in a syntax theme's colors.
func sample(tokens [highlight.Heading + 1]string) giode.Element {
	st := func(k highlight.Kind) giode.Styles {
		return giode.Styles{Color: tokens[k], FontFamily: theme.CodeFont, FontSize: 11}
	}
	base := giode.Styles{FontFamily: theme.CodeFont, FontSize: 11, MaxLines: 1}
	return giode.Stack(giode.Styles{Gap: 2},
		giode.RichText(base, giode.Span("func ", st(highlight.Keyword)), giode.Span("greet", st(highlight.Function)),
			giode.Span("(n ", st(highlight.Plain)), giode.Span("int", st(highlight.Type)), giode.Span(") {", st(highlight.Punct))),
		giode.RichText(base, giode.Span("  s := ", st(highlight.Plain)), giode.Span(`"hi"`, st(highlight.String)),
			giode.Span(" + ", st(highlight.Punct)), giode.Span("42", st(highlight.Number))),
		giode.RichText(base, giode.Span("  // done", st(highlight.Comment))),
	)
}

func (k *syntaxCard) view(st giode.PressState) giode.Element {
	c, v := theme.Current(), theme.Values()
	id, name := "", "Match theme"
	t := k.t
	if t != nil {
		id, name = t.ID, t.Name
	} else {
		t = theme.SyntaxByID(theme.UIThemeByID(theme.Active().UI).Syntax)
	}
	edge, width := c.Separator, 1
	if theme.Active().Syntax == id {
		edge, width = c.FocusLine, 2
	} else if st.Hovered {
		edge = c.PopupLine
	}
	return giode.Stack(giode.Styles{Width: 168, Gap: 6, Padding: giode.UniformInset(6), BorderRadius: 12, BorderWidth: properties.BorderWidth(width), BorderColor: edge, Cursor: giode.CursorPointer},
		giode.Box(giode.Styles{Width: -1, Background: v.Editor, BorderRadius: 8, Padding: giode.UniformInset(8), BorderWidth: 1, BorderColor: c.CardLine}, sample(t.Tokens)),
		giode.Text(name, giode.Styles{Color: c.Text, FontSize: theme.SmallSize, MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis}),
	)
}

// Layout draws the card.
func (k *syntaxCard) Layout(gtx giode.Context) giode.Dimensions { return k.press.Layout(gtx) }
