package search

import (
	"strings"

	"github.com/xmhrnx/giode"

	isearch "litecode/internal/search"
	"litecode/ui/theme"
)

// matchLine draws a matching line with the match highlighted, leading
// whitespace dropped.
func matchLine(m isearch.TextMatch) giode.Element {
	c := theme.Current()
	runes := []rune(m.Text)
	start := min(m.Col, len(runes))
	end := min(m.Col+m.Len, len(runes))
	before := strings.TrimLeft(string(runes[:start]), " \t")
	code := giode.Styles{FontFamily: theme.CodeFont, FontSize: 12}
	return giode.RichText(giode.Styles{Color: c.TextCode, FontFamily: theme.CodeFont, FontSize: 12, MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis},
		giode.Span(before, code),
		giode.Span(string(runes[start:end]), giode.Styles{FontFamily: theme.CodeFont, FontSize: 12, Background: c.Match, Color: c.Text}),
		giode.Span(string(runes[end:]), code),
	)
}
