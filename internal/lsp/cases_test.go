package lsp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"litecode/internal/complete"
)

// TestGoplsCases checks what the editor would list first for typical Go
// completions: a package member, a struct's fields and methods, and keys
// in a struct literal.
func TestGoplsCases(t *testing.T) {
	args := Find("", []string{"gopls"})
	if args == nil {
		t.Skip("gopls not installed")
	}
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/m\n\ngo 1.22\n"), 0o644)
	const src = `package main

import "encoding/json"

type Person struct {
	Name    string
	Age     int
	Address string
}

func (p Person) Greeting() string { return "hi " + p.Name }

func main() {
	p := Person{}
	data, _ := json.M
	_ = p.Gr
	_ = p.Ad
	_ = Person{Na}
}
`
	path := filepath.Join(dir, "main.go")
	os.WriteFile(path, []byte(src), 0o644)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	m := NewManager(dir)
	defer m.Stop()
	s, err := m.Server(ctx, [][]string{args})
	if err != nil {
		t.Fatal(err)
	}
	s.Open(path, "go", 1, src)
	lines := strings.Split(src, "\n")
	for _, c := range []struct{ after, typed, want string }{
		{"json.M", "M", "Marshal"},
		{"p.Gr", "Gr", "Greeting"},
		{"p.Ad", "Ad", "Address"},
		{"Person{Na", "Na", "Name"},
	} {
		line, col := -1, 0
		for i, l := range lines {
			if j := strings.Index(l, c.after); j >= 0 {
				line, col = i, j+len(c.after)
			}
		}
		var top string
		for try := 0; try < 20 && top != c.want; try++ {
			items, err := s.Complete(ctx, path, Position{Line: line, Character: col}, "")
			if err != nil {
				t.Fatal(err)
			}
			ranked := make([]complete.Item, len(items))
			for i, it := range items {
				ranked[i] = complete.Item{Label: it.Label, Filter: it.Filter, Sort: it.SortText, Index: i}
			}
			if r := complete.Rank(ranked, c.typed); len(r) > 0 {
				top = r[0].Label
			}
			if top != c.want {
				time.Sleep(300 * time.Millisecond)
			}
		}
		if top != c.want {
			t.Errorf("after %q: first is %q, want %q", c.after, top, c.want)
		}
	}
}
