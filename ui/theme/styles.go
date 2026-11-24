package theme

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"
)

// Sizes, in dp.
const (
	TitleBarHeight   = 44
	StripeWidth      = 48
	StatusBarHeight  = 28
	ToolHeaderHeight = 38
	RowHeight        = 26
	TabStripHeight   = 38
	PopupRadius      = 12
	ControlRadius    = 6
)

// Label is a UI text style.
func Label(color string) giode.Styles {
	return giode.Styles{Color: color, FontSize: UISize, MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis}
}

// Code is a code text style.
func Code(color string) giode.Styles {
	return giode.Styles{Color: color, FontSize: CodeSize, FontFamily: CodeFont, MaxLines: 1, TabWidth: 4}
}

// PopupPanel is the frame of floating popups (Search Anything, branches).
func PopupPanel(width int) giode.Styles {
	c := Current()
	return giode.Styles{
		Width: properties.Width(width), Background: c.Raised, BackgroundGradient: giode.Linear(180, c.PopupTop, c.Raised),
		BorderRadius: 16, BorderWidth: 1, BorderColor: c.PopupLine,
		BoxShadow: giode.BoxShadow{Y: 24, Blur: 60, Spread: 2, Color: c.Shadow}, FlexDirection: giode.FlexDirectionColumn,
		Overflow: giode.OverflowHidden,
	}
}

// DialogPanel is the style of modal dialogs.
func DialogPanel(width int) giode.Styles {
	c := Current()
	return giode.Styles{Width: properties.Width(width), Background: c.Raised, BackgroundGradient: giode.Linear(180, c.PopupTop, c.Raised),
		BorderRadius: 16, BoxShadow: giode.BoxShadow{Y: 24, Blur: 60, Color: c.Shadow}}
}

// NewMenu creates a menu in the IDE's style: a raised card with the
// popups' edge and shadow.
func NewMenu() *giode.MenuElement {
	c := Current()
	return giode.Menu(giode.Styles{
		Background: c.MenuBg, Color: c.Text, BorderRadius: 12, FontSize: UISize, Padding: giode.UniformInset(5),
		BorderWidth: 1, BorderColor: c.PopupLine, BoxShadow: giode.BoxShadow{Y: 12, Blur: 32, Color: c.Shadow + "/0.8"},
	}).
		HoverColor(c.AccentSoft).SeparatorColor(c.PopupLine)
}

// InputStyles is the style of single-line fields.
func InputStyles() giode.Styles {
	c := Current()
	return giode.Styles{
		Background: c.InputBg, Color: c.Text, BorderWidth: 1, BorderColor: c.InputLine, BorderRadius: 8,
		FontSize: UISize, Padding: giode.SymmetricInset(12, 8),
	}
}

// PrimaryButton is a filled accent button.
func PrimaryButton() giode.Styles {
	c := Current()
	return giode.Styles{Background: c.Accent, BackgroundGradient: AccentGradient(), Color: c.AccentText, BorderRadius: 8,
		FontSize: UISize, FontWeight: giode.FontWeightSemiBold, Padding: giode.SymmetricInset(16, 8), BoxShadow: GlowShadow()}
}

// SecondaryButton is an outlined button.
func SecondaryButton() giode.Styles {
	c := Current()
	return giode.Styles{Background: c.Raised, BorderColor: c.PopupLine, BorderWidth: 1, Color: c.Text, BorderRadius: 8, FontSize: UISize,
		FontWeight: giode.FontWeightMedium, Padding: giode.SymmetricInset(16, 8)}
}

// Separator is a 1dp hairline.
func Separator() giode.Element {
	return giode.Box(giode.Styles{Width: -1, Height: 1, Background: Current().Separator})
}

// VSeparator is a vertical hairline.
func VSeparator() giode.Element {
	return giode.Box(giode.Styles{Width: 1, Height: -1, Background: Current().Separator})
}
