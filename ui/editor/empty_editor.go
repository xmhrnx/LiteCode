package editor

import (
	"github.com/xmhrnx/giode"

	"litecode/internal/keymap"
	"litecode/ui/core"
	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// emptyHints are the actions the empty editor suggests.
var emptyHints = []keymap.Action{keymap.SearchEverywhere, keymap.GotoFile, keymap.RecentFiles, keymap.FindInPath, keymap.GotoAction}

// EmptyEditor is what the editor shows with no file open: the project, and
// the shortcuts worth knowing as key caps on a glowing card.
type EmptyEditor struct {
	ide *core.IDE
}

// NewEmptyEditor creates the empty state.
func NewEmptyEditor(ide *core.IDE) *EmptyEditor {
	return &EmptyEditor{ide: ide}
}

func (e *EmptyEditor) hint(a keymap.Action) giode.Element {
	c := theme.Current()
	return giode.Row(giode.Styles{Width: -1, Height: 34, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
		giode.Text(keymap.Names[a], giode.Styles{Color: c.TextMuted, FontSize: theme.UISize, FlexGrow: 1}),
		widgets.Keycaps(e.ide.Shortcut(a)),
	)
}

func (e *EmptyEditor) card() giode.Element {
	c := theme.Current()
	var rows []giode.Element
	for i, a := range emptyHints {
		if i > 0 {
			rows = append(rows, giode.Box(giode.Styles{Width: -1, Height: 1, Background: c.Separator}))
		}
		rows = append(rows, e.hint(a))
	}
	return giode.Stack(giode.Styles{
		Width: 380, Padding: giode.SymmetricInset(18, 8), BorderRadius: 14, BorderWidth: 1, BorderColor: c.CardLine,
		BackgroundGradient: giode.Linear(160, c.TileFrom, c.TileTo),
		BoxShadow:          giode.BoxShadow{Y: 18, Blur: 50, Color: c.SoftGlow},
	}, rows...)
}

// Layout draws the empty state.
func (e *EmptyEditor) Layout(gtx giode.Context) giode.Dimensions {
	c := theme.Current()
	name := "LiteCode"
	if e.ide.Project != nil {
		name = e.ide.Project.Name
	}
	return giode.Stack(giode.Styles{
		Width: -1, Height: -1, Background: c.Editor, BackgroundGradient: giode.Linear(180, c.EditorTop, c.Editor, c.Editor),
		Align: giode.AlignCenter, Justify: giode.JustifyCenter, Gap: 6,
	},
		giode.Box(giode.Styles{BorderRadius: 16, BoxShadow: giode.BoxShadow{Y: 10, Blur: 34, Color: c.Glow}}, widgets.ProjectBadge(name, 56)),
		giode.Box(giode.Styles{Height: 6}),
		giode.Text(name, giode.Styles{Color: c.Text, FontSize: 20, FontWeight: giode.FontWeightSemiBold}),
		giode.Text("Open a file to start editing", giode.Styles{Color: c.TextFaint, FontSize: theme.UISize}),
		giode.Box(giode.Styles{Height: 18}),
		e.card(),
		giode.Box(giode.Styles{Height: 14}),
		giode.Text("Drop files here to open them", giode.Styles{Color: c.TextFaint, FontSize: theme.SmallSize}),
	).Layout(gtx)
}
