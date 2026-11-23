package git

import (
	"context"
	"strings"
)

// Push pushes the current branch, setting its upstream on the first push.
func (r *Repo) Push(ctx context.Context) error {
	if _, err := r.run(ctx, "rev-parse", "--abbrev-ref", "@{upstream}"); err == nil {
		_, err := r.run(ctx, "push")
		return err
	}
	branch, err := r.CurrentBranch(ctx)
	if err != nil {
		return err
	}
	remote, err := r.defaultRemote(ctx)
	if err != nil {
		return err
	}
	_, err = r.run(ctx, "push", "-u", remote, branch)
	return err
}

// Update pulls the current branch (Pull, Ctrl+T).
func (r *Repo) Update(ctx context.Context) error {
	_, err := r.run(ctx, "pull", "--no-edit")
	return err
}

// Fetch fetches every remote and prunes deleted branches.
func (r *Repo) Fetch(ctx context.Context) error {
	_, err := r.run(ctx, "fetch", "--all", "--prune")
	return err
}

// Clone clones url into dir.
func Clone(ctx context.Context, url, dir string) error {
	_, err := (&Repo{Dir: "."}).run(ctx, "clone", url, dir)
	return err
}

func (r *Repo) defaultRemote(ctx context.Context) (string, error) {
	out, err := r.run(ctx, "remote")
	if err != nil {
		return "", err
	}
	remotes := strings.Fields(out)
	for _, name := range remotes {
		if name == "origin" {
			return name, nil
		}
	}
	if len(remotes) == 0 {
		return "", errNoRemote
	}
	return remotes[0], nil
}
