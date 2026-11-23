package welcome

import (
	"github.com/xmhrnx/giode"

	"litecode/internal/recent"
	"litecode/ui/core"
	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// Screen is the welcome window's content.
type Screen struct {
	ide    *core.IDE
	bar    *giode.TitleBarElement
	search *giode.InputElement
	open   *giode.ButtonElement
	clone  *giode.ButtonElement
	cloneD *CloneDialog
	list   *giode.ScrollElement
	rows   *giode.KeyedCache[string, *recentProjectRow]
}

// NewScreen creates the welcome screen.
func NewScreen(ide *core.IDE) *Screen {
	s := &Screen{ide: ide, cloneD: NewCloneDialog(ide), rows: giode.NewKeyedCache[string, *recentProjectRow]()}
	c := theme.Current()
	s.bar = giode.TitleBar(ide.App, "", giode.Styles{Background: "transparent", Color: c.TextMuted, Height: theme.TitleBarHeight, FontSize: theme.UISize})
	if giode.IsPlatform(giode.PlatformMacOS) {
		s.bar.Variant(giode.TitleBarMac)
	}
	search := theme.InputStyles()
	search.Background = c.Panel
	s.search = widgets.TextField("Search projects", search)
	s.open = giode.Button("Open Folder", theme.PrimaryButton()).
		Icon(theme.Icon("open", c.AccentText)).OnClick(ide.ChooseProject)
	s.clone = giode.Button("Clone Repository", theme.SecondaryButton()).Variant(giode.VariantOutlined).
		Icon(theme.Icon("clone", c.TextMuted)).OnClick(s.cloneD.Open)
	s.list = giode.Scroll(s.listView, giode.Styles{ScrollBar: theme.ScrollBar()})
	return s
}

func (s *Screen) menu(p recent.Project, e giode.ClickEvent) {
	path := p.Path
	theme.NewMenu().
		Item("Open", func() { s.ide.OpenProject(path) }).
		Item("Copy Path", func() { s.ide.App.SetClipboard(path) }).
		Separator().
		Item("Remove from Recent", func() { s.ide.Error(s.ide.Recent.Remove(path)) }).OpenAtClick(e)
}

func (s *Screen) listView() giode.Element {
	c := theme.Current()
	projects := s.ide.Recent.Filter(s.search.Text())
	if len(projects) == 0 {
		return giode.Stack(giode.Styles{Width: -1, Padding: giode.UniformInset(40), Align: giode.AlignCenter, Gap: 6},
			giode.Text("No recent projects", giode.Styles{Color: c.TextMuted, FontSize: 14}),
			giode.Text("Open a folder or clone a repository to start.", giode.Styles{Color: c.TextFaint, FontSize: theme.UISize}))
	}
	var kids []giode.Element
	for _, p := range projects {
		r := s.rows.Get(p.Path, func() *recentProjectRow {
			return newRecentProjectRow(func(p recent.Project) { s.ide.OpenProject(p.Path) }, s.menu)
		})
		r.p = p
		kids = append(kids, r)
	}
	return giode.Grid(giode.Styles{Width: -1, GridColumns: "repeat(auto-fill, minmax(220, 1fr))", Gap: 12, RowGap: 12, Padding: giode.Inset{Bottom: 12}}, kids...)
}

func (s *Screen) hero() giode.Element {
	c := theme.Current()
	return giode.Row(giode.Styles{
		Width: -1, Padding: giode.UniformInset(26), Gap: 18, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		BorderRadius: 18, BorderWidth: 1, BorderColor: "#3A3170",
		BackgroundGradient: giode.Linear(120, "#2E2466", "#1C1838", "#14141C"),
		BoxShadow:          giode.BoxShadow{Y: 20, Blur: 60, Color: "#5B4BE033"},
	},
		giode.Box(giode.Styles{BorderRadius: 16, BoxShadow: theme.GlowShadow()}, widgets.BrandMark(56)),
		giode.Stack(giode.Styles{Gap: 4, FlexGrow: 1},
			giode.Text("Welcome to LiteCode", giode.Styles{Color: c.Text, FontSize: 22, FontWeight: giode.FontWeightSemiBold}),
			giode.Text("A fast, keyboard-first editor with Git built in.", giode.Styles{Color: "#B9B4DC", FontSize: theme.UISize}),
			giode.Text("Version "+version(), giode.Styles{Color: "#8A84B8", FontSize: 11})),
		s.clone, s.open,
	)
}

// version is the app version the CLI stamps in, or "dev".
func version() string {
	if giode.AppVersion != "" {
		return giode.AppVersion
	}
	return "dev"
}

// Layout draws the screen.
func (s *Screen) Layout(gtx giode.Context) giode.Dimensions {
	c := theme.Current()
	column := giode.Stack(giode.Styles{Width: -1, MaxWidth: 860, Height: -1, Gap: 18, Padding: giode.Inset{Top: 12, Bottom: 12}},
		s.hero(),
		giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Padding: giode.Inset{Top: 8}},
			giode.Text("Recent projects", giode.Styles{Color: c.Text, FontSize: 15, FontWeight: giode.FontWeightSemiBold}),
			giode.Box(giode.Styles{FlexGrow: 1}),
			giode.Box(giode.Styles{Width: 260}, s.search)),
		giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, s.list),
	)
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Background: c.Editor, BackgroundGradient: theme.Backdrop()},
		s.bar,
		giode.Row(giode.Styles{Width: -1, FlexGrow: 1, Justify: giode.JustifyCenter, Padding: giode.SymmetricInset(24, 0)}, column),
		s.cloneD,
	).Layout(gtx)
}
