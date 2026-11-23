package search

import (
	"path/filepath"
	"strings"

	"gioui.org/io/key"
	"github.com/xmhrnx/giode"

	"litecode/ui/core"
	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// RecentFiles is the Ctrl+E popup: recently opened files, filtered as you type.
type RecentFiles struct {
	ide       *core.IDE
	popup     *widgets.Popup
	field     *giode.InputElement
	rows      *giode.KeyedCache[string, *widgets.ListRow]
	recent    func() []string
	shown     []string
	highlight int
}

// NewRecentFiles creates the popup; recent lists files, most recent first.
func NewRecentFiles(ide *core.IDE, recent func() []string) *RecentFiles {
	r := &RecentFiles{ide: ide, recent: recent, rows: giode.NewKeyedCache[string, *widgets.ListRow]()}
	c := theme.Current()
	r.field = widgets.TextField("Recent Files", giode.Styles{
		Width: -1, Background: c.Raised, BorderColor: c.Raised, Color: c.Text, FontSize: theme.UISize, Padding: giode.SymmetricInset(12, 10),
	}).OnChange(widgets.Typing(func(string) { r.filter() })).
		OnKeys(r.key, key.NameUpArrow, key.NameDownArrow, key.NameReturn, key.NameEnter)
	r.popup = widgets.NewPopup(90, r.content)
	return r
}

// Open shows the popup with the second most recent file highlighted, so
// Ctrl+E, Enter goes back to the previous file as in the IDE.
func (r *RecentFiles) Open() {
	r.field.SetText("")
	r.filter()
	r.highlight = min(1, max(len(r.shown)-1, 0))
	r.popup.Open()
	r.field.Focus()
}

func (r *RecentFiles) filter() {
	q := strings.ToLower(r.field.Text())
	r.shown = r.shown[:0]
	for _, p := range r.recent() {
		if q == "" || strings.Contains(strings.ToLower(filepath.Base(p)), q) {
			r.shown = append(r.shown, p)
		}
	}
	r.highlight = 0
}

func (r *RecentFiles) key(e key.Event) {
	if e.State != key.Press || len(r.shown) == 0 {
		return
	}
	switch e.Name {
	case key.NameUpArrow:
		r.highlight = (r.highlight - 1 + len(r.shown)) % len(r.shown)
	case key.NameDownArrow:
		r.highlight = (r.highlight + 1) % len(r.shown)
	default:
		r.open(r.shown[r.highlight])
	}
}

func (r *RecentFiles) open(path string) {
	r.popup.Close()
	r.ide.OpenFile(path, 0, 0)
}

func (r *RecentFiles) content() giode.Element {
	var kids []giode.Element
	for i, p := range r.shown {
		path := p
		row := r.rows.Get(path, func() *widgets.ListRow {
			return widgets.NewListRow(func(giode.ClickEvent) { r.open(path) })
		})
		row.Icon = theme.FileIcon(filepath.Base(p), false)
		row.Label, row.Selected = filepath.Base(p), i == r.highlight
		row.Detail = ""
		if r.ide.Project != nil {
			row.Detail = filepath.Dir(r.ide.Project.Rel(p))
		}
		kids = append(kids, row)
	}
	if len(kids) == 0 {
		kids = append(kids, giode.Box(giode.Styles{Padding: giode.UniformInset(14)},
			giode.Text("No recent files", giode.Styles{Color: theme.Current().TextFaint, FontSize: theme.UISize})))
	}
	return giode.Stack(theme.PopupPanel(520),
		r.field, theme.Separator(),
		giode.Stack(giode.Styles{Width: -1, Padding: giode.Inset{Top: 4, Bottom: 6}}, kids...))
}

// Element is the popup layer; it must be a direct child of the window root.
func (r *RecentFiles) Element() giode.Element {
	return r.popup.Element()
}
