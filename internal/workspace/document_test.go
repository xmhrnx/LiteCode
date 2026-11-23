package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCRLFRoundTrip(t *testing.T) {
	p := filepath.Join(t.TempDir(), "a.txt")
	_ = os.WriteFile(p, []byte("a\r\nb\r\n"), 0o600)
	d, err := Open(p)
	if err != nil || d.Saved != "a\nb\n" || !d.CRLF {
		t.Fatalf("%v %+v", err, d)
	}
	if !d.Modified("a\nc\n") {
		t.Fatal("modified")
	}
	if err := d.Save("a\nc\n"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(p)
	info, _ := os.Stat(p)
	if string(data) != "a\r\nc\r\n" || info.Mode().Perm() != 0o600 {
		t.Fatalf("%q %v", data, info.Mode())
	}
	_ = os.WriteFile(p, []byte{0, 1, 2}, 0o600)
	if _, err := Open(p); err != ErrBinary {
		t.Fatal(err)
	}
}
