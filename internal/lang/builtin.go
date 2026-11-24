package lang

var prettierCmd = &Command{Args: []string{"prettier", "--stdin-filepath", "${file}"}, Stdin: true}

// Builtin is the set of languages LiteCode knows without plugins.
func Builtin() []*Language {
	cStyle := [2]string{"/*", "*/"}
	return []*Language{
		{ID: "go", Name: "Go", Extensions: []string{".go"}, LineComment: "//", BlockComment: cStyle, IndentTabs: true,
			Formatter: &Command{Args: []string{"gofmt"}, Stdin: true}},
		{ID: "gomod", Name: "Go Module", FileNames: []string{"go.mod", "go.work"}, LineComment: "//", IndentTabs: true},
		{ID: "typescript", Name: "TypeScript", Extensions: []string{".ts", ".tsx", ".mts", ".cts"}, LineComment: "//", BlockComment: cStyle, IndentSize: 2, Prettier: true},
		{ID: "javascript", Name: "JavaScript", Extensions: []string{".js", ".jsx", ".mjs", ".cjs"}, LineComment: "//", BlockComment: cStyle, IndentSize: 2, Prettier: true},
		{ID: "json", Name: "JSON", Extensions: []string{".json", ".jsonc"}, IndentSize: 2, Prettier: true},
		{ID: "css", Name: "CSS", Extensions: []string{".css", ".scss", ".less"}, BlockComment: cStyle, LineComment: "//", IndentSize: 2, Prettier: true},
		{ID: "html", Name: "HTML", Extensions: []string{".html", ".htm", ".vue", ".svelte"}, BlockComment: [2]string{"<!--", "-->"}, IndentSize: 2, Prettier: true},
		{ID: "markdown", Name: "Markdown", Extensions: []string{".md", ".markdown"}, BlockComment: [2]string{"<!--", "-->"}, IndentSize: 2, Prettier: true},
		{ID: "yaml", Name: "YAML", Extensions: []string{".yml", ".yaml"}, LineComment: "#", IndentSize: 2, Prettier: true},
		{ID: "toml", Name: "TOML", Extensions: []string{".toml"}, LineComment: "#", IndentSize: 2},
		{ID: "python", Name: "Python", Extensions: []string{".py", ".pyi"}, LineComment: "#", IndentSize: 4},
		{ID: "rust", Name: "Rust", Extensions: []string{".rs"}, LineComment: "//", BlockComment: cStyle, IndentSize: 4,
			Formatter: &Command{Args: []string{"rustfmt", "--emit", "stdout"}, Stdin: true}},
		{ID: "shell", Name: "Shell Script", Extensions: []string{".sh", ".bash", ".zsh"}, FileNames: []string{".bashrc", ".zshrc"}, LineComment: "#", IndentSize: 2},
		{ID: "dockerfile", Name: "Dockerfile", FileNames: []string{"Dockerfile"}, LineComment: "#", IndentSize: 4},
		{ID: "makefile", Name: "Makefile", FileNames: []string{"Makefile", "GNUmakefile"}, LineComment: "#", IndentTabs: true},
		{ID: "java", Name: "Java", Extensions: []string{".java"}, LineComment: "//", BlockComment: cStyle, IndentSize: 4},
		{ID: "kotlin", Name: "Kotlin", Extensions: []string{".kt", ".kts"}, LineComment: "//", BlockComment: cStyle, IndentSize: 4},
		{ID: "c", Name: "C/C++", Extensions: []string{".c", ".h", ".cc", ".cpp", ".hpp"}, LineComment: "//", BlockComment: cStyle, IndentSize: 4},
		{ID: "sql", Name: "SQL", Extensions: []string{".sql"}, LineComment: "--", BlockComment: cStyle, IndentSize: 2},
		{ID: "gitignore", Name: "Ignore file", FileNames: []string{".gitignore", ".dockerignore", ".prettierignore"}, LineComment: "#"},
	}
}
