// Package lang describes the languages LiteCode knows: which files they cover,
// how they comment and indent, how to format them, and (reserved for when the
// editor supports it) which language server to start. Languages come from the
// built-in set and from plugins: folders with a plugin.json under
// <config>/litecode/plugins.
package lang

import (
	"path/filepath"
	"strings"
)

// Command is an external program, with ${file} replaced by the file's path
// and ${root} by the project root. With Stdin the text is piped in and the
// result read from stdout.
type Command struct {
	Args  []string `json:"args"`
	Stdin bool     `json:"stdin"`
}

// Language is one language's description.
type Language struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Extensions   []string  `json:"extensions"` // with the dot: ".go"
	FileNames    []string  `json:"fileNames"`  // exact names: "Dockerfile"
	LineComment  string    `json:"lineComment"`
	BlockComment [2]string `json:"blockComment"`
	IndentTabs   bool      `json:"indentTabs"`
	IndentSize   int       `json:"indentSize"`
	Prettier     bool      `json:"prettier"`  // formatted by Prettier when it's enabled
	Formatter    *Command  `json:"formatter"` // used when Prettier doesn't apply
	LSP          *Command  `json:"lsp"`
	Keywords     []string  `json:"keywords"`
}

// Matches reports whether path is a file of this language.
func (l *Language) Matches(path string) bool {
	base := filepath.Base(path)
	for _, n := range l.FileNames {
		if strings.EqualFold(n, base) {
			return true
		}
	}
	ext := strings.ToLower(filepath.Ext(base))
	for _, e := range l.Extensions {
		if e == ext {
			return true
		}
	}
	return false
}

// Indent is one level of indentation for the language.
func (l *Language) Indent(tabSize int) string {
	if l.IndentTabs {
		return "\t"
	}
	n := l.IndentSize
	if n == 0 {
		n = tabSize
	}
	return strings.Repeat(" ", n)
}

// PlainText is the language of files nothing else claims.
var PlainText = &Language{ID: "text", Name: "Plain text", IndentSize: 4}
