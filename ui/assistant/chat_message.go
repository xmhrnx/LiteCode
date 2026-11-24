package assistant

import (
	"github.com/xmhrnx/giode"

	"litecode/internal/llm"
	"litecode/ui/theme"
)

// chatMessage is one turn of the conversation: the user's as a tinted
// bubble on the right, the assistant's as Markdown with a copy button.
type chatMessage struct {
	key  int // the message's index, for text selection
	msg  llm.Message
	note string // a file shared with this message
	copy *giode.CopyButtonElement
	live bool // still streaming
}

func newChatMessage(key int) *chatMessage {
	m := &chatMessage{key: key}
	c := theme.Current()
	m.copy = giode.CopyButton(func() string { return m.msg.Content },
		giode.Styles{Color: c.TextFaint, FontSize: 11, Padding: giode.SymmetricInset(6, 2)}).Labels("Copy", "Copied")
	return m
}

func (m *chatMessage) user() giode.Element {
	c := theme.Current()
	kids := []giode.Element{}
	if m.note != "" {
		kids = append(kids, giode.Row(giode.Styles{Gap: 5, Align: giode.AlignCenter},
			theme.IconSized("file-code", c.TextAccent, 12),
			giode.Text(m.note, giode.Styles{Color: c.TextAccent, FontSize: 11})))
	}
	kids = append(kids, giode.Text(m.msg.Content, giode.Styles{Color: c.Text, FontSize: theme.UISize}).Selectable(m.key))
	return giode.Row(giode.Styles{Width: -1, Justify: giode.JustifyEnd},
		giode.Stack(giode.Styles{
			MaxWidth: 520, Gap: 6, Padding: giode.SymmetricInset(14, 10), BorderRadius: 14,
			BackgroundGradient: giode.Linear(135, c.BubbleFrom, c.BubbleTo), BorderWidth: 1, BorderColor: c.AccentLine,
		}, kids...))
}

func (m *chatMessage) assistant() giode.Element {
	c := theme.Current()
	head := []giode.Element{
		giode.Box(giode.Styles{Width: 22, Height: 22, BorderRadius: 7, BackgroundGradient: theme.AccentGradient(), Align: giode.AlignCenter, Justify: giode.JustifyCenter},
			theme.IconSized("panel-assistant", c.AccentText, 13)),
		giode.Text("Assistant", giode.Styles{Color: c.TextMuted, FontSize: theme.SmallSize, FontWeight: giode.FontWeightMedium, FlexGrow: 1}),
	}
	if m.live {
		head = append(head, giode.Spinner(giode.Styles{Width: 12, Height: 12, Color: c.TextAccent}))
	} else {
		head = append(head, m.copy)
	}
	body := m.msg.Content
	if body == "" && m.live {
		body = "…"
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 8},
		giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, head...),
		giode.Box(giode.Styles{Padding: giode.Inset{Left: 30}},
			giode.Markdown(body, giode.MarkdownOptions{
				Styles:         giode.Styles{Color: c.TextCode, FontSize: theme.UISize, LineHeight: 20},
				CodeFont:       theme.CodeFont,
				CodeBackground: c.Editor,
				LinkColor:      c.TextAccent,
				Selectable:     m.key,
			})),
	)
}

// Layout draws the message.
func (m *chatMessage) Layout(gtx giode.Context) giode.Dimensions {
	if m.msg.Role == llm.User {
		return m.user().Layout(gtx)
	}
	return m.assistant().Layout(gtx)
}
