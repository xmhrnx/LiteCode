// Package theme holds LiteCode's look: its color scheme, the Inter and JetBrains
// Mono fonts, the icons, and the styles several components share.
package theme

import "github.com/xmhrnx/giode"

// Scheme is every color the IDE paints with. Giode's palette has a fixed set
// of tokens, so the IDE-specific ones live here; Current returns the scheme of
// the active theme.
type Scheme struct {
	// Backdrop behind the floating cards: a gradient, top-left to bottom-right
	Backdrop []string
	// Cards: the activity bar, panels and editor float on the backdrop
	Card     string
	CardLine string // 1dp edge that catches the light
	Glow     string // accent shadow under highlighted controls

	// Surfaces, darkest to lightest
	Editor     string // editor, diff and lists of content
	Panel      string // panels, title bar, status bar
	Raised     string // popups, menus, dialogs
	Separator  string // hairlines between areas
	PopupLine  string // borders of raised surfaces and inputs
	MenuBg     string // context and dropdown menus
	Hover      string // row and button hover
	Pressed    string // button press, toggled buttons
	Selection  string // selected rows
	InputBg    string
	InputLine  string
	FocusLine  string
	TabActive  string // marker of the selected tab
	Accent     string // primary buttons, active panel icons, links
	AccentFrom string // accent gradient stops
	AccentTo   string
	AccentSoft string // accent-tinted backgrounds
	AccentText string // text on Accent

	// Text
	Text       string
	TextCode   string
	TextMuted  string
	TextFaint  string
	TextAccent string

	// VCS file status
	StatusModified  string
	StatusAdded     string
	StatusDeleted   string
	StatusUntracked string
	StatusConflict  string

	// Diff
	DiffInserted string
	DiffDeleted  string
	DiffChanged  string

	// Log reference labels
	RefHead   string
	RefLocal  string
	RefRemote string
	RefTag    string

	// Search hits
	Match string

	Danger  string
	Warning string
	Success string
}

// darkScheme is LiteCode Dark: cool neutrals with a violet accent.
var darkScheme = Scheme{
	Backdrop: []string{"#261E4E", "#13121C", "#0C0D11", "#0F1520"},
	Card:     "#16171D",
	CardLine: "#25272F",
	Glow:     "#7C6CF666",

	Editor:     "#131419",
	Panel:      "#16171D",
	Raised:     "#1D1E26",
	Separator:  "#22242C",
	PopupLine:  "#2E303A",
	MenuBg:     "#22232C",
	Hover:      "#1F2028",
	Pressed:    "#272935",
	Selection:  "#2A2652",
	InputBg:    "#111217",
	InputLine:  "#2C2E38",
	FocusLine:  "#8B7CFF",
	TabActive:  "#8B7CFF",
	Accent:     "#7C6CF6",
	AccentFrom: "#9D8CFF",
	AccentTo:   "#6650F0",
	AccentSoft: "#26224A",
	AccentText: "#FFFFFF",

	Text:       "#E6E7EC",
	TextCode:   "#D3D5DC",
	TextMuted:  "#9A9EAA",
	TextFaint:  "#636776",
	TextAccent: "#B1A6FF",

	StatusModified:  "#74B3FF",
	StatusAdded:     "#5CCF8F",
	StatusDeleted:   "#7D818C",
	StatusUntracked: "#F28B74",
	StatusConflict:  "#FF6076",

	DiffInserted: "#163627",
	DiffDeleted:  "#3C2028",
	DiffChanged:  "#1F2A46",

	RefHead:   "#F5C451",
	RefLocal:  "#5CCF8F",
	RefRemote: "#58C4E8",
	RefTag:    "#9A9EAA",

	Match: "#3B3470",

	Danger:  "#FF6B6B",
	Warning: "#F5C451",
	Success: "#5CCF8F",
}

var current = &darkScheme

// Current returns the active color scheme.
func Current() *Scheme { return current }

// Apply makes the scheme Giode's palette too, so built-in components
// (menus, dialogs, tabs, inputs, toasts) match the IDE.
func Apply() {
	c := current
	giode.SetMode(giode.ModeDark)
	giode.SetPalette(giode.ModeDark, giode.Palette{
		Primary:    c.Accent,
		Secondary:  c.TextMuted,
		Background: c.Editor,
		Surface:    c.Raised,
		Text:       c.Text,
		Border:     c.PopupLine,
		OnPrimary:  c.AccentText,
		Muted:      c.TextMuted,
		Accent:     c.TextAccent,
		Success:    c.Success,
		Warning:    c.Warning,
		Danger:     c.Danger,
		Info:       c.Accent,
	})
}
