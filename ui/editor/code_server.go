package editor

import (
	"context"
	"time"

	"litecode/internal/lsp"
)

// completeTimeout bounds how long a completion request may take.
const completeTimeout = 4 * time.Second

// serverState is a view's link to its language server: which one, and
// what it has been told.
type serverState struct {
	args    []string    // the server's command; nil when none is installed
	srv     *lsp.Server // set once it has started
	sentRev int         // the text revision the server has
	failed  error
}

// startServer starts (or joins) the language server for the view's file
// and opens the document in it.
func (v *CodeView) startServer() {
	if v.ide.Project == nil {
		return
	}
	v.server.args = lsp.Find(v.lang.ServerCommands()...)
	if v.server.args == nil {
		return
	}
	servers, args := v.ide.Project.Servers, v.server.args
	go func() {
		srv, err := servers.Server(context.Background(), args)
		v.ide.App.Post(func() {
			if v.closed {
				return
			}
			v.server.srv, v.server.failed = srv, err
			if srv != nil {
				rev := v.editor.buf.Rev()
				if srv.Open(v.doc.Path, v.lang.ServerLanguageID(v.doc.Path), rev, v.editor.Text()) == nil {
					v.server.sentRev = rev
				}
			}
		})
	}()
}

// ServerName is the running language server's name, "" for none.
func (v *CodeView) ServerName() string {
	if v.server.srv == nil {
		return ""
	}
	return v.server.srv.Name
}

// syncServer sends the server the text when it changed since it last got it.
func (v *CodeView) syncServer() {
	if s := v.server.srv; s != nil {
		if rev := v.editor.buf.Rev(); rev != v.server.sentRev {
			if s.Change(v.doc.Path, rev, v.editor.Text()) == nil {
				v.server.sentRev = rev
			}
		}
	}
}

// ServerTriggers are the characters after which the server has completions.
func (v *CodeView) ServerTriggers() []string {
	if v.server.srv == nil {
		return nil
	}
	return v.server.srv.Triggers()
}

// ServerComplete asks the server for completions at a position (its column in
// UTF-16 units); done gets them on the UI goroutine, with ok false when
// there's no server or it failed.
func (v *CodeView) ServerComplete(line, char16 int, trigger string, done func(items []lsp.Completion, ok bool)) {
	s := v.server.srv
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
}

// ServerResolve fills in a completion's edits (an import) when the server sends
// them only on request; done runs on the UI goroutine.
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
	if s := v.server.srv; s != nil {
		_ = s.Close(v.doc.Path)
	}
}
