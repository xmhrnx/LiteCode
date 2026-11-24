package editor

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"

	"litecode/internal/highlight"
	"litecode/internal/lsp"
	"litecode/ui/theme"
)

// Sizes of the completion list, in dp.
const (
	listWidth = 440
	listRowH  = 24
)

// completionList draws the open completion list under the caret: a kind
// badge, the name and its detail per row, the selected row tinted. A
// click takes a completion.
type completionList struct {
	e    *TextEditor
	rows [popupRows]*giode.PressableElement
}

func newCompletionList(e *TextEditor) *completionList {
	l := &completionList{e: e}
	for i := range l.rows {
		l.rows[i] = giode.Pressable(func(st giode.PressState) giode.Element { return l.row(i, st) }).
			Focusable(false).
			OnPrimaryClick(func(giode.ClickEvent) {
				if c := e.comp; c != nil {
					e.accept(c.first + i)
				}
			})
	}
	return l
}

// badge is a completion kind's letter and color.
func badge(k lsp.CompletionKind) (string, string) {
	cc := theme.CodeScheme()
	switch k {
	case lsp.KindMethod, lsp.KindFunction, lsp.KindConstructor:
		return "f", cc.Tokens[highlight.Function]
	case lsp.KindField, lsp.KindProperty:
		return "p", cc.Tokens[highlight.Property]
	case lsp.KindVariable, lsp.KindValue, lsp.KindReference:
		return "v", cc.Tokens[highlight.Property]
	case lsp.KindClass, lsp.KindStruct, lsp.KindInterface, lsp.KindEnum, lsp.KindTypeParameter:
		return "T", cc.Tokens[highlight.Type]
	case lsp.KindModule, lsp.KindFolder, lsp.KindFile:
		return "m", cc.Tokens[highlight.String]
	case lsp.KindKeyword, lsp.KindOperator:
		return "k", cc.Tokens[highlight.Keyword]
	case lsp.KindConstant, lsp.KindEnumMember, lsp.KindUnit:
		return "c", cc.Tokens[highlight.Constant]
	case lsp.KindSnippet:
		return "s", cc.Tokens[highlight.Tag]
	}
	return "w", theme.Current().TextFaint
}

func (l *completionList) row(slot int, st giode.PressState) giode.Element {
	c := l.e.comp
	if c == nil || c.first+slot >= len(c.shown) {
		return giode.Box(giode.Styles{})
	}
	i := c.first + slot
	it := c.items[c.shown[i].Index]
	sc := theme.Current()
	bg := "transparent"
	switch {
	case i == c.sel:
		bg = theme.CodeScheme().Popup
	case st.Hovered:
		bg = sc.Hover
	}
	letter, color := badge(it.Kind)
	return giode.Row(giode.Styles{Width: -1, Height: listRowH, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		Padding: giode.SymmetricInset(6, 0), BorderRadius: 5, Background: bg},
		giode.Box(giode.Styles{Width: 16, Height: 16, BorderRadius: 4, Background: giode.Mix(color, sc.Raised, 0.18), Align: giode.AlignCenter, Justify: giode.JustifyCenter},
			giode.Text(letter, giode.Styles{Color: color, FontSize: 10, FontFamily: theme.CodeFont, FontWeight: giode.FontWeightBold})),
		giode.Text(it.Label, giode.Styles{Color: sc.Text, FontFamily: theme.CodeFont, FontSize: properties.FontSize(l.e.m.fontSize), MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis}),
		giode.Box(giode.Styles{FlexGrow: 1}),
		giode.Text(it.Detail, giode.Styles{Color: sc.TextFaint, FontSize: theme.SmallSize, MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis, MaxWidth: 200}),
	)
}

// height is the list's height in dp for what it shows.
func (l *completionList) height() float32 {
	n := 1
	if c := l.e.comp; c != nil {
		n = max(1, min(popupRows, len(c.shown)))
	}
	return float32(n*listRowH + 8)
}

// Layout draws the rows in view, or "Loading…" while the server answers.
func (l *completionList) Layout(gtx giode.Context) giode.Dimensions {
	sc := theme.Current()
	c := l.e.comp
	var kids []giode.Element
	if c != nil && len(c.shown) == 0 {
		kids = append(kids, giode.Box(giode.Styles{Height: listRowH, Padding: giode.SymmetricInset(8, 0), Justify: giode.JustifyCenter},
			giode.Text("Loading suggestions…", giode.Styles{Color: sc.TextFaint, FontSize: theme.SmallSize})))
	}
	for i := 0; c != nil && i < popupRows && c.first+i < len(c.shown); i++ {
		kids = append(kids, l.rows[i])
	}
	return giode.Stack(giode.Styles{Width: listWidth, Padding: giode.UniformInset(4), Background: sc.Raised, BorderRadius: 8,
		BorderWidth: 1, BorderColor: sc.PopupLine, BoxShadow: giode.BoxShadow{Y: 8, Blur: 24, Color: sc.Shadow}}, kids...).Layout(gtx)
}
