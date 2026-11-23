package frame

import (
	"github.com/xmhrnx/giode"

	"litecode/internal/keymap"
	"litecode/ui/core"
	"litecode/ui/theme"
)

// BranchWidget shows the current branch in the title bar and opens the
// branches popup. It's empty outside a git repository.
type BranchWidget struct {
	ide    *core.IDE
	button *giode.PressableElement
}

// NewBranchWidget creates the widget.
func NewBranchWidget(ide *core.IDE) *BranchWidget {
	w := &BranchWidget{ide: ide}
	w.button = widgetButton(
		func() giode.Element { return theme.Icon("branch", theme.Current().TextMuted) },
		func() string {
			if b := ide.Project.Git.Branch; b != "" {
				return b
			}
			return "Git"
		},
		func(giode.ClickEvent) { ide.Run(keymap.Branches) })
	return w
}

// Layout draws the widget.
func (w *BranchWidget) Layout(gtx giode.Context) giode.Dimensions {
	if w.ide.Project == nil || w.ide.Project.Git == nil {
		return giode.Dimensions{}
	}
	return w.button.Layout(gtx)
}
