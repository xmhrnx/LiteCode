// Package search holds the navigation popups: Search Everywhere (with its
// Files and Actions tabs), Find in Files, Recent Files and Go to Line.
package search

import (
	"path"
	"strconv"
	"strings"

	"gioui.org/io/key"
	"github.com/xmhrnx/giode"

	"litecode/internal/keymap"
	isearch "litecode/internal/search"
	"litecode/ui/core"
	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// Tab is a Search Everywhere tab.
type Tab int

const (
	TabAll Tab = iota
	TabFiles
	TabActions
)

var tabNames = []string{"All", "Files", "Actions"}

// maxFileResults bounds the file list.
const maxFileResults = 60

// result is one line of the list.
type result struct {
	key    string
	file   string // relative path, for files
	action keymap.Action
}

// SearchEverywhere is the double-Shift popup.
type SearchEverywhere struct {
	ide       *core.IDE
	popup     *widgets.Popup
	field     *giode.InputElement
	list      *giode.ScrollElement
	tabs      []*giode.PressableElement
	rows      *giode.KeyedCache[string, *widgets.ListRow]
	tab       Tab
	results   []result
	highlight int
	filesSeen int // how many project files the results were computed with
}

// NewSearchEverywhere creates the popup.
func NewSearchEverywhere(ide *core.IDE) *SearchEverywhere {
	s := &SearchEverywhere{ide: ide, rows: giode.NewKeyedCache[string, *widgets.ListRow]()}
	c := theme.Current()
	s.field = widgets.TextField("Search files and actions  (name:line jumps to a line)", giode.Styles{
		Width: -1, Background: c.Raised, BorderColor: c.Raised, Color: c.Text, FontSize: 15, Padding: giode.SymmetricInset(8, 14),
	}).OnChange(widgets.Typing(func(string) { s.update() })).
		OnKeys(s.key, key.NameUpArrow, key.NameDownArrow, key.NameReturn, key.NameEnter, key.NameTab)
	s.list = giode.Scroll(s.view, giode.Styles{ScrollBar: theme.ScrollBar(), MaxHeight: 440}).FitContent(true)
	for i := range tabNames {
		t := Tab(i)
		s.tabs = append(s.tabs, giode.Pressable(func(st giode.PressState) giode.Element {
			c := theme.Current()
			bg, fg := "", c.TextFaint
			if s.tab == t {
				bg, fg = c.AccentSoft, c.TextAccent
			} else if st.Hovered {
				bg, fg = c.Hover, c.TextMuted
			}
			return giode.Box(giode.Styles{Background: bg, BorderRadius: 6, Padding: giode.SymmetricInset(9, 4)},
				giode.Text(tabNames[t], giode.Styles{Color: fg, FontSize: theme.SmallSize, FontWeight: giode.FontWeightMedium}))
		}).Focusable(false).OnClick(func(giode.ClickEvent) { s.setTab(t) }))
	}
	s.popup = widgets.NewPopup(70, s.content)
	return s
}

// Open shows the popup on a tab.
func (s *SearchEverywhere) Open(t Tab) {
	if s.ide.Project == nil {
		return
	}
	s.tab = t
	s.field.SetText("")
	s.update()
	s.popup.Open()
	s.field.Focus()
}

// IsOpen reports whether it shows.
func (s *SearchEverywhere) IsOpen() bool { return s.popup.IsOpen() }

func (s *SearchEverywhere) setTab(t Tab) {
	s.tab = t
	s.update()
	s.field.Focus()
}

// splitLine reads "name:12" or "name:12:5".
func splitLine(q string) (name string, line, col int) {
	parts := strings.Split(q, ":")
	name = parts[0]
	if len(parts) > 1 {
		line, _ = strconv.Atoi(parts[1])
	}
	if len(parts) > 2 {
		col, _ = strconv.Atoi(parts[2])
	}
	return name, line, col
}

func (s *SearchEverywhere) update() {
	q, _, _ := splitLine(strings.TrimSpace(s.field.Text()))
	files := s.ide.Project.Files(s.ide)
	s.filesSeen = len(files)
	s.results = s.results[:0]
	s.highlight = 0
	if s.tab != TabActions && q != "" {
		for _, m := range isearch.MatchFiles(files, q, maxFileResults) {
			s.results = append(s.results, result{key: "f:" + m.Path, file: m.Path})
		}
	}
	if s.tab != TabFiles {
		lq := strings.ToLower(q)
		for _, a := range keymap.Actions() {
			if !s.ide.Has(a) {
				continue
			}
			if (q == "" && s.tab == TabActions) || (q != "" && strings.Contains(strings.ToLower(keymap.Names[a]), lq)) {
				s.results = append(s.results, result{key: "a:" + string(a), action: a})
			}
		}
	}
}

func (s *SearchEverywhere) key(e key.Event) {
	if e.State != key.Press {
		return
	}
	n := len(s.results)
	switch e.Name {
	case key.NameTab:
		step := 1
		if e.Modifiers.Contain(key.ModShift) {
			step = len(tabNames) - 1
		}
		s.setTab(Tab((int(s.tab) + step) % len(tabNames)))
	case key.NameUpArrow:
		if n > 0 {
			s.highlight = (s.highlight - 1 + n) % n
		}
	case key.NameDownArrow:
		if n > 0 {
			s.highlight = (s.highlight + 1) % n
		}
	default:
		if n > 0 {
			s.activate(s.results[s.highlight])
		}
	}
}

func (s *SearchEverywhere) activate(r result) {
	s.popup.Close()
	if r.file != "" {
		_, line, col := splitLine(strings.TrimSpace(s.field.Text()))
		s.ide.OpenFile(s.ide.Project.Abs(r.file), line, col)
		return
	}
	s.ide.Run(r.action)
}

func (s *SearchEverywhere) row(i int, r result) giode.Element {
	k := r.key
	row := s.rows.Get(k, func() *widgets.ListRow {
		return widgets.NewListRow(func(giode.ClickEvent) {
			for _, cur := range s.results {
				if cur.key == k {
					s.activate(cur)
					return
				}
			}
		})
	})
	c := theme.Current()
	row.Selected = i == s.highlight
	if r.file != "" {
		row.Icon = theme.FileIcon(path.Base(r.file), false)
		row.Label, row.Detail, row.Hint = path.Base(r.file), path.Dir(r.file), ""
		if row.Detail == "." {
			row.Detail = ""
		}
	} else {
		row.Icon = theme.Icon("action", c.TextMuted)
		row.Label, row.Detail, row.Hint = keymap.Names[r.action], "", s.ide.Shortcut(r.action)
	}
	return row
}

func (s *SearchEverywhere) view() giode.Element {
	if len(s.ide.Project.Files(s.ide)) != s.filesSeen {
		s.update()
	}
	c := theme.Current()
	if len(s.results) == 0 {
		msg := "Nothing found"
		if strings.TrimSpace(s.field.Text()) == "" {
			msg = "Type to search files and actions"
		}
		return giode.Box(giode.Styles{Width: -1, Padding: giode.UniformInset(16), Justify: giode.JustifyCenter},
			giode.Text(msg, giode.Styles{Color: c.TextFaint, FontSize: theme.UISize}))
	}
	var kids []giode.Element
	section := ""
	for i, r := range s.results {
		sec := "Files"
		if r.file == "" {
			sec = "Actions"
		}
		if s.tab == TabAll && sec != section {
			section = sec
			kids = append(kids, widgets.SectionTitle(sec))
		}
		kids = append(kids, s.row(i, r))
	}
	return giode.Stack(giode.Styles{Width: -1, Padding: giode.Inset{Top: 4, Bottom: 6}}, kids...)
}

func (s *SearchEverywhere) content() giode.Element {
	c := theme.Current()
	tabs := []giode.Element{}
	for _, t := range s.tabs {
		tabs = append(tabs, t)
	}
	return giode.Stack(theme.PopupPanel(680),
		giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Padding: giode.Inset{Left: 16, Right: 12}, Wrap: giode.WrapNoWrap},
			theme.IconSized("search", c.TextMuted, 18),
			giode.Box(giode.Styles{FlexGrow: 1}, s.field),
			giode.Row(giode.Styles{Gap: 2, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, tabs...)),
		giode.Box(giode.Styles{Width: -1, Height: 1, Background: c.PopupLine}),
		s.list,
		giode.Box(giode.Styles{Width: -1, Height: 1, Background: c.PopupLine}),
		giode.Row(giode.Styles{Width: -1, Gap: 16, Align: giode.AlignCenter, Padding: giode.SymmetricInset(16, 9), Background: c.Card},
			footerHint("↑↓", "navigate"), footerHint("↵", "open"), footerHint("Tab", "switch scope"), footerHint("Esc", "close")),
	)
}

// Element is the popup layer; it must be a direct child of the window root.
func (s *SearchEverywhere) Element() giode.Element {
	return s.popup.Element()
}

// footerHint is one "key action" pair of the popup's footer.
func footerHint(key, action string) giode.Element {
	c := theme.Current()
	return giode.Row(giode.Styles{Gap: 6, Align: giode.AlignCenter},
		widgets.Keycaps(key),
		giode.Text(action, giode.Styles{Color: c.TextFaint, FontSize: 11}))
}
