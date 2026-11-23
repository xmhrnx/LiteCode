package format

import (
	"context"
	"os/exec"
	"testing"

	"litecode/internal/lang"
)

func TestGofmt(t *testing.T) {
	if _, err := exec.LookPath("gofmt"); err != nil {
		t.Skip("gofmt not installed")
	}
	l := lang.NewRegistry("").ForPath("a.go")
	got, err := Format(context.Background(), t.TempDir(), "a.go", "package a\nfunc  x( ){}\n", l, Options{})
	if err != nil || got != "package a\n\nfunc x() {}\n" {
		t.Fatalf("%v %q", err, got)
	}
	if _, err := Format(context.Background(), ".", "a.txt", "x", lang.PlainText, Options{}); err != ErrNoFormatter {
		t.Fatal(err)
	}
}
