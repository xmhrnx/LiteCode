package lsp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGopls(t *testing.T) {
	args := Find("", []string{"gopls"})
	if args == nil {
		t.Skip("gopls not installed")
	}
	dir := t.TempDir()
	write := func(name, text string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	write("go.mod", "module example.com/m\n\ngo 1.22\n")
	text := "package main\n\nfunc main() {\n\tstrings.Spl\n}\n"
	path := write("main.go", text)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	m := NewManager(dir)
	defer m.Stop()
	s, err := m.Server(ctx, [][]string{args})
	if err != nil || s == nil {
		t.Fatal(err)
	}
	if err := s.Open(path, "go", 1, text); err != nil {
		t.Fatal(err)
	}
	var found *Completion
	for try := 0; try < 20 && found == nil; try++ {
		items, err := s.Complete(ctx, path, Position{Line: 3, Character: 12}, "")
		if err != nil {
			t.Fatal(err)
		}
		for i := range items {
			if items[i].Label == "Split" {
				found = &items[i]
			}
		}
		if found == nil {
			time.Sleep(500 * time.Millisecond)
		}
	}
	if found == nil {
		t.Fatal("no Split completion")
	}
	c, err := s.Resolve(ctx, *found)
	if err != nil {
		t.Fatal(err)
	}
	imported := false
	for _, e := range c.Extra {
		if strings.Contains(e.NewText, `"strings"`) {
			imported = true
		}
	}
	if !imported {
		t.Fatalf("no import edit: %+v", c.Extra)
	}
}
