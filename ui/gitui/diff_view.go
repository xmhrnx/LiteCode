package gitui

import (
	"context"
	"fmt"

	"github.com/xmhrnx/giode"

	"litecode/internal/diff"
	"litecode/internal/keymap"
	"litecode/ui/core"
	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// DiffView compares two versions of a file side by side. It opens in an
// editor tab.
type DiffView struct {
	ide     *core.IDE
	req     core.DiffRequest
	rows    []diff.Row
	starts  []int
	current int // index into starts
	loaded  bool
	err     error
	list    *giode.VirtualListElement
	prev    *widgets.ToolButton
	next    *widgets.ToolButton
	reload  *widgets.ToolButton
}

// NewDiffView creates a viewer and starts loading both sides.
func NewDiffView(ide *core.IDE, req core.DiffRequest) *DiffView {
	v := &DiffView{ide: ide, req: req}
	v.list = giode.VirtualList(func() int { return len(v.rows) }, diffRowHeight,
		func(i int) giode.Element { return diffRow(v.rows[i]) },
		giode.Styles{ScrollBar: theme.ScrollBar(), Background: theme.Current().Editor})
	v.prev = widgets.NewToolButton("prev", "Previous Change ("+ide.Shortcut(keymap.PreviousDifference)+")", v.Prev)
	v.next = widgets.NewToolButton("next", "Next Change ("+ide.Shortcut(keymap.NextDifference)+")", v.Next)
	v.reload = widgets.NewToolButton("refresh", "Refresh", v.load)
	v.load()
	return v
}

func (v *DiffView) load() {
	go func() {
		ctx := context.Background()
		left, errL := v.req.Left(ctx)
		right, errR := v.req.Right(ctx)
		rows := diff.SideBySide(left, right)
		v.ide.App.Post(func() {
			v.loaded = true
			v.err = errL
			if v.err == nil {
				v.err = errR
			}
			v.rows, v.starts, v.current = rows, diff.ChangeStarts(rows), -1
			if len(v.starts) > 0 {
				v.Next()
			}
		})
	}()
}

// Next jumps to the next block of changes (F7).
func (v *DiffView) Next() { v.jump(1) }

// Prev jumps to the previous block of changes (Shift+F7).
func (v *DiffView) Prev() { v.jump(-1) }

func (v *DiffView) jump(step int) {
	if len(v.starts) == 0 {
		return
	}
	v.current = min(max(v.current+step, 0), len(v.starts)-1)
	v.list.ScrollToIndex(max(v.starts[v.current]-3, 0), giode.ScrollAlignStart)
}

func (v *DiffView) header() giode.Element {
	c := theme.Current()
	added, removed := diff.Stats(v.rows)
	summary := fmt.Sprintf("%d changes", len(v.starts))
	if len(v.starts) == 1 {
		summary = "1 change"
	}
	if v.loaded && len(v.starts) == 0 {
		summary = "Contents are identical"
	}
	labels := giode.Row(giode.Styles{Width: -1, Wrap: giode.WrapNoWrap, Padding: giode.SymmetricInset(12, 4)},
		giode.Text(v.req.LeftLabel, giode.Styles{Color: c.TextMuted, FontSize: theme.SmallSize, FlexGrow: 1}),
		giode.Text(v.req.RightLabel, giode.Styles{Color: c.TextMuted, FontSize: theme.SmallSize, FlexGrow: 1}),
	)
	bar := giode.Row(giode.Styles{Width: -1, Height: 34, Align: giode.AlignCenter, Gap: 4, Wrap: giode.WrapNoWrap, Padding: giode.SymmetricInset(8, 0)},
		v.prev, v.next, v.reload,
		giode.Box(giode.Styles{Width: 8}),
		giode.Text(v.req.Path, theme.Label(c.Text)),
		giode.Box(giode.Styles{FlexGrow: 1}),
		giode.Text(fmt.Sprintf("+%d", added), giode.Styles{Color: c.StatusAdded, FontSize: theme.SmallSize}),
		giode.Text(fmt.Sprintf("-%d", removed), giode.Styles{Color: c.StatusUntracked, FontSize: theme.SmallSize}),
		giode.Box(giode.Styles{Width: 8}),
		giode.Text(summary, giode.Styles{Color: c.TextMuted, FontSize: theme.SmallSize}),
	)
	return giode.Stack(giode.Styles{Width: -1, Background: c.Panel}, bar, theme.Separator(), labels, theme.Separator())
}

// Layout draws the viewer.
func (v *DiffView) Layout(gtx giode.Context) giode.Dimensions {
	c := theme.Current()
	var body giode.Element = v.list
	switch {
	case v.err != nil:
		body = giode.Box(giode.Styles{Padding: giode.UniformInset(16)}, giode.Text(v.err.Error(), giode.Styles{Color: c.Danger}))
	case !v.loaded:
		body = giode.Box(giode.Styles{Width: -1, Height: -1, Align: giode.AlignCenter, Justify: giode.JustifyCenter}, giode.Spinner())
	}
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Background: c.Editor},
		v.header(),
		giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, body),
	).Layout(gtx)
}
