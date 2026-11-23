// Package welcome is the window shown with no project open: recent projects,
// Open Folder and Clone Repository.
package welcome

import (
	"fmt"
	"time"

	"github.com/xmhrnx/giode"

	"litecode/internal/project"
	"litecode/internal/recent"
	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// recentProjectRow is one recent project as a card: badge, name, path and
// when it was last opened. A click opens it; a right-click offers to remove it.
type recentProjectRow struct {
	p     recent.Project
	press *giode.PressableElement
}

func newRecentProjectRow(onOpen func(recent.Project), onMenu func(recent.Project, giode.ClickEvent)) *recentProjectRow {
	r := &recentProjectRow{}
	r.press = giode.Pressable(r.view).
		OnClick(widgets.PrimaryClick(func(giode.ClickEvent) { onOpen(r.p) })).
		OnSecondaryClick(func(e giode.ClickEvent) { onMenu(r.p, e) })
	return r
}

// ago writes how long ago t was, roughly.
func ago(t, now time.Time) string {
	d := now.Sub(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%d min ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%d h ago", int(d.Hours()))
	case d < 48*time.Hour:
		return "yesterday"
	default:
		return t.Format("Jan 2")
	}
}

func (r *recentProjectRow) view(s giode.PressState) giode.Element {
	c := theme.Current()
	st := giode.Styles{
		Width: -1, Height: 118, BorderRadius: 14, BorderWidth: 1, BorderColor: c.CardLine, Background: c.Card,
		FlexDirection: giode.FlexDirectionColumn, Gap: 4, Padding: giode.UniformInset(16),
	}
	if s.Hovered {
		st.BorderColor = c.FocusLine
		st.BackgroundGradient = giode.Linear(160, "#211D3D", c.Card)
		st.BoxShadow = giode.BoxShadow{Y: 10, Blur: 28, Color: "#5B4BE040"}
	}
	return giode.Box(st,
		giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			widgets.ProjectBadge(r.p.Name, 34),
			giode.Box(giode.Styles{FlexGrow: 1}),
			giode.Text(ago(r.p.OpenedAt, time.Now()), giode.Styles{Color: c.TextFaint, FontSize: 11})),
		giode.Box(giode.Styles{Height: 6}),
		giode.Text(r.p.Name, giode.Styles{Color: c.Text, FontSize: 14, FontWeight: giode.FontWeightSemiBold, MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis}),
		giode.Text(project.ShortPath(r.p.Path), giode.Styles{Color: c.TextFaint, FontSize: theme.SmallSize, MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis}),
	)
}

// Layout draws the card.
func (r *recentProjectRow) Layout(gtx giode.Context) giode.Dimensions {
	return r.press.Layout(gtx)
}
