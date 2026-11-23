package widgets

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"
)

// BrandMark is LiteCode's logo: a feather made of code lines on a violet
// tile, as in assets/icon.svg. Giode draws icons in one color, so the mark
// is white here, without the tinted tokens of the app icon.
func BrandMark(size int) giode.Element {
	mark := size * 70 / 100
	return giode.Box(giode.Styles{
		Width: properties.Width(size), Height: properties.Height(size), BorderRadius: properties.BorderRadius(size * 24 / 100),
		BackgroundGradient: giode.Linear(135, "#A493FF", "#6E58F2", "#3F2DB8"), Align: giode.AlignCenter, Justify: giode.JustifyCenter,
	}, giode.Icon("brand", giode.Styles{Width: properties.Width(mark), Height: properties.Height(mark), Color: "#FFFFFF"}))
}
