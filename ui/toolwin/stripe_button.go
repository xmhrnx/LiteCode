package toolwin

import (
	"github.com/xmhrnx/giode"

	"litecode/internal/layout"
	"litecode/ui/core"
	"litecode/ui/theme"
)

// StripeButton opens and closes one panel from the activity bar; a
// right-click offers to move it to another side.
type StripeButton struct {
	ide   *core.IDE
	id    layout.WindowID
	press *giode.PressableElement
	tip   *giode.TooltipElement
	menu  *giode.MenuElement
}

// NewStripeButton creates the button of a panel.
func NewStripeButton(ide *core.IDE, w core.ToolWindow) *StripeButton {
	b := &StripeButton{ide: ide, id: w.ID(), menu: newWindowMenu(ide, w.ID())}
	b.press = giode.Pressable(func(s giode.PressState) giode.Element {
		c := theme.Current()
		open := ide.Layout.IsOpen(b.id)
		st := giode.Styles{Width: 34, Height: 34, BorderRadius: 10, Align: giode.AlignCenter, Justify: giode.JustifyCenter}
		fg := c.TextFaint
		switch {
		case open:
			st.BackgroundGradient, st.BoxShadow, fg = theme.AccentGradient(), theme.GlowShadow(), c.AccentText
		case s.Hovered:
			st.Background, fg = c.Hover, c.Text
		}
		return giode.Box(st, theme.IconSized(w.Icon(), fg, 18))
	}).Label(w.Title()).
		OnPrimaryClick(func(giode.ClickEvent) { ide.Run(panelActions[b.id]) }).
		OnSecondaryClick(b.menu.OpenAtClick)
	b.tip = giode.Tooltip(b.press, w.Title()+shortcutHint(ide, w.ID()))
	return b
}

// shortcutHint is " (Alt+1)" for a panel whose toggle action has a stroke.
func shortcutHint(ide *core.IDE, id layout.WindowID) string {
	if s := ide.Shortcut(panelActions[id]); s != "" {
		return "  " + s
	}
	return ""
}

// Layout draws the button and keeps its menu ready.
func (b *StripeButton) Layout(gtx giode.Context) giode.Dimensions {
	return giode.Box(giode.Styles{}, b.tip, b.menu).Layout(gtx)
}
