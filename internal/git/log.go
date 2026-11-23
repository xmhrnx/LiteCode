package git

import (
	"context"
	"strconv"
	"strings"
	"time"
)

// Commit is one entry of the history.
type Commit struct {
	Hash    string
	Short   string
	Author  string
	Email   string
	Time    time.Time
	Refs    []string // branch names pointing at it
	Tags    []string
	Subject string
	Parents []string
}

const logFormat = "%H%x00%h%x00%an%x00%ae%x00%at%x00%D%x00%s%x00%P%x1e"

// Log returns up to limit commits of every branch, newest first. An empty
// branch means all branches.
func (r *Repo) Log(ctx context.Context, branch string, limit int) ([]Commit, error) {
	args := []string{"log", "--date-order", "-n", strconv.Itoa(limit), "--format=" + logFormat}
	if branch == "" {
		args = append(args, "--branches", "--remotes", "--tags", "HEAD")
	} else {
		args = append(args, branch)
	}
	out, err := r.run(ctx, args...)
	if err != nil {
		return nil, err
	}
	return parseLog(out), nil
}

// Outgoing lists commits on the current branch not yet on its upstream.
func (r *Repo) Outgoing(ctx context.Context) ([]Commit, error) {
	out, err := r.run(ctx, "log", "--format="+logFormat, "@{upstream}..HEAD")
	if err != nil {
		// No upstream: everything on the branch is outgoing.
		out, err = r.run(ctx, "log", "-n", "50", "--format="+logFormat, "HEAD")
		if err != nil {
			return nil, err
		}
	}
	return parseLog(out), nil
}

func parseLog(out string) []Commit {
	var commits []Commit
	for _, rec := range strings.Split(out, "\x1e") {
		rec = strings.TrimLeft(rec, "\n")
		f := strings.Split(rec, "\x00")
		if len(f) < 8 {
			continue
		}
		secs, _ := strconv.ParseInt(f[4], 10, 64)
		c := Commit{Hash: f[0], Short: f[1], Author: f[2], Email: f[3], Time: time.Unix(secs, 0), Subject: f[6]}
		if f[7] = strings.TrimSpace(f[7]); f[7] != "" {
			c.Parents = strings.Fields(f[7])
		}
		for _, ref := range strings.Split(f[5], ", ") {
			ref = strings.TrimPrefix(strings.TrimSpace(ref), "HEAD -> ")
			switch {
			case strings.HasPrefix(ref, "tag: "):
				c.Tags = append(c.Tags, strings.TrimPrefix(ref, "tag: "))
			case ref != "" && ref != "HEAD" && !strings.HasSuffix(ref, "/HEAD"):
				c.Refs = append(c.Refs, ref)
			}
		}
		commits = append(commits, c)
	}
	return commits
}
