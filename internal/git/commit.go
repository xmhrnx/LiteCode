package git

import (
	"context"
	"os"
	"path/filepath"
)

// CommitOptions says what to commit.
type CommitOptions struct {
	Message string
	Files   []Change // the checked files of the commit window
	Amend   bool
}

// Commit commits exactly the given files, whatever else is staged, the way the
// Changes panel does: untracked files are added first.
func (r *Repo) Commit(ctx context.Context, o CommitOptions) error {
	var untracked, paths []string
	for _, c := range o.Files {
		if c.Kind == Untracked {
			untracked = append(untracked, c.Path)
		}
		paths = append(paths, c.Path)
		if c.OldPath != "" {
			paths = append(paths, c.OldPath)
		}
	}
	if len(untracked) > 0 {
		if _, err := r.run(ctx, append([]string{"add", "--"}, untracked...)...); err != nil {
			return err
		}
	}
	args := []string{"commit", "-F", "-"}
	if o.Amend {
		args = append(args, "--amend")
	}
	if len(paths) > 0 {
		args = append(append(args, "--only", "--"), paths...)
	}
	_, err := r.runInput(ctx, []byte(o.Message), args...)
	return err
}

// LastMessage returns HEAD's message, for Amend.
func (r *Repo) LastMessage(ctx context.Context) (string, error) {
	return r.Message(ctx, "HEAD")
}

// Rollback discards local changes to files: tracked files go back to HEAD,
// added files are unstaged and deleted, untracked files are deleted.
func (r *Repo) Rollback(ctx context.Context, changes []Change) error {
	var tracked []string
	for _, c := range changes {
		switch c.Kind {
		case Untracked:
			if err := os.Remove(filepath.Join(r.Dir, c.Path)); err != nil {
				return err
			}
		case Added:
			if _, err := r.run(ctx, "rm", "-f", "--cached", "--", c.Path); err != nil {
				return err
			}
			if err := os.Remove(filepath.Join(r.Dir, c.Path)); err != nil && !os.IsNotExist(err) {
				return err
			}
		default:
			tracked = append(tracked, c.Path)
			if c.OldPath != "" {
				tracked = append(tracked, c.OldPath)
			}
		}
	}
	if len(tracked) == 0 {
		return nil
	}
	_, err := r.run(ctx, append([]string{"restore", "--source=HEAD", "--staged", "--worktree", "--"}, tracked...)...)
	return err
}
