package settings

import (
	"strings"

	"github.com/xmhrnx/giode"

	"litecode/internal/keymap"
	"litecode/ui/theme"
)

// keymapRow is one action of the Keymap page: its name and its strokes.
type keymapRow struct {
	action    keymap.Action
	strokes   []string
	recording bool
	conflict  string
	press     *giode.PressableElement
}

func newKeymapRow(onClick func(r *keymapRow)) *keymapRow {
	r := &keymapRow{}
	r.press = giode.Pressable(r.view).Focusable(false).OnClick(func(giode.ClickEvent) { onClick(r) })
	return r
}

func (r *keymapRow) view(s giode.PressState) giode.Element {
	c := theme.Current()
	bg := ""
	switch {
	case r.recording:
		bg = c.Selection
	case s.Hovered:
		bg = c.Hover
	}
	var strokes []string
	for _, st := range r.strokes {
		strokes = append(strokes, formatStroke(st))
	}
	right := strings.Join(strokes, ", ")
	color := c.TextAccent
	if r.recording {
		right, color = "Press a key combination… (Backspace removes, Esc cancels)", c.Text
	}
	kids := []giode.Element{
		giode.Text(keymap.Names[r.action], giode.Styles{Color: c.Text, FontSize: theme.UISize, FlexGrow: 1}),
	}
	if r.conflict != "" {
		kids = append(kids, giode.Text(r.conflict, giode.Styles{Color: c.Warning, FontSize: theme.SmallSize}))
	}
	kids = append(kids, giode.Text(right, giode.Styles{Color: color, FontSize: theme.UISize}))
	return giode.Row(giode.Styles{Width: -1, Height: 28, Background: bg, BackgroundGradient: theme.RowWash(r.recording), BorderRadius: 7, Align: giode.AlignCenter, Gap: 12, Padding: giode.SymmetricInset(10, 0), Wrap: giode.WrapNoWrap}, kids...)
}

// Layout draws the row.
func (r *keymapRow) Layout(gtx giode.Context) giode.Dimensions {
	return r.press.Layout(gtx)
}
