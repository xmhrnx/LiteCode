// Package projectview is the Project tool window: the project's files as a
// tree, loaded a folder at a time as folders are expanded.
package projectview

import (
	"path/filepath"

	"github.com/xmhrnx/giode"

	"litecode/internal/project"
)

// fileNode is a file or folder in the tree.
type fileNode struct {
	path     string
	name     string
	dir      bool
	loaded   bool
	children []giode.TreeNode
}

func newFileNode(e project.Entry) *fileNode {
	return &fileNode{path: e.Path, name: e.Name, dir: e.IsDir}
}

func (n *fileNode) Key() any                   { return n.path }
func (n *fileNode) Children() []giode.TreeNode { return n.children }
func (n *fileNode) HasChildren() bool          { return n.dir && (!n.loaded || len(n.children) > 0) }
func (n *fileNode) Label() string              { return n.name }

// load reads the folder's entries now.
func (n *fileNode) load() error {
	entries, err := project.ReadDir(n.path)
	if err != nil {
		return err
	}
	n.setEntries(entries)
	return nil
}

// setEntries replaces the children, keeping already loaded subfolders.
func (n *fileNode) setEntries(entries []project.Entry) {
	old := map[string]*fileNode{}
	for _, c := range n.children {
		fn := c.(*fileNode)
		old[fn.path] = fn
	}
	n.children = n.children[:0]
	for _, e := range entries {
		if prev, ok := old[e.Path]; ok && prev.dir == e.IsDir {
			n.children = append(n.children, prev)
			continue
		}
		n.children = append(n.children, newFileNode(e))
	}
	n.loaded = true
}

// find returns the loaded node at path below n.
func (n *fileNode) find(path string) *fileNode {
	if n.path == path {
		return n
	}
	for _, c := range n.children {
		fn := c.(*fileNode)
		if fn.dir && isUnder(path, fn.path) || fn.path == path {
			if found := fn.find(path); found != nil {
				return found
			}
		}
	}
	return nil
}

func isUnder(path, dir string) bool {
	rel, err := filepath.Rel(dir, path)
	return err == nil && rel != "." && !filepath.IsAbs(rel) && rel[:min(2, len(rel))] != ".."
}
