package theme

import "litecode/internal/highlight"

// CodeColors are the editor's own colors: its text by token kind, and
// the caret, selection, current line, gutter and guides.
type CodeColors struct {
	Tokens      [highlight.Heading + 1]string
	Caret       string
	Selection   string
	CurrentLine string
	LineNumber  string
	LineActive  string // the caret line's number
	Guide       string // indent guides
	Composition string // underline of text an input method is composing
}

// darkCode goes with darkScheme: soft pastels around the violet accent.
var darkCode = CodeColors{
	Tokens: [...]string{
		highlight.Plain:    "#D3D5DC",
		highlight.Keyword:  "#C792EA",
		highlight.Type:     "#7FD6C2",
		highlight.Function: "#82AAFF",
		highlight.String:   "#A5D88A",
		highlight.Number:   "#F2A97A",
		highlight.Comment:  "#6C7183",
		highlight.Constant: "#F2A97A",
		highlight.Punct:    "#9BA1B4",
		highlight.Property: "#8FC5FF",
		highlight.Tag:      "#F07A93",
		highlight.Heading:  "#B1A6FF",
	},
	Caret:       "#A99BFF",
	Selection:   "#3A3470",
	CurrentLine: "#1A1B22",
	LineNumber:  "#4B4F5C",
	LineActive:  "#B6B9C4",
	Guide:       "#23252E",
	Composition: "#A99BFF",
}

// CodeScheme returns the active editor colors.
func CodeScheme() *CodeColors { return &darkCode }
