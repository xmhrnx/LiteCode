package editor

import (
	"context"
	"time"

	"github.com/xmhrnx/giode"

	"litecode/internal/lsp"
)

// Timeouts of a completion request, and of waiting for a server that is
// still starting (a large Go module takes gopls a while to load).
const (
	completeTimeout = 5 * time.Second
	startWait       = 30 * time.Second
)

// serverState is a view's link to its language server: which one, and
// what it has been told.
type serverState struct {
	args    [][]string  // the installed servers to try, in order; nil for none
	srv     *lsp.Server // set once it has started
	opened  bool        // the server has the document
	sentRev int         // the text revision the server has
	failed  error
}

// startServer starts (or joins) the language server for the view's file
// and opens the document in it. With no server installed it offers to
// install one.
func (v *CodeView) startServer() {
	if v.ide.Project == nil {
		return
	}
	v.server = serverState{args: lsp.FindAll(v.ide.Project.Root, v.lang.ServerCommands()...)}
	if v.server.args == nil {
		v.offerInstall()
		return
	}
	v.whenServer(func(*lsp.Server) {})
}

// whenServer runs fn with the server once it has started and has the
// document, at once when it already has; fn gets nil when there's none.
func (v *CodeView) whenServer(fn func(*lsp.Server)) {
	if v.server.srv != nil {
		fn(v.server.srv)
		return
	}
	if v.server.args == nil || v.server.failed != nil || v.ide.Project == nil {
		fn(nil)
		return
	}
	servers, args := v.ide.Project.Servers, v.server.args
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), startWait)
		defer cancel()
		srv, err := servers.Server(ctx, args)
		v.ide.App.Post(func() {
			if v.closed {
				return
			}
			if err != nil && err != context.DeadlineExceeded && v.server.failed == nil {
				v.server.failed = err
				if key := "failed:" + v.lang.ID; !v.ide.ServerOffers[key] {
					v.ide.ServerOffers[key] = true
					v.ide.Toasts.Show(giode.ToastError, "The "+v.lang.Name+" language server didn't start: "+err.Error())
				}
			}
			if srv != nil && !v.server.opened {
				rev := v.editor.buf.Rev()
				if srv.Open(v.doc.Path, v.lang.ServerLanguageID(v.doc.Path), rev, v.editor.Text()) == nil {
					v.server.srv, v.server.opened, v.server.sentRev = srv, true, rev
				}
			}
			fn(v.server.srv)
			v.ide.App.Invalidate()
		})
	}()
}

// offerInstall tells the user, once per language and session, that the
// language has no server, offering to install it when a tool that
// installs it is there. The status bar keeps an Install button after the
// message goes.
func (v *CodeView) offerInstall() {
	id := v.lang.ID
	if v.ide.ServerOffers[id] || len(v.lang.ServerCommands()) == 0 {
		return
	}
	v.ide.ServerOffers[id] = true
	in, ok := v.Installer()
	if !ok {
		v.ide.Toasts.Show(giode.ToastInfo, "Code completion for "+v.lang.Name+" needs a language server ("+v.lang.ServerCommands()[0][0]+"), which isn't installed.")
		return
	}
	v.ide.Toasts.ShowAction(giode.ToastInfo, "Code completion for "+v.lang.Name+" needs "+in.Server+", which isn't installed.", "Install", v.InstallServer)
}

// Installer is how to install the file's missing language server, when
// there's no server and a tool that installs one is there.
func (v *CodeView) Installer() (lsp.Installer, bool) {
	if v.server.args != nil || v.ide.Project == nil {
		return lsp.Installer{}, false
	}
	return lsp.InstallerFor(v.lang.ID, v.ide.Project.Root)
}

// InstallServer installs the file's language server, then starts it in
// every open file it serves.
func (v *CodeView) InstallServer() {
	in, ok := v.Installer()
	if !ok || v.ide.Installing[in.Server] {
		return
	}
	ide, name := v.ide, v.lang.Name
	ide.Installing[in.Server] = true
	ide.Info("Installing " + in.Server + "… (" + in.Command() + ")")
	ide.Background(func(ctx context.Context) error {
		err := lsp.Install(ctx, in)
		ide.App.Post(func() { delete(ide.Installing, in.Server) })
		return err
	}, func() {
		ide.Info(in.Server + " is installed: completion for " + name + " is on.")
		if ide.ServersInstalled != nil {
			ide.ServersInstalled(v.lang.ID)
		}
	})
}

// Installing reports whether the file's language server is being installed.
func (v *CodeView) Installing() bool {
	in, ok := v.Installer()
	return ok && v.ide.Installing[in.Server]
}

// ServerStatus says how the view's language server is doing, for the
// status bar: its name, that it's starting, or that there's none.
func (v *CodeView) ServerStatus() string {
	switch {
	case v.server.srv != nil:
		return v.server.srv.Name
	case v.server.failed != nil:
		return "language server failed"
	case v.server.args != nil:
		return "starting language server…"
	case len(v.lang.ServerCommands()) > 0:
		return "no language server"
	}
	return ""
}

// syncServer sends the server the text when it changed since it last got it.
func (v *CodeView) syncServer() {
	if s := v.server.srv; s != nil && v.server.opened {
		if rev := v.editor.buf.Rev(); rev != v.server.sentRev {
			if s.Change(v.doc.Path, rev, v.editor.Text()) == nil {
				v.server.sentRev = rev
			}
		}
	}
}

// HasServer reports whether a language server is running or starting for
// the file: completions should wait for it rather than guess.
func (v *CodeView) HasServer() bool {
	return v.server.srv != nil || v.server.args != nil && v.server.failed == nil
}

// ServerTriggers are the characters after which the server has
// completions; "." while it's still starting.
func (v *CodeView) ServerTriggers() []string {
	switch {
	case v.server.srv != nil:
		return v.server.srv.Triggers()
	case v.HasServer():
		return []string{"."}
	}
	return nil
}

// ServerComplete asks the server for completions at a position (its
// column in UTF-16 units), waiting for it to start if it's starting; done
// gets them on the UI goroutine, with ok false when there's no server or
// it failed.
func (v *CodeView) ServerComplete(line, char16 int, trigger string, done func(items []lsp.Completion, ok bool)) {
	v.whenServer(func(s *lsp.Server) {
		if s == nil {
			done(nil, false)
			return
		}
		v.syncServer()
		path := v.doc.Path
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), completeTimeout)
			defer cancel()
			items, err := s.Complete(ctx, path, lsp.Position{Line: line, Character: char16}, trigger)
			v.ide.App.Post(func() { done(items, err == nil) })
		}()
	})
}

// ServerResolve fills in a completion's edits (an import) when the server
// sends them only on request; done runs on the UI goroutine.
func (v *CodeView) ServerResolve(c lsp.Completion, done func(lsp.Completion)) {
	s := v.server.srv
	if s == nil {
		done(c)
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), completeTimeout)
		defer cancel()
		r, err := s.Resolve(ctx, c)
		if err != nil {
			r = c
		}
		v.ide.App.Post(func() { done(r) })
	}()
}

// Dispose tells the server the file closed.
func (v *CodeView) Dispose() {
	v.closed = true
	if s := v.server.srv; s != nil && v.server.opened {
		_ = s.Close(v.doc.Path)
	}
}

// RetryServer looks for the language server again (after one was
// installed) when the view has none.
func (v *CodeView) RetryServer() {
	if v.server.srv == nil {
		v.startServer()
	}
}
