package theme

import (
	"github.com/xmhrnx/giode"

	"litecode/assets"
)

// Font families.
const (
	UIFont   = "Inter"
	CodeFont = "JetBrains Mono"
)

// Font sizes.
const (
	UISize      = 13
	SmallSize   = 12
	CodeSize    = 13
	CodeLineGap = 20
)

// LoadFonts registers Inter for the UI and JetBrains Mono for code, and makes
// Inter the default for every component.
func LoadFonts() error {
	files := []struct {
		name, family string
		weight       giode.FontWeight
	}{
		{"Inter-Regular.ttf", UIFont, ""},
		{"Inter-Medium.ttf", UIFont, giode.FontWeightMedium},
		{"Inter-SemiBold.ttf", UIFont, giode.FontWeightSemiBold},
		{"JetBrainsMono-Regular.ttf", CodeFont, ""},
		{"JetBrainsMono-Bold.ttf", CodeFont, giode.FontWeightBold},
	}
	var ff []giode.FontFile
	for _, f := range files {
		data, err := assets.Fonts.ReadFile("fonts/" + f.name)
		if err != nil {
			return err
		}
		ff = append(ff, giode.FontFile{Family: f.family, Weight: f.weight, Data: data})
	}
	if err := giode.RegisterFont(ff...); err != nil {
		return err
	}
	giode.SetDefaultFont(UIFont)
	return nil
}
