package core

import (
	"context"
	"time"

	"litecode/internal/git"
)

// LogLimit is how many commits the Git window loads.
const LogLimit = 1000

// GitState is the repository as the UI last saw it. It's only changed on the
// UI goroutine; Refresh loads in the background and posts the result.
type GitState struct {
	repo     *git.Repo
	Branch   string
	Branches []git.Branch
	Changes  []git.Change
	// UntrackedDirs are folders with nothing versioned in them.
	UntrackedDirs []string
	Log           []git.Commit
	// Busy names the operation running ("Pushing…"), empty when idle.
	Busy string
	// Version goes up whenever anything above changes, for caches.
	Version int

	refreshing bool
	pending    bool // a refresh was asked for while one ran
	pendingLog bool
}

// Repo returns the repository.
func (g *GitState) Repo() *git.Repo { return g.repo }

// Refresh reloads the branch, branches and changes; withLog reloads the log too.
func (g *GitState) Refresh(ide *IDE, withLog bool) {
	if g.refreshing {
		g.pending = true
		g.pendingLog = g.pendingLog || withLog
		return
	}
	g.refreshing = true
	go func() {
		ctx := context.Background()
		branch, errB := g.repo.CurrentBranch(ctx)
		branches, errBs := g.repo.Branches(ctx)
		changes, errC := g.repo.Status(ctx)
		dirs, _ := g.repo.UntrackedDirs(ctx)
		var log []git.Commit
		var errL error
		if withLog {
			log, errL = g.repo.Log(ctx, "", LogLimit)
		}
		ide.App.Post(func() {
			g.refreshing = false
			if g.pending {
				again := g.pendingLog
				g.pending, g.pendingLog = false, false
				defer g.Refresh(ide, again)
			}
			if errC != nil {
				ide.Error(errC)
				return
			}
			if errB == nil {
				g.Branch = branch
			}
			if errBs == nil {
				g.Branches = branches
			}
			g.Changes, g.UntrackedDirs = changes, dirs
			if withLog && errL == nil {
				g.Log = log
			}
			g.Version++
		})
	}()
}

// Do runs a git operation in the background with a busy label, then
// refreshes everything and calls done (unless the operation failed).
func (g *GitState) Do(ide *IDE, label string, op func(ctx context.Context, r *git.Repo) error, done func()) {
	if g.Busy != "" {
		ide.Info("Wait for " + g.Busy + " to finish")
		return
	}
	g.Busy = label
	g.Version++
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		err := op(ctx, g.repo)
		ide.App.Post(func() {
			g.Busy = ""
			g.Version++
			g.Refresh(ide, true)
			if err != nil {
				ide.Error(err)
				return
			}
			if done != nil {
				done()
			}
		})
	}()
}
