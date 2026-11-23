package git

import (
	"context"
	"strings"
)

// Branch is a local or remote-tracking branch.
type Branch struct {
	Name     string // "main" or "origin/main"
	Remote   bool
	Current  bool
	Upstream string
	Ahead    int
	Behind   int
}

// CurrentBranch returns the checked-out branch, or the short hash when detached.
func (r *Repo) CurrentBranch(ctx context.Context) (string, error) {
	out, err := r.run(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	name := strings.TrimSpace(out)
	if name == "HEAD" {
		hash, err := r.run(ctx, "rev-parse", "--short", "HEAD")
		return strings.TrimSpace(hash), err
	}
	return name, nil
}

// Branches lists local branches, then remote ones, each in name order.
func (r *Repo) Branches(ctx context.Context) ([]Branch, error) {
	out, err := r.run(ctx, "for-each-ref", "--sort=refname",
		"--format=%(refname)%00%(HEAD)%00%(upstream:short)%00%(upstream:track,nobracket)", "refs/heads", "refs/remotes")
	if err != nil {
		return nil, err
	}
	var branches []Branch
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		parts := strings.Split(line, "\x00")
		if len(parts) < 4 || strings.HasSuffix(parts[0], "/HEAD") {
			continue
		}
		b := Branch{Current: parts[1] == "*", Upstream: parts[2]}
		switch {
		case strings.HasPrefix(parts[0], "refs/heads/"):
			b.Name = strings.TrimPrefix(parts[0], "refs/heads/")
		case strings.HasPrefix(parts[0], "refs/remotes/"):
			b.Name = strings.TrimPrefix(parts[0], "refs/remotes/")
			b.Remote = true
		default:
			continue
		}
		b.Ahead, b.Behind = parseTrack(parts[3])
		branches = append(branches, b)
	}
	return branches, nil
}

// parseTrack reads "ahead 2, behind 1".
func parseTrack(s string) (ahead, behind int) {
	for _, part := range strings.Split(s, ",") {
		fields := strings.Fields(part)
		if len(fields) != 2 {
			continue
		}
		n := atoi(fields[1])
		switch fields[0] {
		case "ahead":
			ahead = n
		case "behind":
			behind = n
		}
	}
	return ahead, behind
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return n
		}
		n = n*10 + int(c-'0')
	}
	return n
}

// Checkout switches to a branch. A remote branch ("origin/feature") is checked
// out as a new local branch tracking it.
func (r *Repo) Checkout(ctx context.Context, b Branch) error {
	if b.Remote {
		local := b.Name[strings.Index(b.Name, "/")+1:]
		_, err := r.run(ctx, "switch", "--track", "-c", local, b.Name)
		return err
	}
	_, err := r.run(ctx, "switch", b.Name)
	return err
}

// NewBranch creates a branch at from (HEAD when empty), switching to it when
// checkout is set.
func (r *Repo) NewBranch(ctx context.Context, name, from string, checkout bool) error {
	args := []string{"branch", name}
	if checkout {
		args = []string{"switch", "-c", name}
	}
	if from != "" {
		args = append(args, from)
	}
	_, err := r.run(ctx, args...)
	return err
}

// DeleteBranch deletes a local branch.
func (r *Repo) DeleteBranch(ctx context.Context, name string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	_, err := r.run(ctx, "branch", flag, name)
	return err
}

// Merge merges a branch into the current one.
func (r *Repo) Merge(ctx context.Context, name string) error {
	_, err := r.run(ctx, "merge", "--no-edit", name)
	return err
}

// Rebase rebases the current branch onto another.
func (r *Repo) Rebase(ctx context.Context, onto string) error {
	_, err := r.run(ctx, "rebase", onto)
	return err
}
