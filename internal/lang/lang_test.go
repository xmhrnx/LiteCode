package lang

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPluginOverridesBuiltin(t *testing.T) {
	dir := t.TempDir()
	pdir := filepath.Join(dir, "zig")
	_ = os.MkdirAll(pdir, 0o755)
	_ = os.WriteFile(filepath.Join(pdir, "plugin.json"), []byte(`{
		"name": "Zig", "version": "1.0",
		"languages": [{"id": "zig", "name": "Zig", "extensions": [".zig"], "lineComment": "//",
		  "formatter": {"args": ["zig", "fmt", "--stdin"], "stdin": true}},
		  {"id": "go2", "name": "Go (custom)", "extensions": [".go"], "lineComment": "//"}]
	}`), 0o644)
	r := NewRegistry(dir)
	if len(r.Errors) > 0 {
		t.Fatal(r.Errors)
	}
	if l := r.ForPath("/x/main.zig"); l.ID != "zig" || l.Formatter == nil {
		t.Fatalf("%+v", l)
	}
	if l := r.ForPath("a.go"); l.ID != "go2" {
		t.Fatalf("override: %+v", l)
	}
	if l := r.ForPath("Makefile"); l.ID != "makefile" || l.Indent(4) != "\t" {
		t.Fatalf("%+v", l)
	}
	if l := r.ForPath("x.unknown"); l != PlainText {
		t.Fatalf("%+v", l)
	}
}
