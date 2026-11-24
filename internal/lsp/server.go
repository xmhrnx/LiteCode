package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"litecode/internal/proc"
)

// Server is a running language server.
type Server struct {
	Name   string // the program, for the status bar
	cmd    *exec.Cmd
	conn   *conn
	cancel context.CancelFunc
	done   chan struct{}
	// resolves is set when the server fills in completions on request.
	resolves bool
	// triggers are the characters that ask the server for completions.
	triggers []string
}

// URI is a file path as a file:// URI.
func URI(path string) string {
	path = filepath.ToSlash(path)
	if runtime.GOOS == "windows" {
		path = "/" + path
	}
	return (&url.URL{Scheme: "file", Path: path}).String()
}

// Start runs a language server for a project folder and initializes it.
func Start(ctx context.Context, args []string, root string) (*Server, error) {
	if len(args) == 0 {
		return nil, errors.New("no language server command")
	}
	sctx, cancel := context.WithCancel(context.Background())
	cmd := proc.Command(sctx, args[0], args[1:]...)
	cmd.Dir = root
	cmd.Stderr = io.Discard
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, err
	}
	s := &Server{Name: filepath.Base(args[0]), cmd: cmd, conn: newConn(stdin), cancel: cancel, done: make(chan struct{})}
	go func() {
		s.conn.read(stdout)
		close(s.done)
	}()
	go func() { _ = cmd.Wait() }()
	if err := s.initialize(ctx, root); err != nil {
		s.Stop()
		return nil, err
	}
	return s, nil
}

// request sends a request and waits for its result, or for ctx.
func (s *Server) request(ctx context.Context, method string, params, result any) error {
	id, ch := s.conn.call(method, params)
	select {
	case r := <-ch:
		if r.err != nil {
			return r.err
		}
		if result == nil || len(r.result) == 0 {
			return nil
		}
		return json.Unmarshal(r.result, result)
	case <-ctx.Done():
		s.conn.cancel(id)
		return ctx.Err()
	}
}

func (s *Server) initialize(ctx context.Context, root string) error {
	params := map[string]any{
		"processId": os.Getpid(),
		"rootUri":   URI(root),
		"workspaceFolders": []map[string]string{
			{"uri": URI(root), "name": filepath.Base(root)},
		},
		"clientInfo": map[string]string{"name": "LiteCode"},
		"capabilities": map[string]any{
			"textDocument": map[string]any{
				"synchronization": map[string]any{"didSave": true},
				"completion": map[string]any{
					"completionItem": map[string]any{
						"snippetSupport":       false,
						"insertReplaceSupport": false,
						"labelDetailsSupport":  true,
						"documentationFormat":  []string{"plaintext"},
						"resolveSupport":       map[string]any{"properties": []string{"additionalTextEdits", "detail"}},
					},
					"contextSupport": true,
				},
			},
			"workspace": map[string]any{"workspaceFolders": true, "configuration": true},
			"general":   map[string]any{"positionEncodings": []string{"utf-16"}},
		},
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	var result struct {
		Capabilities struct {
			CompletionProvider *struct {
				ResolveProvider   bool     `json:"resolveProvider"`
				TriggerCharacters []string `json:"triggerCharacters"`
			} `json:"completionProvider"`
		} `json:"capabilities"`
	}
	if err := s.request(ctx, "initialize", params, &result); err != nil {
		return err
	}
	if p := result.Capabilities.CompletionProvider; p != nil {
		s.resolves, s.triggers = p.ResolveProvider, p.TriggerCharacters
	}
	return s.conn.notify("initialized", map[string]any{})
}

// Open tells the server a document is open with this text.
func (s *Server) Open(path, languageID string, version int, text string) error {
	return s.conn.notify("textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": URI(path), "languageId": languageID, "version": version, "text": text},
	})
}

// Change gives the server a document's whole new text.
func (s *Server) Change(path string, version int, text string) error {
	return s.conn.notify("textDocument/didChange", map[string]any{
		"textDocument":   map[string]any{"uri": URI(path), "version": version},
		"contentChanges": []map[string]string{{"text": text}},
	})
}

// Saved tells the server a document was written.
func (s *Server) Saved(path string) error {
	return s.conn.notify("textDocument/didSave", map[string]any{"textDocument": map[string]string{"uri": URI(path)}})
}

// Close tells the server a document was closed.
func (s *Server) Close(path string) error {
	return s.conn.notify("textDocument/didClose", map[string]any{"textDocument": map[string]string{"uri": URI(path)}})
}

// Complete asks for the completions at a position. trigger is the
// character that asked for them ("." after a value), or "" when typing
// or Ctrl+Space did.
func (s *Server) Complete(ctx context.Context, path string, at Position, trigger string) ([]Completion, error) {
	cc := map[string]any{"triggerKind": 1}
	if trigger != "" {
		cc = map[string]any{"triggerKind": 2, "triggerCharacter": trigger}
	}
	var result json.RawMessage
	err := s.request(ctx, "textDocument/completion", map[string]any{
		"textDocument": map[string]string{"uri": URI(path)},
		"position":     at,
		"context":      cc,
	}, &result)
	if err != nil {
		return nil, err
	}
	// A list, or {isIncomplete, items}.
	var items []json.RawMessage
	if strings.HasPrefix(strings.TrimSpace(string(result)), "[") {
		_ = json.Unmarshal(result, &items)
	} else {
		var list struct {
			Items []json.RawMessage `json:"items"`
		}
		_ = json.Unmarshal(result, &list)
		items = list.Items
	}
	out := make([]Completion, 0, len(items))
	for _, raw := range items {
		if c, err := decode(raw); err == nil {
			out = append(out, c)
		}
	}
	return out, nil
}

// Triggers are the characters after which the server has completions
// ("." in most languages).
func (s *Server) Triggers() []string { return s.triggers }

// Resolve fills in what a server leaves out of its first answer, such as
// the import a completion needs. Servers that send everything at once
// aren't asked.
func (s *Server) Resolve(ctx context.Context, c Completion) (Completion, error) {
	if c.raw == nil || !s.resolves || len(c.Extra) > 0 {
		return c, nil
	}
	var raw json.RawMessage
	if err := s.request(ctx, "completionItem/resolve", c.raw, &raw); err != nil || len(raw) == 0 || string(raw) == "null" {
		return c, err
	}
	r, err := decode(raw)
	if err != nil {
		return c, err
	}
	if len(r.Extra) == 0 {
		r.Extra = c.Extra
	}
	if r.Edit == nil {
		r.Edit = c.Edit
	}
	return r, nil
}

// Stop shuts the server down, killing it if it doesn't go.
func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if s.request(ctx, "shutdown", nil, nil) == nil {
		_ = s.conn.notify("exit", nil)
	}
	select {
	case <-s.done:
	case <-time.After(time.Second):
	}
	s.cancel()
}
