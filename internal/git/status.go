package git

import (
	"context"
	"sort"
	"strings"
)

// ChangeKind is how a file differs from HEAD.
type ChangeKind int

const (
	Modified ChangeKind = iota
	Added
	Deleted
	Renamed
	Untracked
	Conflicted
)

// Change is one changed file in the working tree.
type Change struct {
	Path     string // relative to the repository root, slash separated
	OldPath  string // for renames
	Kind     ChangeKind
	Staged   bool
	Unstaged bool
}

// String names the kind the way the IDE labels it.
func (k ChangeKind) String() string {
	return [...]string{"Modified", "Added", "Deleted", "Renamed", "Untracked", "Conflict"}[k]
}

// Status lists changed and untracked files, sorted by path.
func (r *Repo) Status(ctx context.Context) ([]Change, error) {
	out, err := r.run(ctx, "status", "--porcelain=v1", "-z", "-uall")
	if err != nil {
		return nil, err
	}
	return parseStatus(out), nil
}

func parseStatus(out string) []Change {
	var changes []Change
	fields := strings.Split(out, "\x00")
	for i := 0; i < len(fields); i++ {
		f := fields[i]
		if len(f) < 4 {
			continue
		}
		x, y, path := f[0], f[1], f[3:]
		c := Change{Path: path, Staged: x != ' ' && x != '?', Unstaged: y != ' '}
		switch {
		case x == '?':
			c.Kind = Untracked
		case x == 'U' || y == 'U' || (x == 'A' && y == 'A') || (x == 'D' && y == 'D'):
			c.Kind = Conflicted
		case x == 'R' || y == 'R':
			c.Kind = Renamed
			if i+1 < len(fields) {
				c.OldPath = fields[i+1]
				i++
			}
		case x == 'A':
			c.Kind = Added
		case x == 'D' || y == 'D':
			c.Kind = Deleted
		default:
			c.Kind = Modified
		}
		changes = append(changes, c)
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].Path < changes[j].Path })
	return changes
}

// UntrackedDirs lists the folders git reports as wholly unversioned, with
// their trailing slash dropped.
func (r *Repo) UntrackedDirs(ctx context.Context) ([]string, error) {
	out, err := r.run(ctx, "status", "--porcelain=v1", "-z", "-unormal")
	if err != nil {
		return nil, err
	}
	var dirs []string
	for _, f := range strings.Split(out, "\x00") {
		if strings.HasPrefix(f, "?? ") && strings.HasSuffix(f, "/") {
			dirs = append(dirs, strings.TrimSuffix(f[3:], "/"))
		}
	}
	return dirs, nil
}
