package search

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestMatchFilesRanksPrefixAndHumps(t *testing.T) {
	files := []string{"ui/editor/editor_tabs.go", "internal/git/status.go", "ui/git/commit_panel.go", "docs/ET.md"}
	got := MatchFiles(files, "edtab", 10)
	if len(got) == 0 || got[0].Path != "ui/editor/editor_tabs.go" {
		t.Fatalf("%+v", got)
	}
	got = MatchFiles(files, "status", 10)
	if len(got) != 1 || got[0].Path != "internal/git/status.go" {
		t.Fatalf("%+v", got)
	}
	if got := MatchFiles(files, "git/com", 10); len(got) != 1 {
		t.Fatalf("path query %+v", got)
	}
}

func TestFindInFiles(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "a.go"), []byte("package a\nfunc Hello() {}\n// hello world\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "b.bin"), []byte("hello\x00"), 0o644)
	got, err := FindInFiles(context.Background(), dir, []string{"a.go", "b.bin"}, "hello", TextOptions{}, 10)
	if err != nil || len(got) != 2 || got[0].Line != 2 || got[0].Col != 5 || got[1].Col != 3 {
		t.Fatalf("%v %+v", err, got)
	}
	got, _ = FindInFiles(context.Background(), dir, []string{"a.go"}, "hello", TextOptions{CaseSensitive: true}, 10)
	if len(got) != 1 || got[0].Line != 3 {
		t.Fatalf("case %+v", got)
	}
}
