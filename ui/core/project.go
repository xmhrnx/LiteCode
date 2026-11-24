package core

import (
	"context"
	"path/filepath"

	"litecode/internal/git"
	"litecode/internal/lsp"
	"litecode/internal/project"
)

// Project is the open project.
type Project struct {
	Root string
	Name string
	Repo *git.Repo // nil outside a git repository
	Git  *GitState
	// Servers are the project's language servers.
	Servers *lsp.Manager

	files    []string
	filesErr error
	loading  bool
}

// OpenProject reads what the IDE needs of a project folder.
func OpenProject(ctx context.Context, root string) *Project {
	p := &Project{Root: filepath.Clean(root), Name: filepath.Base(root)}
	p.Servers = lsp.NewManager(p.Root)
	if repo, err := git.Open(ctx, root); err == nil {
		p.Repo = repo
		p.Git = &GitState{repo: repo}
	}
	return p
}

// Files returns the project's file list, loading it in the background the
// first time (and after Invalidate); until then it's empty.
func (p *Project) Files(ide *IDE) []string {
	if p.files == nil && !p.loading {
		p.loading = true
		go func() {
			files, err := project.Files(context.Background(), p.Root)
			ide.App.Post(func() {
				p.files, p.filesErr, p.loading = files, err, false
				if p.files == nil {
					p.files = []string{}
				}
				ide.Error(err)
			})
		}()
	}
	return p.files
}

// InvalidateFiles makes the next Files call reload the list.
func (p *Project) InvalidateFiles() {
	p.files = nil
}

// Rel returns path relative to the project root, slash separated.
func (p *Project) Rel(path string) string {
	rel, err := filepath.Rel(p.Root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}

// Abs returns the absolute path of a project-relative path.
func (p *Project) Abs(rel string) string {
	return filepath.Join(p.Root, filepath.FromSlash(rel))
}
