package lsp

import (
	"context"
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

// Server returns the first of the candidate commands that starts,
// starting it the first time; it waits while one is starting. It returns
// nil when there are no candidates or none started.
func (m *Manager) Server(ctx context.Context, candidates [][]string) (*Server, error) {
	if len(candidates) == 0 {
		return nil, nil
	}
	key := keyOf(candidates)
	m.mu.Lock()
	e := m.servers[key]
	if e == nil {
		e = &entry{ready: make(chan struct{})}
		m.servers[key] = e
		go func() {
			sctx, cancel := context.WithTimeout(context.Background(), startTimeout)
			defer cancel()
			for _, args := range candidates {
				if e.srv, e.err = Start(sctx, args, m.root); e.err == nil {
					break
				}
			}
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

// keyOf names a candidate list, so languages with the same servers share
// one.
func keyOf(candidates [][]string) string {
	parts := make([]string, len(candidates))
	for i, c := range candidates {
		parts[i] = strings.Join(c, " ")
	}
	return strings.Join(parts, "|")
}

// Running returns the server the candidates started, without waiting.
func (m *Manager) Running(candidates [][]string) *Server {
	m.mu.Lock()
	e := m.servers[keyOf(candidates)]
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
