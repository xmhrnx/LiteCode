package assistant

import (
	"github.com/xmhrnx/giode"

	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// sendButton is the round button of the composer: send, or stop while a
// reply streams.
type sendButton struct {
	busy  func() bool
	press *giode.PressableElement
}

func newSendButton(busy func() bool, onSend, onStop func()) *sendButton {
	b := &sendButton{busy: busy}
	b.press = giode.Pressable(func(s giode.PressState) giode.Element {
		c := theme.Current()
		st := giode.Styles{Width: 30, Height: 30, BorderRadius: 15, Align: giode.AlignCenter, Justify: giode.JustifyCenter}
		icon := "send"
		if b.busy() {
			icon = "stop"
			st.Background = c.Pressed
		} else {
			st.BackgroundGradient = theme.AccentGradient()
			if s.Hovered {
				st.BoxShadow = theme.GlowShadow()
			}
		}
		return giode.Box(st, theme.IconSized(icon, c.AccentText, 15))
	}).Label("Send").Focusable(false).OnClick(widgets.PrimaryClick(func(giode.ClickEvent) {
		if b.busy() {
			onStop()
			return
		}
		onSend()
	}))
	return b
}

// Layout draws the button.
func (b *sendButton) Layout(gtx giode.Context) giode.Dimensions {
	return b.press.Layout(gtx)
}
