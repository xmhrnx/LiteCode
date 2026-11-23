// Package workspace holds the files open in the editor and the per-project
// session (which files were open) so a project reopens as it was left.
package workspace

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"unicode/utf8"
)

// MaxEditableSize is the largest file the editor opens.
const MaxEditableSize = 8 << 20

// ErrBinary means the file isn't text.
var ErrBinary = errors.New("binary file")

// ErrTooLarge means the file is over MaxEditableSize.
var ErrTooLarge = errors.New("file too large to edit")

// Document is a file open in the editor.
type Document struct {
	Path  string
	Saved string // the text as last read or written
	CRLF  bool   // the file used Windows line endings; kept on save
}

// Open reads a text file. Line endings are normalised to "\n" while editing.
func Open(path string) (*Document, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() > MaxEditableSize {
		return nil, ErrTooLarge
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if bytes.IndexByte(data[:min(len(data), 8000)], 0) >= 0 || !utf8.Valid(data) {
		return nil, ErrBinary
	}
	text := string(data)
	d := &Document{Path: path, CRLF: strings.Contains(text, "\r\n")}
	d.Saved = strings.ReplaceAll(text, "\r\n", "\n")
	return d, nil
}

// Modified reports whether text differs from what's on disk.
func (d *Document) Modified(text string) bool {
	return text != d.Saved
}

// Save writes text to the file, keeping its permissions and line endings.
func (d *Document) Save(text string) error {
	mode := os.FileMode(0o644)
	if info, err := os.Stat(d.Path); err == nil {
		mode = info.Mode().Perm()
	}
	out := text
	if d.CRLF {
		out = strings.ReplaceAll(text, "\n", "\r\n")
	}
	if err := os.WriteFile(d.Path, []byte(out), mode); err != nil {
		return err
	}
	d.Saved = text
	return nil
}
