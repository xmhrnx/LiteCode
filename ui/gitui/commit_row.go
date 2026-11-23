package gitui

import (
	"time"

	"github.com/xmhrnx/giode"

	"litecode/internal/git"
	"litecode/ui/theme"
)

// commitRow is one commit of the log: subject with its branch and tag labels,
// author and date.
type commitRow struct {
	commit   git.Commit
	current  string          // the checked-out branch, for its label color
	locals   map[string]bool // local branch names, to tell them from remote ones
	first    bool            // the top row, where the rail starts
	last     bool            // the bottom row, where it ends
	selected bool
	press    *giode.PressableElement
}

func newCommitRow(onSelect func(git.Commit), onMenu func(git.Commit, giode.ClickEvent)) *commitRow {
	r := &commitRow{}
	r.press = giode.Pressable(r.view).
		OnClick(func(giode.ClickEvent) { onSelect(r.commit) }).
		OnSecondaryClick(func(e giode.ClickEvent) {
			onSelect(r.commit)
			onMenu(r.commit, e)
		})
	return r
}

func (r *commitRow) view(s giode.PressState) giode.Element {
	c := theme.Current()
	bg := ""
	switch {
	case r.selected:
		bg = c.Selection
	case s.Hovered:
		bg = c.Hover
	}
	subject := []giode.Element{
		commitRail(r.first, r.last, len(r.commit.Refs) > 0),
		giode.Text(r.commit.Subject, theme.Label(c.Text)),
	}
	for _, ref := range r.commit.Refs {
		subject = append(subject, refLabel(ref, r.current, false, r.locals[ref]))
	}
	for _, tag := range r.commit.Tags {
		subject = append(subject, refLabel(tag, r.current, true, false))
	}
	return giode.Row(giode.Styles{Width: -1, Height: theme.RowHeight, Background: bg, BackgroundGradient: theme.RowWash(r.selected), BorderRadius: 7, Margin: giode.SymmetricInset(6, 0), Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Padding: giode.SymmetricInset(10, 0), Gap: 12},
		giode.Row(giode.Styles{FlexGrow: 1, Height: -1, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Overflow: giode.OverflowHidden}, subject...),
		giode.Text(r.commit.Author, giode.Styles{Color: c.TextMuted, FontSize: theme.UISize, Width: 130, MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis}),
		giode.Text(commitDate(r.commit.Time, time.Now()), giode.Styles{Color: c.TextMuted, FontSize: theme.UISize, Width: 120, MaxLines: 1}),
	)
}

// Layout draws the row.
func (r *commitRow) Layout(gtx giode.Context) giode.Dimensions {
	return r.press.Layout(gtx)
}
