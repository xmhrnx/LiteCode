package gitui

import (
	"context"

	"github.com/xmhrnx/giode"

	"litecode/internal/git"
	"litecode/internal/keymap"
	"litecode/ui/core"
	"litecode/ui/theme"
)

// branchMenu is what a branch offers in the branches popup and the log's
// branches panel. newBranch asks for a name and creates it from a start
// point; done (which may be nil) runs before any chosen action, to close the
// popup the menu came from.
func branchMenu(ide *core.IDE, b git.Branch, newBranch func(from string), done func()) *giode.MenuElement {
	g := ide.Project.Git
	m := theme.NewMenu()
	act := func(fn func()) func() {
		return func() {
			if done != nil {
				done()
			}
			fn()
		}
	}
	if !b.Current {
		m.Item("Checkout", act(func() { checkout(ide, b) }))
	}
	m.Item("New Branch from '"+b.Name+"'…", act(func() { newBranch(b.Name) }))
	if !b.Current {
		m.Separator().
			Item("Merge '"+b.Name+"' into '"+g.Branch+"'", act(func() {
				g.Do(ide, "Merging…", func(ctx context.Context, r *git.Repo) error { return r.Merge(ctx, b.Name) }, nil)
			})).
			Item("Rebase '"+g.Branch+"' onto '"+b.Name+"'", act(func() {
				g.Do(ide, "Rebasing…", func(ctx context.Context, r *git.Repo) error { return r.Rebase(ctx, b.Name) }, nil)
			}))
	}
	if b.Current {
		m.Separator().
			Item("Update", act(func() {
				g.Do(ide, "Updating…", func(ctx context.Context, r *git.Repo) error { return r.Update(ctx) }, nil)
			})).
			Item("Push…", act(func() { ide.Run(keymap.PushAction) }))
	}
	if !b.Current && !b.Remote {
		m.Separator().Item("Delete", act(func() {
			g.Do(ide, "Deleting "+b.Name+"…", func(ctx context.Context, r *git.Repo) error { return r.DeleteBranch(ctx, b.Name, false) }, nil)
		}))
	}
	return m
}

// checkout switches to a branch.
func checkout(ide *core.IDE, b git.Branch) {
	ide.Project.Git.Do(ide, "Checking out "+b.Name+"…", func(ctx context.Context, r *git.Repo) error { return r.Checkout(ctx, b) }, nil)
}
