package projectview

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/xmhrnx/giode"

	"litecode/internal/git"
	"litecode/internal/layout"
	"litecode/internal/project"
	"litecode/ui/core"
	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// WindowID is the Project tool window's ID.
const WindowID layout.WindowID = "files"

// ProjectView is the Project tool window.
type ProjectView struct {
	ide     *core.IDE
	root    *fileNode
	tree    *giode.TreeViewElement
	actions []giode.Element
	prompt  *widgets.PromptDialog
	confirm *widgets.ConfirmDialog

	statusVersion int
	status        map[string]git.ChangeKind // by absolute path, folders included
}

// New creates the Project tool window for the open project; current returns
// the selected editor file, which Select Opened File reveals.
func New(ide *core.IDE, current func() string) *ProjectView {
	p := &ProjectView{ide: ide, prompt: widgets.NewPromptDialog(nil), confirm: widgets.NewConfirmDialog(), statusVersion: -1}
	p.root = &fileNode{path: ide.Project.Root, name: ide.Project.Name, dir: true}
	ide.Error(p.root.load())
	c := theme.Current()
	p.tree = giode.TreeView(func() []giode.TreeNode { return []giode.TreeNode{p.root} }, p.row,
		giode.Styles{ScrollBar: theme.ScrollBar(), Height: theme.RowHeight, Background: c.Panel, Color: c.TextMuted, BorderColor: c.Selection}).
		OnExpand(p.expand).
		OnActivate(func(n giode.TreeNode) {
			fn := n.(*fileNode)
			if fn.dir {
				p.tree.Toggle(fn.path)
				p.expand(fn)
				return
			}
			ide.OpenFile(fn.path, 0, 0)
		})
	p.tree.Expand(p.root.path, true)
	p.actions = []giode.Element{
		widgets.NewToolButton("locate", "Reveal Current File", func() {
			if path := current(); path != "" {
				p.Reveal(path)
			}
		}),
		widgets.NewToolButton("collapse", "Collapse All", p.collapseAll),
		widgets.NewToolButton("refresh", "Refresh", p.Refresh),
	}
	return p
}

func (p *ProjectView) ID() layout.WindowID            { return WindowID }
func (p *ProjectView) Title() string                  { return "Files" }
func (p *ProjectView) Icon() string                   { return "panel-project" }
func (p *ProjectView) HeaderActions() []giode.Element { return p.actions }

func (p *ProjectView) expand(n giode.TreeNode) {
	fn := n.(*fileNode)
	if fn.loaded {
		return
	}
	go func() {
		entries, err := project.ReadDir(fn.path)
		p.ide.App.Post(func() {
			if err != nil {
				p.ide.Error(err)
				fn.loaded = true
				return
			}
			fn.setEntries(entries)
		})
	}()
}

// Reveal opens the folders down to path and selects it.
func (p *ProjectView) Reveal(path string) {
	if !isUnder(path, p.root.path) {
		return
	}
	rel, _ := filepath.Rel(p.root.path, path)
	node := p.root
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		if !node.loaded {
			if err := node.load(); err != nil {
				p.ide.Error(err)
				return
			}
		}
		p.tree.Expand(node.path, true)
		next := node.find(filepath.Join(node.path, part))
		if next == nil {
			return
		}
		node = next
	}
	p.tree.Reveal(path)
}

func (p *ProjectView) collapseAll() {
	var walk func(n *fileNode)
	walk = func(n *fileNode) {
		for _, c := range n.children {
			if fn := c.(*fileNode); fn.dir {
				p.tree.Expand(fn.path, false)
				walk(fn)
			}
		}
	}
	walk(p.root)
}

// Refresh reloads every loaded folder from disk.
func (p *ProjectView) Refresh() {
	var walk func(n *fileNode)
	walk = func(n *fileNode) {
		if !n.loaded {
			return
		}
		p.ide.Error(n.load())
		for _, c := range n.children {
			if fn := c.(*fileNode); fn.dir {
				walk(fn)
			}
		}
	}
	walk(p.root)
	p.ide.Project.InvalidateFiles()
}

// statusOf returns a path's VCS status, recomputing the map when git
// changed. Folders are colored only when nothing in them is versioned, as the
// IDE does by default.
func (p *ProjectView) statusOf(path string) (git.ChangeKind, bool) {
	g := p.ide.Project.Git
	if g == nil {
		return 0, false
	}
	if g.Version != p.statusVersion {
		p.statusVersion = g.Version
		p.status = map[string]git.ChangeKind{}
		root := g.Repo().Dir
		for _, c := range g.Changes {
			p.status[filepath.Join(root, filepath.FromSlash(c.Path))] = c.Kind
		}
		for _, d := range g.UntrackedDirs {
			p.status[filepath.Join(root, filepath.FromSlash(d))] = git.Untracked
		}
	}
	k, ok := p.status[path]
	return k, ok
}

func (p *ProjectView) row(n giode.TreeNode, s giode.TreeRowState) giode.Element {
	fn := n.(*fileNode)
	c := theme.Current()
	color := c.Text
	if k, ok := p.statusOf(fn.path); ok {
		color = theme.StatusColor(k)
	}
	kids := []giode.Element{theme.FileIcon(fn.name, fn.dir), giode.Text(fn.name, theme.Label(color))}
	if fn == p.root {
		kids[1] = giode.Text(fn.name, giode.Styles{Color: color, FontSize: theme.UISize, FontWeight: giode.FontWeightSemiBold})
		kids = append(kids, giode.Text(project.ShortPath(fn.path), theme.Label(c.TextFaint)))
	}
	row := giode.Row(giode.Styles{Gap: 6, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Height: -1}, kids...)
	row.Context(giode.Keyed("project-menu", fn.path, func() *giode.MenuElement { return p.menu(fn) }))
	return row
}

func (p *ProjectView) menu(fn *fileNode) *giode.MenuElement {
	dir := fn.path
	if !fn.dir {
		dir = filepath.Dir(fn.path)
	}
	m := theme.NewMenu().
		Item("New File…", func() { p.create(dir, false) }).
		Item("New Directory…", func() { p.create(dir, true) }).
		Separator().
		Item("Copy Path", func() { p.ide.App.SetClipboard(fn.path) }).
		Item("Copy Relative Path", func() { p.ide.App.SetClipboard(p.ide.Project.Rel(fn.path)) }).
		Item("Open in File Manager", func() {
			go func() {
				err := p.ide.App.OpenFolder(context.Background(), dir)
				p.ide.App.Post(func() { p.ide.Error(err) })
			}()
		})
	if fn != p.root {
		m.Separator().
			Item("Rename…", func() { p.rename(fn) }).
			Item("Delete…", func() { p.remove(fn) })
	}
	return m
}

func (p *ProjectView) create(dir string, folder bool) {
	title := "New File"
	if folder {
		title = "New Directory"
	}
	p.prompt.Ask(title, "", func(name string) {
		path := filepath.Join(dir, name)
		var err error
		if folder {
			err = os.MkdirAll(path, 0o755)
		} else {
			if err = os.MkdirAll(filepath.Dir(path), 0o755); err == nil {
				var f *os.File
				if f, err = os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644); err == nil {
					err = f.Close()
				}
			}
		}
		if err != nil {
			p.ide.Error(err)
			return
		}
		p.Refresh()
		p.Reveal(path)
		if !folder {
			p.ide.OpenFile(path, 0, 0)
		}
		p.refreshGit()
	})
}

func (p *ProjectView) rename(fn *fileNode) {
	p.prompt.Ask("Rename", fn.name, func(name string) {
		if err := os.Rename(fn.path, filepath.Join(filepath.Dir(fn.path), name)); err != nil {
			p.ide.Error(err)
			return
		}
		p.Refresh()
		p.refreshGit()
	})
}

func (p *ProjectView) remove(fn *fileNode) {
	p.confirm.Ask("Delete", "Delete "+p.ide.Project.Rel(fn.path)+"?", "Delete", func() {
		if err := os.RemoveAll(fn.path); err != nil {
			p.ide.Error(err)
			return
		}
		p.Refresh()
		p.refreshGit()
	})
}

func (p *ProjectView) refreshGit() {
	if g := p.ide.Project.Git; g != nil {
		g.Refresh(p.ide, false)
	}
}

// View draws the tree and the dialogs it opens.
func (p *ProjectView) View() giode.Element {
	return giode.Box(giode.Styles{Width: -1, Height: -1}, p.tree, p.prompt, p.confirm)
}
