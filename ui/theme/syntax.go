package theme

import "litecode/internal/highlight"

// CodeColors are the editor's own colors: its text by token kind, the
// caret, selection, current line, gutter and guides, and the terminal's 16
// colors.
type CodeColors struct {
	Tokens      [highlight.Heading + 1]string
	Caret       string
	Selection   string
	CurrentLine string
	LineNumber  string
	LineActive  string // the caret line's number
	Guide       string // indent guides
	Composition string // underline of text an input method is composing
	Popup       string // the completion list's selected row
	ANSI        [16]string
}

// TokenNames name the token kinds in settings and on the Appearance page.
var TokenNames = [highlight.Heading + 1]string{
	highlight.Plain: "Text", highlight.Keyword: "Keyword", highlight.Type: "Type",
	highlight.Function: "Function", highlight.String: "String", highlight.Number: "Number",
	highlight.Comment: "Comment", highlight.Constant: "Constant", highlight.Punct: "Punctuation",
	highlight.Property: "Property", highlight.Tag: "Tag", highlight.Heading: "Heading",
}

// SyntaxTheme is a ready-made set of token and terminal colors. The
// editor's background comes from the UI theme; a syntax theme is made for
// a dark or a light one.
type SyntaxTheme struct {
	ID, Name string
	Dark     bool
	// text, keyword, type, function, string, number, comment, constant,
	// punctuation, property, tag, heading
	Tokens [highlight.Heading + 1]string
	ANSI   [16]string
}

// syn makes a syntax theme from its token colors in TokenNames order.
func syn(id, name string, dark bool, ansi [16]string, tokens ...string) SyntaxTheme {
	t := SyntaxTheme{ID: id, Name: name, Dark: dark, ANSI: ansi}
	copy(t.Tokens[:], tokens)
	return t
}

// Terminal palettes, black to white then their bright versions.
var (
	ansiLiteDark  = [16]string{"#2A2C36", "#F07A93", "#A5D88A", "#F5C451", "#82AAFF", "#C792EA", "#7FD6C2", "#D3D5DC", "#636776", "#FF8FA6", "#B9E79F", "#FFD675", "#9DBDFF", "#D7A8F2", "#97E3D2", "#FFFFFF"}
	ansiLiteLight = [16]string{"#1F2328", "#CF222E", "#1A7F37", "#9A6700", "#0969DA", "#8250DF", "#1B7C83", "#6E7781", "#57606A", "#A40E26", "#2DA44E", "#BF8700", "#218BFF", "#A475F9", "#3192AA", "#8C959F"}
	ansiDracula   = [16]string{"#21222C", "#FF5555", "#50FA7B", "#F1FA8C", "#BD93F9", "#FF79C6", "#8BE9FD", "#F8F8F2", "#6272A4", "#FF6E6E", "#69FF94", "#FFFFA5", "#D6ACFF", "#FF92DF", "#A4FFFF", "#FFFFFF"}
	ansiNord      = [16]string{"#3B4252", "#BF616A", "#A3BE8C", "#EBCB8B", "#81A1C1", "#B48EAD", "#88C0D0", "#E5E9F0", "#4C566A", "#BF616A", "#A3BE8C", "#EBCB8B", "#81A1C1", "#B48EAD", "#8FBCBB", "#ECEFF4"}
	ansiGruvDark  = [16]string{"#282828", "#CC241D", "#98971A", "#D79921", "#458588", "#B16286", "#689D6A", "#A89984", "#928374", "#FB4934", "#B8BB26", "#FABD2F", "#83A598", "#D3869B", "#8EC07C", "#EBDBB2"}
	ansiGruvLight = [16]string{"#FBF1C7", "#CC241D", "#98971A", "#D79921", "#458588", "#B16286", "#689D6A", "#7C6F64", "#928374", "#9D0006", "#79740E", "#B57614", "#076678", "#8F3F71", "#427B58", "#3C3836"}
	ansiTokyo     = [16]string{"#15161E", "#F7768E", "#9ECE6A", "#E0AF68", "#7AA2F7", "#BB9AF7", "#7DCFFF", "#A9B1D6", "#414868", "#F7768E", "#9ECE6A", "#E0AF68", "#7AA2F7", "#BB9AF7", "#7DCFFF", "#C0CAF5"}
	ansiMocha     = [16]string{"#45475A", "#F38BA8", "#A6E3A1", "#F9E2AF", "#89B4FA", "#F5C2E7", "#94E2D5", "#BAC2DE", "#585B70", "#F38BA8", "#A6E3A1", "#F9E2AF", "#89B4FA", "#F5C2E7", "#94E2D5", "#A6ADC8"}
	ansiLatte     = [16]string{"#5C5F77", "#D20F39", "#40A02B", "#DF8E1D", "#1E66F5", "#EA76CB", "#179299", "#ACB0BE", "#6C6F85", "#D20F39", "#40A02B", "#DF8E1D", "#1E66F5", "#EA76CB", "#179299", "#BCC0CC"}
	ansiSolarized = [16]string{"#073642", "#DC322F", "#859900", "#B58900", "#268BD2", "#D33682", "#2AA198", "#EEE8D5", "#002B36", "#CB4B16", "#586E75", "#657B83", "#839496", "#6C71C4", "#93A1A1", "#FDF6E3"}
	ansiOneDark   = [16]string{"#282C34", "#E06C75", "#98C379", "#E5C07B", "#61AFEF", "#C678DD", "#56B6C2", "#ABB2BF", "#5C6370", "#E06C75", "#98C379", "#E5C07B", "#61AFEF", "#C678DD", "#56B6C2", "#FFFFFF"}
	ansiOneLight  = [16]string{"#383A42", "#E45649", "#50A14F", "#C18401", "#4078F2", "#A626A4", "#0184BC", "#A0A1A7", "#696C77", "#E45649", "#50A14F", "#C18401", "#4078F2", "#A626A4", "#0184BC", "#FAFAFA"}
	ansiGitDark   = [16]string{"#484F58", "#FF7B72", "#3FB950", "#D29922", "#58A6FF", "#BC8CFF", "#39C5CF", "#B1BAC4", "#6E7681", "#FFA198", "#56D364", "#E3B341", "#79C0FF", "#D2A8FF", "#56D4DD", "#FFFFFF"}
)

// SyntaxThemes are the ready-made syntax themes. Several follow well-known
// open-source color schemes (MIT licensed; see NOTICE).
var SyntaxThemes = []SyntaxTheme{
	syn("litecode", "LiteCode", true, ansiLiteDark,
		"#D3D5DC", "#C792EA", "#7FD6C2", "#82AAFF", "#A5D88A", "#F2A97A", "#6C7183", "#F2A97A", "#9BA1B4", "#8FC5FF", "#F07A93", "#B1A6FF"),
	syn("litecode-light", "LiteCode Light", false, ansiLiteLight,
		"#24292F", "#7C3AED", "#0E7C6B", "#2457C5", "#2F7D32", "#B45309", "#8A8F9C", "#B45309", "#5B6170", "#1D5FB8", "#C2255C", "#6D28D9"),
	syn("dracula", "Dracula", true, ansiDracula,
		"#F8F8F2", "#FF79C6", "#8BE9FD", "#50FA7B", "#F1FA8C", "#BD93F9", "#6272A4", "#BD93F9", "#F8F8F2", "#66D9EF", "#FF79C6", "#BD93F9"),
	syn("nord", "Nord", true, ansiNord,
		"#D8DEE9", "#81A1C1", "#8FBCBB", "#88C0D0", "#A3BE8C", "#B48EAD", "#616E88", "#B48EAD", "#ECEFF4", "#8FBCBB", "#81A1C1", "#88C0D0"),
	syn("gruvbox-dark", "Gruvbox Dark", true, ansiGruvDark,
		"#EBDBB2", "#FB4934", "#FABD2F", "#B8BB26", "#B8BB26", "#D3869B", "#928374", "#D3869B", "#A89984", "#83A598", "#8EC07C", "#FE8019"),
	syn("gruvbox-light", "Gruvbox Light", false, ansiGruvLight,
		"#3C3836", "#9D0006", "#B57614", "#79740E", "#79740E", "#8F3F71", "#928374", "#8F3F71", "#665C54", "#076678", "#427B58", "#AF3A03"),
	syn("tokyo-night", "Tokyo Night", true, ansiTokyo,
		"#C0CAF5", "#BB9AF7", "#2AC3DE", "#7AA2F7", "#9ECE6A", "#FF9E64", "#565F89", "#FF9E64", "#89DDFF", "#73DACA", "#F7768E", "#7AA2F7"),
	syn("catppuccin-mocha", "Catppuccin Mocha", true, ansiMocha,
		"#CDD6F4", "#CBA6F7", "#F9E2AF", "#89B4FA", "#A6E3A1", "#FAB387", "#6C7086", "#FAB387", "#9399B2", "#89DCEB", "#F38BA8", "#F5C2E7"),
	syn("catppuccin-latte", "Catppuccin Latte", false, ansiLatte,
		"#4C4F69", "#8839EF", "#DF8E1D", "#1E66F5", "#40A02B", "#FE640B", "#9CA0B0", "#FE640B", "#7C7F93", "#04A5E5", "#D20F39", "#EA76CB"),
	syn("solarized-dark", "Solarized Dark", true, ansiSolarized,
		"#93A1A1", "#859900", "#B58900", "#268BD2", "#2AA198", "#D33682", "#586E75", "#CB4B16", "#839496", "#268BD2", "#CB4B16", "#6C71C4"),
	syn("solarized-light", "Solarized Light", false, ansiSolarized,
		"#586E75", "#859900", "#B58900", "#268BD2", "#2AA198", "#D33682", "#93A1A1", "#CB4B16", "#657B83", "#268BD2", "#CB4B16", "#6C71C4"),
	syn("one-dark", "One Dark", true, ansiOneDark,
		"#ABB2BF", "#C678DD", "#E5C07B", "#61AFEF", "#98C379", "#D19A66", "#5C6370", "#D19A66", "#ABB2BF", "#E06C75", "#E06C75", "#61AFEF"),
	syn("one-light", "One Light", false, ansiOneLight,
		"#383A42", "#A626A4", "#C18401", "#4078F2", "#50A14F", "#986801", "#A0A1A7", "#986801", "#383A42", "#E45649", "#E45649", "#4078F2"),
	syn("github-dark", "GitHub Dark", true, ansiGitDark,
		"#E6EDF3", "#FF7B72", "#FFA657", "#D2A8FF", "#A5D6FF", "#79C0FF", "#8B949E", "#79C0FF", "#E6EDF3", "#79C0FF", "#7EE787", "#79C0FF"),
	syn("github-light", "GitHub Light", false, ansiLiteLight,
		"#1F2328", "#CF222E", "#953800", "#8250DF", "#0A3069", "#0550AE", "#6E7781", "#0550AE", "#1F2328", "#0550AE", "#116329", "#0550AE"),
	syn("ember", "Ember", true, ansiGruvDark,
		"#E8DED2", "#FF8A5B", "#F6C177", "#7FB2E5", "#B5D99C", "#F4A6C1", "#7E7469", "#F4A6C1", "#B3A698", "#EBBCBA", "#FF8A5B", "#F6C177"),
	syn("quiet-dark", "Quiet Dark", true, ansiLiteDark,
		"#D4D4D8", "#E4E4E7", "#A1A1AA", "#E4E4E7", "#A3B8A0", "#C8B79A", "#6B6B73", "#C8B79A", "#8E8E96", "#BABAC2", "#E4E4E7", "#FFFFFF"),
	syn("quiet-light", "Quiet Light", false, ansiLiteLight,
		"#27272A", "#18181B", "#52525B", "#18181B", "#4D6B4A", "#7A5D2E", "#A1A1AA", "#7A5D2E", "#71717A", "#3F3F46", "#18181B", "#000000"),
}

// SyntaxByID finds a syntax theme, or nil.
func SyntaxByID(id string) *SyntaxTheme {
	for i := range SyntaxThemes {
		if SyntaxThemes[i].ID == id {
			return &SyntaxThemes[i]
		}
	}
	return nil
}
