package lsp

import (
	"context"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// startTimeout bounds how long a server may take to initialize.
const startTimeout = 60 * time.Second

// Manager runs the language servers of one project, each started the
// first time a file that needs it opens; languages that share a server
// (Go and go.mod) share one.
type Manager struct {
	root    string
	mu      sync.Mutex
	servers map[string]*entry // by command line
}

// entry is a server being started or running.
type entry struct {
	ready chan struct{}
	srv   *Server
	err   error
}

// NewManager makes a manager for a project folder.
func NewManager(root string) *Manager {
	return &Manager{root: root, servers: map[string]*entry{}}
}

// Find returns the first command whose program is installed, or nil.
func Find(commands ...[]string) []string {
	for _, c := range commands {
		if len(c) > 0 {
			if _, err := exec.LookPath(c[0]); err == nil {
				return c
			}
		}
	}
	return nil
}

// Server returns the server args runs, starting it the first time; it
// waits while one is starting. It returns nil when there's no server
// (args is nil) or it failed to start.
func (m *Manager) Server(ctx context.Context, args []string) (*Server, error) {
	if args == nil {
		return nil, nil
	}
	key := strings.Join(args, " ")
	m.mu.Lock()
	e := m.servers[key]
	if e == nil {
		e = &entry{ready: make(chan struct{})}
		m.servers[key] = e
		go func() {
			sctx, cancel := context.WithTimeout(context.Background(), startTimeout)
			defer cancel()
			e.srv, e.err = Start(sctx, args, m.root)
			close(e.ready)
		}()
	}
	m.mu.Unlock()
	select {
	case <-e.ready:
		return e.srv, e.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Running returns the server args runs if it has started, without
// waiting.
func (m *Manager) Running(args []string) *Server {
	m.mu.Lock()
	e := m.servers[strings.Join(args, " ")]
	m.mu.Unlock()
	if e == nil {
		return nil
	}
	select {
	case <-e.ready:
		return e.srv
	default:
		return nil
	}
}

// Stop shuts every server down.
func (m *Manager) Stop() {
	m.mu.Lock()
	servers := m.servers
	m.servers = map[string]*entry{}
	m.mu.Unlock()
	for _, e := range servers {
		go func() {
			<-e.ready
			if e.srv != nil {
				e.srv.Stop()
			}
		}()
	}
}
