package gitui

import (
	"github.com/xmhrnx/giode"

	"litecode/ui/theme"
)

// refLabel draws a branch or tag name pointing at a commit, colored by kind:
// the current branch, other local branches, remote branches, tags.
func refLabel(ref, current string, tag, local bool) giode.Element {
	c := theme.Current()
	color := c.RefRemote
	switch {
	case tag:
		color = c.RefTag
	case ref == current:
		color = c.RefHead
	case local:
		color = c.RefLocal
	}
	return giode.Box(giode.Styles{
		Background: giode.Mix(color, c.Editor, 0.18), BorderRadius: 4, Padding: giode.SymmetricInset(5, 1),
	}, giode.Text(ref, giode.Styles{Color: color, FontSize: 11, MaxLines: 1}))
}
