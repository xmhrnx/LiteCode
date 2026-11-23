package frame

import (
	"github.com/xmhrnx/giode"

	"litecode/internal/keymap"
	"litecode/internal/project"
	"litecode/ui/core"
	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// ProjectWidget is the project's name in the title bar; it opens the recent
// projects menu, with Open and Close Project.
type ProjectWidget struct {
	ide    *core.IDE
	button *giode.PressableElement
	menu   *giode.MenuElement
}

// NewProjectWidget creates the widget.
func NewProjectWidget(ide *core.IDE) *ProjectWidget {
	w := &ProjectWidget{ide: ide, menu: theme.NewMenu()}
	w.button = widgetButton(
		func() giode.Element { return widgets.ProjectBadge(ide.Project.Name, 20) },
		func() string { return ide.Project.Name },
		func(e giode.ClickEvent) {
			w.rebuild()
			w.menu.OpenAtClick(e)
		})
	return w
}

// rebuild fills the menu from the current recent projects list.
func (w *ProjectWidget) rebuild() {
	w.menu.Clear()
	w.menu.Item("Open…", func() { w.ide.Run(keymap.OpenProject) }).Shortcut(w.ide.Keymap.First(keymap.OpenProject))
	w.menu.Item("Close Project", func() { w.ide.CloseProject() })
	w.menu.Separator()
	n := 0
	for _, p := range w.ide.Recent.Projects {
		if p.Path == w.ide.Project.Root {
			continue
		}
		path := p.Path
		w.menu.Item(p.Name+"   "+project.ShortPath(p.Path), func() { w.ide.OpenProject(path) })
		if n++; n == 12 {
			break
		}
	}
	if n == 0 {
		w.menu.DisabledItem("No recent projects")
	}
}

// Layout draws the widget and its menu.
func (w *ProjectWidget) Layout(gtx giode.Context) giode.Dimensions {
	return giode.Box(giode.Styles{}, w.button, w.menu).Layout(gtx)
}
