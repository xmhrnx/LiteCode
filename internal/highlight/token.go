// Package highlight splits lines of code into colored tokens. It's a small
// rule-driven lexer: comments, strings, numbers, keywords, types, calls and
// a few per-language forms (tags, keys, headings). Each line's starting
// state is cached, so after an edit only the lines from it on are redone,
// and only as far as they're drawn.
package highlight

// Kind is what a token is, which picks its color.
type Kind uint8

// The token kinds.
const (
	Plain Kind = iota
	Keyword
	Type
	Function
	String
	Number
	Comment
	Constant
	Punct
	Property // a key in JSON, YAML or TOML; an attribute in HTML
	Tag      // an HTML tag name
	Heading  // a Markdown heading
)

// Token is a run of one kind on a line, in rune columns [Start, End).
type Token struct {
	Start, End int
	Kind       Kind
}

// state is what carries from one line to the next: an open block comment
// or a string that goes on.
type state struct {
	comment bool   // in a block comment
	str     string // in a multi-line string ending with this
}

// Rules describe one language's lexing.
type Rules struct {
	LineComment  []string
	BlockComment [2]string
	Quotes       string   // runes that open single-line strings
	MultiQuotes  []string // delimiters of strings that span lines: "`", `"""`
	Keywords     map[string]bool
	Constants    map[string]bool
	Types        map[string]bool
	CapitalTypes bool // capitalised names are types
	Tags         bool // <name ...> markup
	Keys         bool // name: or "name": at the start is a key
	Headings     bool // # heading lines
	HashNotes    bool // # starts a comment only at a word boundary (shell)
	Variables    bool // $name
	Decorators   bool // @name
}
