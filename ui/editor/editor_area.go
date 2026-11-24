package editor

import (
	"errors"
	"path/filepath"
	"slices"

	"github.com/xmhrnx/giode"

	"litecode/internal/workspace"
	"litecode/ui/core"
	"litecode/ui/theme"
)

// maxRecent is how many files Recent Files remembers.
const maxRecent = 50

// EditorArea is the tabs of open files and diffs, and the selected one.
type EditorArea struct {
	ide    *core.IDE
	tabs   *giode.DocTabsElement
	views  map[string]*CodeView     // open files by path
	extras map[string]giode.Element // other tabs (diffs) by ID
	recent []string                 // tab IDs, most recently selected first
	empty  *EmptyEditor
	// OnSelectionChange runs when another tab is selected (the project tree follows it).
	OnSelectionChange func(path string)
	// OnTabsChange runs after files are opened or closed, to save the session.
	OnTabsChange func()
}

// NewEditorArea creates an empty editor area.
func NewEditorArea(ide *core.IDE) *EditorArea {
	a := &EditorArea{ide: ide, views: map[string]*CodeView{}, extras: map[string]giode.Element{}, empty: NewEmptyEditor(ide)}
	c := theme.Current()
	a.tabs = giode.DocTabs(giode.Styles{Background: c.Panel, Color: c.Text, BorderColor: c.Separator, FontSize: theme.UISize, Height: theme.TabStripHeight}).
		Indicator(giode.Color(c.TabActive)).
		OnPin(func(string, bool) { a.tabsChanged() }).
		OnSelect(a.selected).
		OnClose(a.Close).
		OnCloseTabs(func(ids []string) {
			for _, id := range ids {
				a.Close(id)
			}
		}).
		TabMenu(func(id string, m *giode.MenuElement) {
			if _, ok := a.views[id]; !ok {
				return
			}
			m.Separator().
				Item("Copy Path", func() { ide.App.SetClipboard(id) }).
				Item("Copy Relative Path", func() { ide.App.SetClipboard(ide.Project.Rel(id)) })
		})
	return a
}

func (a *EditorArea) selected(id string) {
	a.touch(id)
	a.tabsChanged()
	if v := a.views[id]; v != nil {
		v.Focus()
	}
	if a.OnSelectionChange != nil {
		a.OnSelectionChange(id)
	}
}

func (a *EditorArea) touch(id string) {
	a.recent = slices.DeleteFunc(a.recent, func(s string) bool { return s == id })
	a.recent = slices.Insert(a.recent, 0, id)
	if len(a.recent) > maxRecent {
		a.recent = a.recent[:maxRecent]
	}
}

// Open opens a file (or selects its tab) and puts the caret at a 1-based
// line and column; line 0 keeps the caret where it was.
func (a *EditorArea) Open(path string, line, col int) {
	path = filepath.Clean(path)
	v := a.views[path]
	if v == nil {
		doc, err := workspace.Open(path)
		if err != nil {
			if errors.Is(err, workspace.ErrBinary) || errors.Is(err, workspace.ErrTooLarge) {
				a.ide.Info(filepath.Base(path) + ": " + err.Error())
				return
			}
			a.ide.Error(err)
			return
		}
		v = NewCodeView(a.ide, doc, func() { a.tabs.SetModified(path, a.views[path].Modified()) })
		a.views[path] = v
		a.tabs.Add(giode.DocTab{ID: path, Title: v.Title(), Tooltip: path, Icon: theme.FileIcon(v.Title(), false)})
	}
	a.tabs.Select(path)
	a.touch(path)
	a.tabsChanged()
	if line > 0 {
		v.GoTo(line, max(col, 1))
	} else {
		v.Focus()
	}
	if a.OnSelectionChange != nil {
		a.OnSelectionChange(path)
	}
}

// OpenTab opens a tab that isn't a file (a diff) under id.
func (a *EditorArea) OpenTab(id, title string, el giode.Element) {
	if _, ok := a.extras[id]; !ok {
		a.tabs.Add(giode.DocTab{ID: id, Title: title, Tooltip: title, Icon: theme.Icon("diff", theme.Current().TextAccent)})
	}
	a.extras[id] = el
	a.tabs.Select(id)
	a.touch(id)
}

// Close closes a tab, saving the file first (no prompt).
func (a *EditorArea) Close(id string) {
	if v := a.views[id]; v != nil && v.Modified() {
		v.Save(false, func() { a.remove(id) })
		return
	}
	a.remove(id)
}

func (a *EditorArea) remove(id string) {
	if v := a.views[id]; v != nil {
		v.Dispose()
	}
	delete(a.views, id)
	delete(a.extras, id)
	a.recent = slices.DeleteFunc(a.recent, func(s string) bool { return s == id })
	a.tabs.Remove(id)
	a.tabsChanged()
}

func (a *EditorArea) tabsChanged() {
	if a.OnTabsChange != nil {
		a.OnTabsChange()
	}
}

// CloseAll closes every tab.
func (a *EditorArea) CloseAll() {
	for _, t := range a.tabs.Tabs() {
		a.Close(t.ID)
	}
}

// Current returns the selected file's view, or nil.
func (a *EditorArea) Current() *CodeView {
	return a.views[a.tabs.Selected()]
}

// CurrentExtra returns the selected tab's element when it isn't a file (a diff).
func (a *EditorArea) CurrentExtra() giode.Element {
	return a.extras[a.tabs.Selected()]
}

// SelectNext moves to the next (1) or previous (-1) tab.
func (a *EditorArea) SelectNext(step int) {
	if a.tabs.Len() > 0 {
		a.tabs.SelectNext(step)
	}
}

// CloseCurrent closes the selected tab.
func (a *EditorArea) CloseCurrent() {
	if id := a.tabs.Selected(); id != "" {
		a.Close(id)
	}
}

// Recent lists open and recently closed files, most recent first.
func (a *EditorArea) Recent() []string {
	var out []string
	for _, id := range a.recent {
		if _, isExtra := a.extras[id]; !isExtra {
			out = append(out, id)
		}
	}
	return out
}

// OpenFiles lists the open files in tab order.
func (a *EditorArea) OpenFiles() []string {
	var out []string
	for _, t := range a.tabs.Tabs() {
		if _, ok := a.views[t.ID]; ok {
			out = append(out, t.ID)
		}
	}
	return out
}

// SaveAll saves every modified file; formatting runs first when asked and
// the user's settings enable it. done runs once all are written.
func (a *EditorArea) SaveAll(withFormat bool, done func()) {
	pending := 0
	for _, v := range a.views {
		if v.Modified() {
			pending++
		}
	}
	if pending == 0 {
		if done != nil {
			done()
		}
		return
	}
	for _, v := range a.views {
		if !v.Modified() {
			continue
		}
		v.Save(withFormat && a.ide.Settings.FormatOnSave, func() {
			pending--
			if pending == 0 && done != nil {
				done()
			}
		})
	}
}

// Session returns what to reopen next time.
func (a *EditorArea) Session() workspace.Session {
	s := workspace.Session{OpenFiles: a.OpenFiles()}
	for _, t := range a.tabs.Tabs() {
		if _, ok := a.views[t.ID]; ok && t.Pinned {
			s.Pinned = append(s.Pinned, t.ID)
		}
	}
	if _, ok := a.views[a.tabs.Selected()]; ok {
		s.Selected = a.tabs.Selected()
	}
	return s
}

// Restore reopens a session's files, skipping ones that are gone.
func (a *EditorArea) Restore(s workspace.Session) {
	for _, p := range s.OpenFiles {
		doc, err := workspace.Open(p)
		if err != nil {
			continue
		}
		path := p
		v := NewCodeView(a.ide, doc, func() { a.tabs.SetModified(path, a.views[path].Modified()) })
		a.views[path] = v
		a.tabs.Add(giode.DocTab{ID: path, Title: v.Title(), Tooltip: path, Icon: theme.FileIcon(v.Title(), false), Pinned: slices.Contains(s.Pinned, path)})
		a.recent = append(a.recent, path)
	}
	sel := s.Selected
	if a.views[sel] == nil && len(a.recent) > 0 {
		sel = a.recent[0]
	}
	if a.views[sel] != nil {
		a.tabs.Select(sel)
		a.touch(sel)
	}
}

func (a *EditorArea) body() giode.Element {
	id := a.tabs.Selected()
	if v := a.views[id]; v != nil {
		return v
	}
	if el := a.extras[id]; el != nil {
		return el
	}
	return a.empty
}

// Layout draws the tabs and the selected tab's content.
func (a *EditorArea) Layout(gtx giode.Context) giode.Dimensions {
	c := theme.Current()
	if a.tabs.Len() == 0 {
		return a.empty.Layout(gtx)
	}
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Background: c.Editor},
		a.tabs,
		giode.Box(giode.Styles{Width: -1, FlexGrow: 1, Overflow: giode.OverflowHidden}, a.body()),
	).Layout(gtx)
}
