package config

// Settings are the user's global preferences. They apply to every project:
// formatting in particular is configured once, not per project.
type Settings struct {
	UIFontSize       int    `json:"uiFontSize"`
	EditorFontSize   int    `json:"editorFontSize"`
	EditorLineHeight int    `json:"editorLineHeight"`
	Theme            string `json:"theme"`
	FormatOnSave     bool   `json:"formatOnSave"`
	UsePrettier      bool   `json:"usePrettier"`
	TabSize          int    `json:"tabSize"`
	ShowWelcome      bool   `json:"showWelcome"`
}

// DefaultSettings are the settings of a fresh install.
func DefaultSettings() Settings {
	return Settings{
		UIFontSize:       13,
		EditorFontSize:   13,
		EditorLineHeight: 20,
		Theme:            "dark",
		FormatOnSave:     true,
		UsePrettier:      true,
		TabSize:          4,
		ShowWelcome:      true,
	}
}

// LoadSettings reads settings.json, filling unset values with the defaults.
func LoadSettings() (Settings, error) {
	s := DefaultSettings()
	path, err := Path("settings.json")
	if err != nil {
		return s, err
	}
	if err := LoadJSON(path, &s); err != nil {
		return DefaultSettings(), err
	}
	return s.withDefaults(), nil
}

// SaveSettings writes settings.json.
func SaveSettings(s Settings) error {
	path, err := Path("settings.json")
	if err != nil {
		return err
	}
	return SaveJSON(path, s)
}

func (s Settings) withDefaults() Settings {
	d := DefaultSettings()
	if s.UIFontSize <= 0 {
		s.UIFontSize = d.UIFontSize
	}
	if s.EditorFontSize <= 0 {
		s.EditorFontSize = d.EditorFontSize
	}
	if s.EditorLineHeight <= 0 {
		s.EditorLineHeight = d.EditorLineHeight
	}
	if s.TabSize <= 0 {
		s.TabSize = d.TabSize
	}
	if s.Theme == "" {
		s.Theme = d.Theme
	}
	return s
}
