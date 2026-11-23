package widgets

import (
	"hash/fnv"
	"strings"
	"unicode"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"
)

// badgeColors are the gradient pairs project badges are painted with.
var badgeColors = [][2]string{
	{"#DB3D3C", "#FF8E42"}, {"#F57236", "#FCBA3F"}, {"#E8A33B", "#F5D44B"}, {"#25B87A", "#94D44B"},
	{"#26AAB8", "#3BCCB5"}, {"#3B6CF5", "#5EC3F5"}, {"#7847F5", "#C24BEB"}, {"#E04BAA", "#F57CA3"},
}

// ProjectBadge is the rounded square with a project's initials, colored by
// its name, as the welcome screen and project widget show.
func ProjectBadge(name string, size int) giode.Element {
	h := fnv.New32a()
	h.Write([]byte(name))
	pair := badgeColors[h.Sum32()%uint32(len(badgeColors))]
	return giode.Box(giode.Styles{
		Width: properties.Width(size), Height: properties.Height(size), BorderRadius: properties.BorderRadius(size / 4),
		BackgroundGradient: giode.Linear(135, pair[0], pair[1]), Align: giode.AlignCenter, Justify: giode.JustifyCenter,
	}, giode.Text(initials(name), giode.Styles{Color: "#FFFFFF", FontSize: properties.FontSize(float32(size) * 0.42), FontWeight: giode.FontWeightSemiBold}))
}

// initials takes the first letters of the first two words of a name.
func initials(name string) string {
	words := strings.FieldsFunc(name, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
	var out []rune
	for _, w := range words {
		out = append(out, unicode.ToUpper([]rune(w)[0]))
		if len(out) == 2 {
			break
		}
	}
	if len(out) == 0 {
		return "?"
	}
	return string(out)
}
