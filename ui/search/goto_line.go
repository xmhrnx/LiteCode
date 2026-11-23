package search

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/xmhrnx/giode"

	"litecode/ui/widgets"
)

// GotoLine is the Ctrl+G dialog: "line" or "line:column".
type GotoLine struct {
	prompt *widgets.PromptDialog
}

// NewGotoLine creates the dialog.
func NewGotoLine() *GotoLine {
	return &GotoLine{prompt: widgets.NewPromptDialog(nil)}
}

// Ask opens the dialog at the caret's position; goTo gets the answer.
func (g *GotoLine) Ask(line, col int, goTo func(line, col int)) {
	g.prompt.Ask("Go to Line:Column", fmt.Sprintf("%d:%d", line, col), func(s string) {
		parts := strings.SplitN(strings.TrimSpace(s), ":", 2)
		l, err := strconv.Atoi(parts[0])
		if err != nil || l < 1 {
			return
		}
		c := 1
		if len(parts) == 2 {
			if n, err := strconv.Atoi(parts[1]); err == nil && n > 0 {
				c = n
			}
		}
		goTo(l, c)
	})
}

// Layout lays the dialog out; it takes no space.
func (g *GotoLine) Layout(gtx giode.Context) giode.Dimensions {
	return g.prompt.Layout(gtx)
}
