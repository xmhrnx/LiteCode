package git

import (
	"context"
	"strings"
)

// Message returns a commit's full message.
func (r *Repo) Message(ctx context.Context, hash string) (string, error) {
	out, err := r.run(ctx, "show", "-s", "--format=%B", hash)
	return strings.TrimSpace(out), err
}

// CommitChanges lists the files a commit changed relative to its first parent.
func (r *Repo) CommitChanges(ctx context.Context, hash string) ([]Change, error) {
	out, err := r.run(ctx, "diff-tree", "--no-commit-id", "-r", "-M", "--root", "--name-status", "-z", hash)
	if err != nil {
		return nil, err
	}
	var changes []Change
	f := strings.Split(out, "\x00")
	for i := 0; i+1 < len(f); i++ {
		code := f[i]
		if code == "" {
			continue
		}
		c := Change{Path: f[i+1]}
		i++
		switch code[0] {
		case 'A':
			c.Kind = Added
		case 'D':
			c.Kind = Deleted
		case 'R', 'C':
			c.Kind = Renamed
			c.OldPath = c.Path
			if i+1 < len(f) {
				c.Path = f[i+1]
				i++
			}
		default:
			c.Kind = Modified
		}
		changes = append(changes, c)
	}
	return changes, nil
}

// FileAt returns a file's content at a revision ("HEAD", a hash, "hash^").
// A file missing at that revision is returned empty, without an error.
func (r *Repo) FileAt(ctx context.Context, rev, path string) (string, error) {
	if _, err := r.run(ctx, "cat-file", "-e", rev+":"+path); err != nil {
		return "", nil
	}
	return r.run(ctx, "show", rev+":"+path)
}
