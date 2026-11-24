package keymap

import (
	"maps"
	"sort"
	"strings"

	"litecode/internal/config"
)

// DoubleShift is the stroke for pressing Shift twice quickly.
const DoubleShift = "DoubleShift"

// Keymap holds the strokes of every action, in Giode shortcut syntax
// ("Mod+Shift+N", "Alt+F12").
type Keymap struct {
	Bindings map[Action][]string `json:"bindings"`
	// Rev goes up on every change, so shortcut tables know to rebuild.
	Rev int `json:"-"`
}

// Default is LiteCode's default keymap.
func Default() *Keymap {
	return &Keymap{Bindings: map[Action][]string{
		SearchEverywhere:   {DoubleShift},
		GotoFile:           {"Mod+Shift+N"},
		GotoAction:         {"Mod+Shift+A"},
		FindInPath:         {"Mod+Shift+F"},
		RecentFiles:        {"Mod+E"},
		GotoLine:           {"Mod+G"},
		Switcher:           {"Ctrl+Tab"},
		SwitcherBackward:   {"Ctrl+Shift+Tab"},
		NextTab:            {"Alt+Right"},
		PreviousTab:        {"Alt+Left"},
		CloseEditor:        {"Mod+F4"},
		SaveAll:            {"Mod+S"},
		ReformatCode:       {"Mod+Alt+L"},
		DuplicateLine:      {"Mod+D"},
		DeleteLine:         {"Mod+Y"},
		CommentLine:        {"Mod+/"},
		MoveLineUp:         {"Alt+Shift+Up"},
		MoveLineDown:       {"Alt+Shift+Down"},
		CodeCompletion:     {"Ctrl+Space"},
		ShowSettings:       {"Mod+Alt+S"},
		ProjectWindow:      {"Alt+1"},
		GitWindow:          {"Alt+9"},
		CommitWindow:       {"Alt+0"},
		TerminalWindow:     {"Alt+F12"},
		AssistantWindow:    {"Alt+I"},
		HideToolWindow:     {"Shift+Esc"},
		HideAllWindows:     {"Mod+Shift+F12"},
		CommitAction:       {"Mod+K"},
		PushAction:         {"Mod+Shift+K"},
		UpdateProject:      {"Mod+T"},
		Branches:           {"Mod+Shift+`"},
		Rollback:           {"Mod+Alt+Z"},
		NextDifference:     {"F7"},
		PreviousDifference: {"Shift+F7"},
	}}
}

// Load returns the default keymap with the user's keymap.json applied on top.
// An action bound to an empty list in the file has no stroke.
func Load() (*Keymap, error) {
	km := Default()
	path, err := config.Path("keymap.json")
	if err != nil {
		return km, err
	}
	var user Keymap
	if err := config.LoadJSON(path, &user); err != nil {
		return km, err
	}
	maps.Copy(km.Bindings, user.Bindings)
	return km, nil
}

// Save writes only the bindings that differ from the default.
func (k *Keymap) Save() error {
	path, err := config.Path("keymap.json")
	if err != nil {
		return err
	}
	def := Default()
	diff := &Keymap{Bindings: map[Action][]string{}}
	for a, strokes := range k.Bindings {
		if !equal(strokes, def.Bindings[a]) {
			diff.Bindings[a] = strokes
		}
	}
	return config.SaveJSON(path, diff)
}

// First returns an action's first stroke, for menus and tooltips.
func (k *Keymap) First(a Action) string {
	if s := k.Bindings[a]; len(s) > 0 {
		return s[0]
	}
	return ""
}

// Set replaces an action's strokes.
func (k *Keymap) Set(a Action, strokes ...string) {
	k.Bindings[a] = strokes
	k.Rev++
}

// Reset restores the default bindings.
func (k *Keymap) Reset() {
	k.Bindings = Default().Bindings
	k.Rev++
}

// Conflicts returns the other actions bound to stroke.
func (k *Keymap) Conflicts(a Action, stroke string) []Action {
	var out []Action
	for other, strokes := range k.Bindings {
		if other == a {
			continue
		}
		for _, s := range strokes {
			if s == stroke {
				out = append(out, other)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// Actions lists every action with a name, sorted by name.
func Actions() []Action {
	out := make([]Action, 0, len(Names))
	for a := range Names {
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return Names[out[i]] < Names[out[j]] })
	return out
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Bound reports whether a stroke runs any action, so an editor leaves it
// to the keymap instead of acting on it.
func (k *Keymap) Bound(stroke string) bool {
	for _, strokes := range k.Bindings {
		for _, s := range strokes {
			if strings.EqualFold(s, stroke) {
				return true
			}
		}
	}
	return false
}
