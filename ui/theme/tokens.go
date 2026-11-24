package theme

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/xmhrnx/giode"

	"litecode/internal/highlight"
)

// Choice is what the user picked: a UI theme, a syntax theme ("" goes
// with the UI theme), and colors they changed by hand, by Scheme field
// name ("Accent") or token name ("Keyword").
type Choice struct {
	UI, Syntax string
	UIColors   map[string]string
	Code       map[string]string
}

// backdropStops is how many backdrop gradient stops get tokens.
const backdropStops = 4

var (
	uiTokens   Scheme     // every color as its theme.* token
	codeTokens CodeColors // the same for the editor
	uiValues   Scheme     // what the tokens stand for now
	codeValues CodeColors
	active     Choice
)

func init() {
	v := reflect.ValueOf(&uiTokens).Elem()
	for i := range v.NumField() {
		f := v.Type().Field(i)
		if f.Type.Kind() == reflect.String {
			v.Field(i).SetString("theme.lc" + f.Name)
		}
	}
	for i := range backdropStops {
		uiTokens.Backdrop = append(uiTokens.Backdrop, "theme.lcBackdrop"+strconv.Itoa(i))
	}
	for k := range codeTokens.Tokens {
		codeTokens.Tokens[k] = "theme.lcTok" + TokenNames[k]
	}
	cv := reflect.ValueOf(&codeTokens).Elem()
	for i := range cv.NumField() {
		f := cv.Type().Field(i)
		if f.Type.Kind() == reflect.String {
			cv.Field(i).SetString("theme.lcCode" + f.Name)
		}
	}
	for i := range codeTokens.ANSI {
		codeTokens.ANSI[i] = "theme.lcAnsi" + strconv.Itoa(i)
	}
	Apply(Choice{})
}

// CodeScheme returns the editor's colors as theme tokens.
func CodeScheme() *CodeColors { return &codeTokens }

// Values is the active scheme's colors as "#RRGGBB" values, for showing
// and editing them.
func Values() *Scheme { return &uiValues }

// CodeValues is the active editor colors as values.
func CodeValues() *CodeColors { return &codeValues }

// Active is the choice in effect.
func Active() Choice { return active }

// IsDark reports whether the active UI theme is dark.
func IsDark() bool { return UIThemeByID(active.UI).Dark }

// SchemeFields are the Scheme's color fields, in order, for the editor
// of custom colors.
func SchemeFields() []string {
	var names []string
	t := reflect.TypeOf(Scheme{})
	for i := range t.NumField() {
		if t.Field(i).Type.Kind() == reflect.String {
			names = append(names, t.Field(i).Name)
		}
	}
	return names
}

// SchemeValue is the value of a Scheme color field by name.
func SchemeValue(s *Scheme, field string) string {
	f := reflect.ValueOf(s).Elem().FieldByName(field)
	if !f.IsValid() || f.Kind() != reflect.String {
		return ""
	}
	return f.String()
}

// Apply makes a choice the active colors: it sets every token, and
// Giode's own palette so its components (menus, dialogs, inputs, toasts)
// match, in one palette change. Styles hold tokens, so the change shows on the next frame.
func Apply(c Choice) {
	active = c
	ui := UIThemeByID(c.UI)
	s := ui.Scheme
	// Base colors changed by hand: build the theme again from them, so
	// what's derived from them (gradients, selection, edges) follows.
	b, rebuilt := ui.b, false
	for name, val := range c.UIColors {
		if f, ok := baseFields[name]; ok && ValidColor(val) {
			*f(&b) = val
			rebuilt = true
		}
	}
	if rebuilt {
		s = build(b)
	}
	s.Backdrop = append([]string(nil), s.Backdrop...)
	sv := reflect.ValueOf(&s).Elem()
	for name, val := range c.UIColors {
		if f := sv.FieldByName(name); f.IsValid() && f.Kind() == reflect.String && ValidColor(val) {
			f.SetString(val)
		}
	}
	syntax := SyntaxByID(c.Syntax)
	if syntax == nil {
		syntax = SyntaxByID(ui.Syntax)
	}
	code := deriveCode(&s, syntax)
	for k, name := range TokenNames {
		if val, ok := c.Code[name]; ok && ValidColor(val) {
			code.Tokens[k] = val
		}
	}
	cv := reflect.ValueOf(&code).Elem()
	for name, val := range c.Code {
		if f := cv.FieldByName(name); f.IsValid() && f.Kind() == reflect.String && ValidColor(val) {
			f.SetString(val)
		}
	}
	uiValues, codeValues = s, code

	custom := map[string]giode.Color{}
	set := func(name, val string) { custom[strings.ToLower(name)] = giode.Color(val) }
	for i := range sv.NumField() {
		f := sv.Type().Field(i)
		if f.Type.Kind() == reflect.String {
			set("lc"+f.Name, sv.Field(i).String())
		}
	}
	for i := range backdropStops {
		set("lcBackdrop"+strconv.Itoa(i), s.Backdrop[min(i, len(s.Backdrop)-1)])
	}
	for k, name := range TokenNames {
		set("lcTok"+name, code.Tokens[k])
	}
	for i := range cv.NumField() {
		f := cv.Type().Field(i)
		if f.Type.Kind() == reflect.String {
			set("lcCode"+f.Name, cv.Field(i).String())
		}
	}
	for i, a := range code.ANSI {
		set("lcAnsi"+strconv.Itoa(i), a)
	}

	p := giode.Palette{
		Primary: s.Accent, Secondary: s.TextMuted, Background: s.Editor, Surface: s.Raised, Text: s.Text,
		Border: s.PopupLine, OnPrimary: s.AccentText, Muted: s.TextMuted, Accent: s.TextAccent,
		Success: s.Success, Warning: s.Warning, Danger: s.Danger, Info: s.Accent,
		Custom: custom,
	}
	giode.SetPalette(giode.ModeDark, p)
	giode.SetPalette(giode.ModeLight, p)
	if ui.Dark {
		giode.SetMode(giode.ModeDark)
	} else {
		giode.SetMode(giode.ModeLight)
	}
}

// deriveCode is the editor colors for a scheme and a syntax theme: the
// tokens come from the syntax theme, the rest follows the editor's
// background and the accent.
func deriveCode(s *Scheme, t *SyntaxTheme) CodeColors {
	c := CodeColors{Tokens: t.Tokens, ANSI: t.ANSI}
	if c.Tokens[highlight.Plain] == "" {
		c.Tokens[highlight.Plain] = s.TextCode
	}
	c.Caret = s.FocusLine
	c.Composition = s.FocusLine
	c.Selection = mix(s.Accent, s.Editor, 0.3)
	c.CurrentLine = mix(s.Text, s.Editor, 0.035)
	c.LineNumber = mix(s.Text, s.Editor, 0.3)
	c.LineActive = mix(s.Text, s.Editor, 0.78)
	c.Guide = mix(s.Text, s.Editor, 0.08)
	c.Popup = mix(s.Accent, s.Raised, 0.28)
	return c
}
