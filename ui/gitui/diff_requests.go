package gitui

import (
	"context"
	"os"
	"path/filepath"

	"litecode/internal/git"
	"litecode/ui/core"
)

// workingTreeDiff compares a changed file's HEAD version with the file on disk.
func workingTreeDiff(ide *core.IDE, c git.Change) core.DiffRequest {
	repo := ide.Project.Git.Repo()
	old := c.Path
	if c.OldPath != "" {
		old = c.OldPath
	}
	return core.DiffRequest{
		Title: filepath.Base(c.Path), Path: c.Path,
		LeftLabel: "HEAD", RightLabel: "Working tree",
		Left: func(ctx context.Context) (string, error) {
			if c.Kind == git.Untracked || c.Kind == git.Added {
				return "", nil
			}
			return repo.FileAt(ctx, "HEAD", old)
		},
		Right: func(context.Context) (string, error) {
			if c.Kind == git.Deleted {
				return "", nil
			}
			data, err := os.ReadFile(filepath.Join(repo.Dir, filepath.FromSlash(c.Path)))
			return string(data), err
		},
	}
}

// commitDiff compares a file as a commit changed it with its first parent.
func commitDiff(ide *core.IDE, commit git.Commit, c git.Change) core.DiffRequest {
	repo := ide.Project.Git.Repo()
	old := c.Path
	if c.OldPath != "" {
		old = c.OldPath
	}
	return core.DiffRequest{
		Title: filepath.Base(c.Path) + " (" + commit.Short + ")", Path: c.Path,
		LeftLabel: commit.Short + "^", RightLabel: commit.Short,
		Left: func(ctx context.Context) (string, error) {
			if len(commit.Parents) == 0 {
				return "", nil
			}
			return repo.FileAt(ctx, commit.Hash+"^", old)
		},
		Right: func(ctx context.Context) (string, error) {
			return repo.FileAt(ctx, commit.Hash, c.Path)
		},
	}
}
