package theme

// UITheme is a ready-made color scheme for the whole window, with the
// syntax theme it comes with.
type UITheme struct {
	ID, Name string
	Dark     bool
	Syntax   string // the syntax theme it pairs with
	Scheme   Scheme
	b        base // what Scheme was built from, to build it again with changes
}

// made is a UI theme built from base colors.
func made(id, name, syntax string, b base) UITheme {
	return UITheme{ID: id, Name: name, Dark: b.dark, Syntax: syntax, Scheme: build(b), b: b}
}

// baseFields map the Scheme colors a theme is built from to their base
// color: changing one of them by hand rebuilds the colors derived from it.
var baseFields = map[string]func(b *base) *string{
	"Accent": func(b *base) *string { return &b.accent }, "Editor": func(b *base) *string { return &b.bg },
	"Panel": func(b *base) *string { return &b.panel }, "Raised": func(b *base) *string { return &b.raised },
	"Text": func(b *base) *string { return &b.text }, "TextMuted": func(b *base) *string { return &b.muted },
	"Success": func(b *base) *string { return &b.success }, "Warning": func(b *base) *string { return &b.warning },
	"Danger": func(b *base) *string { return &b.danger }, "StatusModified": func(b *base) *string { return &b.info },
}

// base are the few colors a built theme starts from.
type base struct {
	dark                           bool
	bg, panel, raised              string // editor, panels, popups
	text, muted, accent            string
	success, warning, danger, info string
}

// build derives a whole scheme from base colors.
func build(b base) Scheme {
	white, black := "#FFFFFF", "#000000"
	s := Scheme{
		Card: b.panel, Editor: b.bg, Panel: b.panel, Raised: b.raised,
		Text: b.text, TextMuted: b.muted, Accent: b.accent,
		Danger: b.danger, Warning: b.warning, Success: b.success,
		StatusAdded: b.success, StatusDeleted: mix(b.muted, b.bg, 0.8), StatusUntracked: mix(b.danger, b.warning, 0.5),
		StatusConflict: b.danger, StatusModified: b.info,
		RefHead: b.warning, RefLocal: b.success, RefRemote: b.info, RefTag: b.muted,
		AccentText: white,
	}
	s.CardLine = mix(b.text, b.panel, 0.08)
	s.Separator = mix(b.text, b.panel, 0.07)
	s.PopupLine = mix(b.text, b.raised, 0.13)
	s.MenuBg = mix(b.text, b.raised, 0.03)
	s.Hover = mix(b.text, b.panel, 0.05)
	s.Pressed = mix(b.text, b.panel, 0.09)
	s.Selection = mix(b.accent, b.panel, 0.26)
	s.InputLine = mix(b.text, b.bg, 0.13)
	s.TextCode = mix(b.text, b.bg, 0.92)
	s.TextFaint = mix(b.text, b.bg, 0.42)
	s.AccentSoft = mix(b.accent, b.panel, 0.18)
	s.Match = mix(b.accent, b.bg, 0.35)
	s.DiffInserted = mix(b.success, b.bg, 0.16)
	s.DiffDeleted = mix(b.danger, b.bg, 0.16)
	s.DiffChanged = mix(b.info, b.bg, 0.16)
	s.Glow = alpha(b.accent, 0x66)
	s.SoftGlow = alpha(b.accent, 0x33)
	s.AccentLine = mix(b.accent, b.raised, 0.45)
	s.Rail = mix(b.accent, b.panel, 0.35)
	s.ScrollThumb = mix(b.text, b.panel, 0.2)
	s.WashFrom, s.WashMid, s.WashTo = mix(b.accent, b.panel, 0.38), mix(b.accent, b.panel, 0.22), mix(b.accent, b.panel, 0.1)
	s.BubbleFrom, s.BubbleTo = mix(b.accent, b.raised, 0.45), mix(b.accent, b.raised, 0.3)
	s.TileFrom, s.TileTo = mix(b.accent, b.raised, 0.08), b.raised
	s.EditorTop = mix(b.accent, b.bg, 0.07)
	s.HeroLine = mix(b.accent, b.panel, 0.4)
	if b.dark {
		s.Backdrop = []string{mix(b.accent, b.bg, 0.25), mix(black, b.bg, 0.1), mix(black, b.bg, 0.3), mix(b.info, b.bg, 0.06)}
		s.InputBg = mix(black, b.bg, 0.15)
		s.FocusLine = mix(white, b.accent, 0.12)
		s.AccentFrom, s.AccentTo = mix(white, b.accent, 0.22), mix(black, b.accent, 0.08)
		s.TextAccent = mix(white, b.accent, 0.4)
		s.Shadow, s.PopupTop = "#000000B0", mix(b.text, b.raised, 0.03)
		s.HeroFrom, s.HeroMid, s.HeroTo = mix(b.accent, b.bg, 0.38), mix(b.accent, b.bg, 0.14), b.bg
		s.HeroText, s.HeroFaint = mix(b.accent, b.text, 0.2), mix(b.accent, b.muted, 0.3)
	} else {
		s.Backdrop = []string{mix(b.accent, b.bg, 0.16), mix(black, b.bg, 0.04), mix(black, b.bg, 0.07), mix(b.info, b.bg, 0.08)}
		s.InputBg = mix(white, b.bg, 0.6)
		s.FocusLine = b.accent
		s.AccentFrom, s.AccentTo = mix(white, b.accent, 0.12), mix(black, b.accent, 0.1)
		s.TextAccent = mix(black, b.accent, 0.15)
		s.Shadow, s.PopupTop = "#1A1A3030", mix(white, b.raised, 0.5)
		s.HeroFrom, s.HeroMid, s.HeroTo = mix(b.accent, b.bg, 0.22), mix(b.accent, b.bg, 0.08), b.bg
		s.HeroText, s.HeroFaint = mix(b.accent, b.text, 0.25), mix(b.accent, b.muted, 0.3)
	}
	s.TabActive = s.FocusLine
	return s
}

// UIThemes are the ready-made UI themes, dark ones first.
var UIThemes = []UITheme{
	{ID: "litecode-dark", Name: "LiteCode Dark", Dark: true, Syntax: "litecode", Scheme: darkScheme, b: base{dark: true,
		bg: "#131419", panel: "#16171D", raised: "#1D1E26", text: "#E6E7EC", muted: "#9A9EAA", accent: "#7C6CF6",
		success: "#5CCF8F", warning: "#F5C451", danger: "#FF6B6B", info: "#74B3FF"}},
	made("midnight", "Midnight", "tokyo-night", base{dark: true,
		bg: "#0F1420", panel: "#131A28", raised: "#1A2233", text: "#E3E8F2", muted: "#8E9AB2", accent: "#4C8DFF",
		success: "#4FCB8D", warning: "#F2C14E", danger: "#FF6B7A", info: "#5CC8FF"}),
	made("emerald", "Emerald", "litecode", base{dark: true,
		bg: "#0F1714", panel: "#131D19", raised: "#1A2621", text: "#E2EEE8", muted: "#8DA69A", accent: "#2FBF85",
		success: "#5BD69A", warning: "#E8C15A", danger: "#FF7272", info: "#5BB8E8"}),
	made("ember", "Ember", "ember", base{dark: true,
		bg: "#161210", panel: "#1B1613", raised: "#241D19", text: "#EFE6DE", muted: "#A89A8E", accent: "#F2763C",
		success: "#8FCB6A", warning: "#F6C177", danger: "#FF6B6B", info: "#7FB2E5"}),
	made("rose", "Rosé", "catppuccin-mocha", base{dark: true,
		bg: "#17121A", panel: "#1C1620", raised: "#251D2A", text: "#EEE4EE", muted: "#A898AA", accent: "#E5669B",
		success: "#7FD1A8", warning: "#F2C27B", danger: "#FF6F7D", info: "#8AB8F0"}),
	made("graphite", "Graphite", "quiet-dark", base{dark: true,
		bg: "#121212", panel: "#171717", raised: "#1F1F1F", text: "#E8E8E8", muted: "#9A9A9A", accent: "#6E8BFF",
		success: "#5CC98A", warning: "#E8B84A", danger: "#FF6464", info: "#62B6F0"}),
	made("nord", "Nord", "nord", base{dark: true,
		bg: "#2E3440", panel: "#2B303B", raised: "#3B4252", text: "#ECEFF4", muted: "#A3ACBD", accent: "#88C0D0",
		success: "#A3BE8C", warning: "#EBCB8B", danger: "#BF616A", info: "#81A1C1"}),
	made("dracula", "Dracula", "dracula", base{dark: true,
		bg: "#282A36", panel: "#21222C", raised: "#343746", text: "#F8F8F2", muted: "#A4A9C4", accent: "#BD93F9",
		success: "#50FA7B", warning: "#F1FA8C", danger: "#FF5555", info: "#8BE9FD"}),
	made("gruvbox-dark", "Gruvbox Dark", "gruvbox-dark", base{dark: true,
		bg: "#282828", panel: "#1D2021", raised: "#32302F", text: "#EBDBB2", muted: "#A89984", accent: "#D79921",
		success: "#B8BB26", warning: "#FABD2F", danger: "#FB4934", info: "#83A598"}),
	made("tokyo-night", "Tokyo Night", "tokyo-night", base{dark: true,
		bg: "#1A1B26", panel: "#16161E", raised: "#24283B", text: "#C0CAF5", muted: "#7982A9", accent: "#7AA2F7",
		success: "#9ECE6A", warning: "#E0AF68", danger: "#F7768E", info: "#7DCFFF"}),
	made("catppuccin-mocha", "Catppuccin Mocha", "catppuccin-mocha", base{dark: true,
		bg: "#1E1E2E", panel: "#181825", raised: "#313244", text: "#CDD6F4", muted: "#A6ADC8", accent: "#CBA6F7",
		success: "#A6E3A1", warning: "#F9E2AF", danger: "#F38BA8", info: "#89B4FA"}),
	made("solarized-dark", "Solarized Dark", "solarized-dark", base{dark: true,
		bg: "#002B36", panel: "#00252F", raised: "#073642", text: "#EEE8D5", muted: "#93A1A1", accent: "#268BD2",
		success: "#859900", warning: "#B58900", danger: "#DC322F", info: "#2AA198"}),
	made("one-dark", "One Dark", "one-dark", base{dark: true,
		bg: "#282C34", panel: "#21252B", raised: "#2F343F", text: "#DCDFE4", muted: "#9DA5B4", accent: "#61AFEF",
		success: "#98C379", warning: "#E5C07B", danger: "#E06C75", info: "#56B6C2"}),
	made("github-dark", "GitHub Dark", "github-dark", base{dark: true,
		bg: "#0D1117", panel: "#010409", raised: "#161B22", text: "#E6EDF3", muted: "#8D96A0", accent: "#2F81F7",
		success: "#3FB950", warning: "#D29922", danger: "#F85149", info: "#58A6FF"}),
	made("litecode-light", "LiteCode Light", "litecode-light", base{
		bg: "#FFFFFF", panel: "#F6F6FA", raised: "#FFFFFF", text: "#1E2030", muted: "#5E6275", accent: "#6A55F0",
		success: "#1F9D55", warning: "#B7791F", danger: "#D92D3B", info: "#2563EB"}),
	made("paper", "Paper", "quiet-light", base{
		bg: "#FCFBF8", panel: "#F3F1EC", raised: "#FFFFFF", text: "#2A2925", muted: "#6E6B63", accent: "#2E6BD8",
		success: "#2F8A4F", warning: "#A86A08", danger: "#C8352E", info: "#2E6BD8"}),
	made("github-light", "GitHub Light", "github-light", base{
		bg: "#FFFFFF", panel: "#F6F8FA", raised: "#FFFFFF", text: "#1F2328", muted: "#59636E", accent: "#0969DA",
		success: "#1A7F37", warning: "#9A6700", danger: "#CF222E", info: "#0969DA"}),
	made("solarized-light", "Solarized Light", "solarized-light", base{
		bg: "#FDF6E3", panel: "#F5EFDC", raised: "#FFFBEF", text: "#073642", muted: "#657B83", accent: "#268BD2",
		success: "#859900", warning: "#B58900", danger: "#DC322F", info: "#2AA198"}),
	made("catppuccin-latte", "Catppuccin Latte", "catppuccin-latte", base{
		bg: "#EFF1F5", panel: "#E6E9EF", raised: "#F7F8FA", text: "#4C4F69", muted: "#6C6F85", accent: "#8839EF",
		success: "#40A02B", warning: "#DF8E1D", danger: "#D20F39", info: "#1E66F5"}),
	made("gruvbox-light", "Gruvbox Light", "gruvbox-light", base{
		bg: "#FBF1C7", panel: "#F2E5BC", raised: "#FDF6DA", text: "#3C3836", muted: "#7C6F64", accent: "#AF3A03",
		success: "#79740E", warning: "#B57614", danger: "#9D0006", info: "#076678"}),
	made("one-light", "One Light", "one-light", base{
		bg: "#FAFAFA", panel: "#F0F0F1", raised: "#FFFFFF", text: "#383A42", muted: "#696C77", accent: "#4078F2",
		success: "#50A14F", warning: "#C18401", danger: "#E45649", info: "#0184BC"}),
}

// UIThemeByID finds a UI theme, falling back to LiteCode Dark.
func UIThemeByID(id string) *UITheme {
	if id == "dark" {
		id = "litecode-dark" // the setting's value before themes
	}
	for i := range UIThemes {
		if UIThemes[i].ID == id {
			return &UIThemes[i]
		}
	}
	return &UIThemes[0]
}
