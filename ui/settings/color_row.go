package settings

import (
	"github.com/xmhrnx/giode"

	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// colorRow is one color to change by hand: a swatch, its name and value,
// and when opened a color picker. A changed color has a reset button.
type colorRow struct {
	label   string
	value   func() string // the color in effect
	changed func() bool   // whether the user set it
	set     func(string)  // "" resets it
	open    bool
	press   *giode.PressableElement
	picker  *giode.ColorPickerElement
	reset   *widgets.ToolButton
}

func newColorRow(label string, value func() string, changed func() bool, set func(string)) *colorRow {
	r := &colorRow{label: label, value: value, changed: changed, set: set}
	r.press = giode.Pressable(r.header).Focusable(false).Label(label).
		OnPrimaryClick(func(giode.ClickEvent) { r.toggle() })
	r.reset = widgets.NewToolButton("revert", "Reset to the theme's color", func() {
		set("")
		if r.picker != nil {
			r.picker.SetValue(theme.Hex(value()))
		}
	})
	return r
}

func (r *colorRow) toggle() {
	r.open = !r.open
	if r.open && r.picker == nil {
		r.picker = giode.Input.Color(theme.Hex(r.value()), giode.Styles{Width: 260}).
			OnChange(func(c string) { r.set(theme.Hex(c)) })
	} else if r.open {
		r.picker.SetValue(theme.Hex(r.value()))
	}
}

func (r *colorRow) header(st giode.PressState) giode.Element {
	c := theme.Current()
	bg := "transparent"
	if st.Hovered || r.open {
		bg = c.Hover
	}
	kids := []giode.Element{
		giode.Box(giode.Styles{Width: 18, Height: 18, BorderRadius: 5, Background: r.value(), BorderWidth: 1, BorderColor: c.PopupLine}),
		giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(r.label, giode.Styles{Color: c.Text, FontSize: theme.UISize})),
		giode.Text(theme.Hex(r.value()), theme.Code(c.TextMuted)),
	}
	if r.changed() {
		kids = append(kids, r.reset)
	} else {
		kids = append(kids, giode.Box(giode.Styles{Width: 24}))
	}
	return giode.Row(giode.Styles{Width: -1, Height: 30, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		Padding: giode.SymmetricInset(8, 0), BorderRadius: 6, Background: bg, Cursor: giode.CursorPointer}, kids...)
}

// Layout draws the row, and the picker under it when it's open.
func (r *colorRow) Layout(gtx giode.Context) giode.Dimensions {
	if !r.open {
		return r.press.Layout(gtx)
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 6},
		r.press,
		giode.Box(giode.Styles{Padding: giode.Inset{Left: 36, Bottom: 8}}, r.picker),
	).Layout(gtx)
}
