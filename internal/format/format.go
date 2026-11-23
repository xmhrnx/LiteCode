// Package format runs code formatters. Prettier is found once per project
// (its node_modules/.bin first, then PATH) and used for every project
// without setup; other languages use the formatter their language declares.
package format

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"litecode/internal/lang"
	"litecode/internal/proc"
)

// Timeout bounds a formatter run.
const Timeout = 10 * time.Second

// ErrNoFormatter means nothing formats this language.
var ErrNoFormatter = errors.New("no formatter for this file")

// Options says which formatters may run.
type Options struct {
	Prettier bool
}

// Format returns text formatted for the file at path inside root.
func Format(ctx context.Context, root, path, text string, l *lang.Language, o Options) (string, error) {
	cmd := commandFor(root, l, o)
	if cmd == nil {
		return text, ErrNoFormatter
	}
	args := expand(cmd.Args, root, path)
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()
	c := proc.Command(ctx, args[0], args[1:]...)
	c.Dir = root
	var stdout, stderr bytes.Buffer
	c.Stdout, c.Stderr = &stdout, &stderr
	if cmd.Stdin {
		c.Stdin = strings.NewReader(text)
	}
	if err := c.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return text, fmt.Errorf("%s: %s", filepath.Base(args[0]), firstLine(msg))
	}
	if !cmd.Stdin {
		data, err := os.ReadFile(path)
		return string(data), err
	}
	return stdout.String(), nil
}

// Available reports whether a formatter would run for the language.
func Available(root string, l *lang.Language, o Options) bool {
	return commandFor(root, l, o) != nil
}

func commandFor(root string, l *lang.Language, o Options) *lang.Command {
	if o.Prettier && l.Prettier {
		if bin := FindPrettier(root); bin != "" {
			return &lang.Command{Args: []string{bin, "--stdin-filepath", "${file}"}, Stdin: true}
		}
	}
	if l.Formatter == nil || len(l.Formatter.Args) == 0 {
		return nil
	}
	if _, err := exec.LookPath(l.Formatter.Args[0]); err != nil {
		return nil
	}
	return l.Formatter
}

// FindPrettier returns the project's own Prettier, else one on PATH, else "".
func FindPrettier(root string) string {
	name := "prettier"
	if runtime.GOOS == "windows" {
		name = "prettier.cmd"
	}
	for dir := root; ; dir = filepath.Dir(dir) {
		p := filepath.Join(dir, "node_modules", ".bin", name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
		if filepath.Dir(dir) == dir {
			break
		}
	}
	if p, err := exec.LookPath("prettier"); err == nil {
		return p
	}
	return ""
}

func expand(args []string, root, path string) []string {
	out := make([]string, len(args))
	for i, a := range args {
		a = strings.ReplaceAll(a, "${file}", path)
		out[i] = strings.ReplaceAll(a, "${root}", root)
	}
	return out
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
