// Package assistant is the Assistant panel: a chat with the model chosen in
// Settings › AI Assistant, optionally about the open file.
package assistant

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/xmhrnx/giode"

	"litecode/internal/layout"
	"litecode/internal/llm"
	"litecode/ui/core"
	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// WindowID is the Assistant panel's ID.
const WindowID layout.WindowID = "assistant"

// maxSharedFile is the largest file sent with a message, in bytes. Larger
// files aren't cut down silently: the message says why they weren't sent.
const maxSharedFile = 256 << 10

// entry is one turn: what's sent to the model and what's shown.
type entry struct {
	msg     llm.Message
	display string
	note    string
	live    bool
}

// Window is the Assistant panel.
type Window struct {
	ide     *core.IDE
	entries []*entry
	views   *giode.KeyedCache[int, *chatMessage]
	list    *giode.ScrollElement
	input   *giode.InputElement
	send    *sendButton
	share   *contextChip
	actions []giode.Element
	connect *giode.ButtonElement
	cancel  context.CancelFunc
	errMsg  string
}

// New creates the panel.
func New(ide *core.IDE) *Window {
	w := &Window{ide: ide, views: giode.NewKeyedCache[int, *chatMessage]()}
	c := theme.Current()
	w.list = giode.Scroll(w.conversation, giode.Styles{ScrollBar: theme.ScrollBar()}).Anchor(giode.AnchorBottom)
	w.input = widgets.TextField("Ask about your code…", giode.Styles{
		Width: -1, Background: "transparent", BorderColor: "transparent", BorderRadius: 1, Color: c.Text, FontSize: theme.UISize,
		Padding: giode.SymmetricInset(4, 6),
	}).Multiline(true).Lines(1, 8).SubmitOn(giode.SubmitEnterUnlessShift).Submit(func(string) { w.Send() })
	w.send = newSendButton(w.busy, w.Send, w.Stop)
	w.share = newContextChip(func() string {
		if ide.CurrentFile == nil {
			return ""
		}
		path, _ := ide.CurrentFile()
		return path
	})
	w.connect = giode.Button("Connect a Model", theme.PrimaryButton()).OnClick(w.openSettings)
	w.actions = []giode.Element{
		widgets.NewToolButton("add", "New Chat", w.NewChat),
		widgets.NewToolButton("settings", "AI Settings", w.openSettings),
	}
	return w
}

func (w *Window) ID() layout.WindowID            { return WindowID }
func (w *Window) Title() string                  { return "Assistant" }
func (w *Window) Icon() string                   { return "panel-assistant" }
func (w *Window) HeaderActions() []giode.Element { return w.actions }

// Focus puts the keyboard in the composer.
func (w *Window) Focus() { w.input.Focus() }

func (w *Window) busy() bool { return w.cancel != nil }

func (w *Window) openSettings() {
	if w.ide.OpenSettings != nil {
		w.ide.OpenSettings("AI Assistant")
	}
}

// provider returns the active provider and model.
func (w *Window) provider() (llm.Provider, string, bool) {
	s := w.ide.AI.Settings
	p, ok := s.Provider(s.Active)
	if !ok || p.DefaultModel == "" {
		return p, "", false
	}
	return p, p.DefaultModel, true
}

// NewChat stops any reply and clears the conversation.
func (w *Window) NewChat() {
	w.Stop()
	w.entries, w.errMsg = nil, ""
	w.input.Focus()
}

// Stop cancels the reply being streamed.
func (w *Window) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
}

// withFile adds the open file to a message, or explains why it can't.
func (w *Window) withFile(text string) (content, note string, err error) {
	if !w.share.Sharing() {
		return text, "", nil
	}
	path, body := w.ide.CurrentFile()
	name := filepath.Base(path)
	if w.ide.Project != nil {
		name = w.ide.Project.Rel(path)
	}
	if len(body) > maxSharedFile {
		return "", "", fmt.Errorf("%s is %d KB, over the %d KB the assistant sends; turn off file sharing or ask about a smaller file",
			name, len(body)>>10, maxSharedFile>>10)
	}
	lang := w.ide.Langs.ForPath(path).ID
	content = fmt.Sprintf("%s\n\nOpen file `%s`:\n```%s\n%s\n```", text, name, lang, body)
	return content, name, nil
}

// Send sends the composer's text.
func (w *Window) Send() {
	text := strings.TrimSpace(w.input.Text())
	if text == "" || w.busy() {
		return
	}
	p, model, ok := w.provider()
	if !ok {
		w.errMsg = "Choose a provider and a model in Settings › AI Assistant."
		return
	}
	content, note, err := w.withFile(text)
	if err != nil {
		w.errMsg = err.Error()
		return
	}
	w.errMsg = ""
	w.input.SetText("")
	w.entries = append(w.entries, &entry{msg: llm.Message{Role: llm.User, Content: content}, display: text, note: note})
	reply := &entry{msg: llm.Message{Role: llm.Assistant}, live: true}
	w.entries = append(w.entries, reply)
	w.list.ScrollToBottom()
	req := llm.Request{
		Model: model, System: w.ide.AI.Settings.SystemPrompt, MaxTokens: w.ide.AI.Settings.MaxTokens,
		Messages: w.history(),
	}
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel
	go w.stream(ctx, p, req, reply)
}

// history is the conversation to send: every finished turn, then the new one.
func (w *Window) history() []llm.Message {
	var msgs []llm.Message
	for _, e := range w.entries {
		if e.live {
			continue
		}
		msgs = append(msgs, e.msg)
	}
	return msgs
}

// stream runs off the UI goroutine. Text is batched: at most one post is
// pending at a time, so a fast model doesn't flood the UI goroutine.
func (w *Window) stream(ctx context.Context, p llm.Provider, req llm.Request, reply *entry) {
	var mu sync.Mutex
	var pending strings.Builder
	posted := false
	flush := func() {
		mu.Lock()
		text := pending.String()
		pending.Reset()
		posted = false
		mu.Unlock()
		reply.msg.Content += text
	}
	client, err := core.ClientFor(p)
	if err == nil {
		err = client.Stream(ctx, req, func(s string) {
			mu.Lock()
			pending.WriteString(s)
			post := !posted
			posted = true
			mu.Unlock()
			if post {
				w.ide.App.Post(flush)
			}
		})
	}
	w.ide.App.Post(func() {
		flush()
		w.cancel = nil
		reply.live = false
		if err == nil {
			return
		}
		if errors.Is(err, context.Canceled) {
			if reply.msg.Content == "" {
				w.dropReply(reply)
			}
			return
		}
		w.errMsg = core.DescribeAIError(p, err)
		if reply.msg.Content == "" {
			w.dropReply(reply)
		}
	})
}

// dropReply removes an empty reply and the message it answered, so the
// conversation stays user/assistant alternating; the text goes back in the
// composer for another try.
func (w *Window) dropReply(reply *entry) {
	for i, e := range w.entries {
		if e == reply && i > 0 {
			w.input.SetText(w.entries[i-1].display)
			w.entries = w.entries[:i-1]
			return
		}
	}
}

func (w *Window) conversation() giode.Element {
	var kids []giode.Element
	for i, e := range w.entries {
		m := w.views.Get(i, func() *chatMessage { return newChatMessage(i) })
		m.msg, m.live, m.note = e.msg, e.live, e.note
		if e.msg.Role == llm.User {
			m.msg.Content = e.display
		}
		kids = append(kids, m)
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 18, Padding: giode.UniformInset(14)}, kids...)
}

func (w *Window) emptyState() giode.Element {
	c := theme.Current()
	p, model, ok := w.provider()
	kids := []giode.Element{
		giode.Box(giode.Styles{Width: 48, Height: 48, BorderRadius: 14, BackgroundGradient: theme.AccentGradient(), BoxShadow: theme.GlowShadow(), Align: giode.AlignCenter, Justify: giode.JustifyCenter},
			theme.IconSized("panel-assistant", c.AccentText, 24)),
		giode.Text("Ask anything about your code", giode.Styles{Color: c.Text, FontSize: 15, FontWeight: giode.FontWeightSemiBold}),
	}
	if ok {
		kids = append(kids, giode.Text(p.Name+" · "+model, giode.Styles{Color: c.TextFaint, FontSize: theme.SmallSize}),
			giode.Text("The open file is shared with your message while its chip is on.", giode.Styles{Color: c.TextFaint, FontSize: theme.SmallSize, TextAlign: giode.TextAlignCenter}))
	} else {
		kids = append(kids, giode.Text("Use any hosted provider with your own key, a local model, or a custom endpoint.",
			giode.Styles{Color: c.TextMuted, FontSize: theme.SmallSize, TextAlign: giode.TextAlignCenter}), giode.Box(giode.Styles{Height: 4}), w.connect)
	}
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Align: giode.AlignCenter, Justify: giode.JustifyCenter, Gap: 8, Padding: giode.UniformInset(24)}, kids...)
}

func (w *Window) composer() giode.Element {
	c := theme.Current()
	model := "No model"
	if p, m, ok := w.provider(); ok {
		model = p.Name + " · " + m
	}
	var kids []giode.Element
	if w.errMsg != "" {
		kids = append(kids, giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Padding: giode.SymmetricInset(10, 8), BorderRadius: 10,
			Background: giode.Mix(c.Danger, c.Raised, 0.1), BorderWidth: 1, BorderColor: giode.Mix(c.Danger, c.Raised, 0.35), Wrap: giode.WrapNoWrap},
			theme.IconSized("alert", c.Danger, 14), giode.Text(w.errMsg, giode.Styles{Color: c.Text, FontSize: theme.SmallSize, FlexGrow: 1})))
	}
	kids = append(kids, giode.Stack(giode.Styles{
		Width: -1, Gap: 6, Padding: giode.UniformInset(10), BorderRadius: 14, BorderWidth: 1, BorderColor: c.PopupLine, Background: c.Editor,
	},
		w.input,
		giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			w.share,
			giode.Text(model, giode.Styles{Color: c.TextFaint, FontSize: 11, FlexGrow: 1, MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis, TextAlign: giode.TextAlignEnd}),
			w.send),
	))
	return giode.Stack(giode.Styles{Width: -1, Gap: 8, Padding: giode.UniformInset(10)}, kids...)
}

// View draws the panel.
func (w *Window) View() giode.Element {
	var body giode.Element = w.list
	if len(w.entries) == 0 {
		body = w.emptyState()
	}
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Background: theme.Current().Panel},
		giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, body),
		w.composer(),
	)
}
