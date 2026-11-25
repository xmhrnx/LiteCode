package lang

import (
	"path/filepath"
	"strings"
)

// servers are the language servers LiteCode tries for the built-in
// languages, in order; the first one installed runs.
var servers = map[string][][]string{
	"go":         {{"gopls"}},
	"gomod":      {{"gopls"}},
	"typescript": tsServers,
	"javascript": tsServers,
	"python":     {{"pyright-langserver", "--stdio"}, {"basedpyright-langserver", "--stdio"}, {"pylsp"}, {"jedi-language-server"}},
	"rust":       {{"rust-analyzer"}},
	"c":          {{"clangd"}},
	"java":       {{"jdtls"}},
	"kotlin":     {{"kotlin-language-server"}},
	"css":        {{"vscode-css-language-server", "--stdio"}},
	"html":       {{"vscode-html-language-server", "--stdio"}},
	"json":       {{"vscode-json-language-server", "--stdio"}},
	"yaml":       {{"yaml-language-server", "--stdio"}},
	"toml":       {{"taplo", "lsp", "stdio"}},
	"shell":      {{"bash-language-server", "start"}},
	"dockerfile": {{"docker-langserver", "--stdio"}},
	"markdown":   {{"marksman", "server"}},
	"sql":        {{"sqls"}},
}

// tsServers serve TypeScript and JavaScript: the classic server (which
// needs TypeScript 5 or earlier), then TypeScript 7's own.
var tsServers = [][]string{
	{"typescript-language-server", "--stdio"}, {"vtsls", "--stdio"},
	{"tsc", "--lsp", "--stdio"}, {"tsgo", "--lsp", "--stdio"},
}

// ServerCommands are the language servers to try for the language: the
// one a plugin set, then the built-in choices.
func (l *Language) ServerCommands() [][]string {
	var out [][]string
	if l.LSP != nil && len(l.LSP.Args) > 0 {
		out = append(out, l.LSP.Args)
	}
	return append(out, servers[l.ID]...)
}

// ServerLanguageID is the language's name for a file as language servers
// know it ("typescriptreact" for a .tsx file).
func (l *Language) ServerLanguageID(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch {
	case ext == ".tsx":
		return "typescriptreact"
	case ext == ".jsx":
		return "javascriptreact"
	case ext == ".scss" || ext == ".less":
		return ext[1:]
	case l.ID == "c" && ext != ".c" && ext != ".h":
		return "cpp"
	case l.ID == "shell":
		return "shellscript"
	case l.ID == "gomod":
		return "go.mod"
	}
	return l.ID
}
