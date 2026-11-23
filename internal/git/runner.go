// Package git drives the git command line for one repository. Every call runs
// `git` as a child process, so nothing is loaded until it's asked for.
package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"litecode/internal/proc"
)

// Repo is a working tree on disk.
type Repo struct {
	Dir string
}

// ErrNotRepo is returned by Open for a folder outside any git working tree.
var ErrNotRepo = errors.New("not a git repository")

// Open returns the repository containing dir, rooted at its top level.
func Open(ctx context.Context, dir string) (*Repo, error) {
	out, err := (&Repo{Dir: dir}).run(ctx, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, ErrNotRepo
	}
	return &Repo{Dir: strings.TrimSpace(out)}, nil
}

// run executes git with args in the repository and returns stdout.
func (r *Repo) run(ctx context.Context, args ...string) (string, error) {
	return r.runInput(ctx, nil, args...)
}

func (r *Repo) runInput(ctx context.Context, stdin []byte, args ...string) (string, error) {
	cmd := proc.Command(ctx, "git", append([]string{"-c", "core.quotepath=off", "-c", "color.ui=never"}, args...)...)
	cmd.Dir = r.Dir
	cmd.Env = append(cmd.Environ(), "GIT_TERMINAL_PROMPT=0", "LC_ALL=C")
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return stdout.String(), fmt.Errorf("git %s: %s", args[0], msg)
	}
	return stdout.String(), nil
}
