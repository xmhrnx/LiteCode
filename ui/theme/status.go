package theme

import "litecode/internal/git"

// StatusColor is the label color of a file in a VCS status.
func StatusColor(k git.ChangeKind) string {
	c := Current()
	switch k {
	case git.Added:
		return c.StatusAdded
	case git.Deleted:
		return c.StatusDeleted
	case git.Untracked:
		return c.StatusUntracked
	case git.Conflicted:
		return c.StatusConflict
	default:
		return c.StatusModified
	}
}
