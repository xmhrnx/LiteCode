package widgets

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"
)

// BrandMark is LiteCode's logo: a violet tile with an "L".
func BrandMark(size int) giode.Element {
	return giode.Box(giode.Styles{
		Width: properties.Width(size), Height: properties.Height(size), BorderRadius: properties.BorderRadius(size * 3 / 10),
		BackgroundGradient: giode.Linear(135, "#9A8CFF", "#5B4BE0"), Align: giode.AlignCenter, Justify: giode.JustifyCenter,
	}, giode.Text("L", giode.Styles{Color: "#FFFFFF", FontSize: properties.FontSize(float32(size) * 0.5), FontWeight: giode.FontWeightSemiBold}))
}
