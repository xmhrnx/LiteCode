package assistant

import (
	"path/filepath"

	"github.com/xmhrnx/giode"

	"litecode/ui/theme"
)

// contextChip toggles sharing the open file with the next message.
type contextChip struct {
	on    bool
	file  func() string
	press *giode.PressableElement
	tip   *giode.TooltipElement
}

func newContextChip(file func() string) *contextChip {
	ch := &contextChip{on: true, file: file}
	ch.press = giode.Pressable(func(s giode.PressState) giode.Element {
		c := theme.Current()
		name := ch.file()
		if name == "" {
			return giode.Text("No file open", giode.Styles{Color: c.TextFaint, FontSize: 11})
		}
		st := giode.Styles{Gap: 5, Align: giode.AlignCenter, BorderRadius: 10, Padding: giode.SymmetricInset(8, 3), BorderWidth: 1, BorderColor: c.PopupLine}
		fg := c.TextFaint
		if ch.on {
			st.Background, st.BorderColor, fg = c.AccentSoft, "#463C96", c.TextAccent
		} else if s.Hovered {
			st.Background = c.Hover
		}
		return giode.Row(st, theme.IconSized("file-code", fg, 12), giode.Text(filepath.Base(name), giode.Styles{Color: fg, FontSize: 11}))
	}).Label("Share open file").Focusable(false).OnPrimaryClick(func(giode.ClickEvent) { ch.on = !ch.on })
	ch.tip = giode.Tooltip(ch.press, "Send the open file with your message")
	return ch
}

// Sharing reports whether the open file goes with the next message.
func (ch *contextChip) Sharing() bool { return ch.on && ch.file() != "" }

// Layout draws the chip.
func (ch *contextChip) Layout(gtx giode.Context) giode.Dimensions {
	return ch.tip.Layout(gtx)
}
