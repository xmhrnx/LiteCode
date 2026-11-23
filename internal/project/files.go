// Package project reads a project folder: its file list for Go to File and
// Find in Files, and directory listings for the project tree.
package project

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// Entry is one item of a directory listing.
type Entry struct {
	Name  string
	Path  string // absolute
	IsDir bool
}

// SkippedDirs are never listed or searched outside a git repository.
var SkippedDirs = map[string]bool{
	".git": true, "node_modules": true, ".idea": true, ".vscode": true,
	"dist": true, "build": true, "target": true, "vendor": true, ".next": true,
	"__pycache__": true, ".venv": true, ".gradle": true,
}

// Files lists the project's files, relative and slash separated. In a git
// repository that is tracked plus untracked-but-not-ignored files, so
// .gitignore is honoured; elsewhere the tree is walked, skipping SkippedDirs.
func Files(ctx context.Context, root string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "ls-files", "-co", "--exclude-standard", "-z")
	cmd.Dir = root
	if out, err := cmd.Output(); err == nil {
		var files []string
		for _, f := range strings.Split(string(out), "\x00") {
			if f != "" {
				files = append(files, f)
			}
		}
		sort.Strings(files)
		return files, nil
	}
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if d.IsDir() {
			if path != root && SkippedDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	return files, err
}

// ReadDir lists a directory the way the project tree shows it: folders
// first, then files, each in case-insensitive name order.
func ReadDir(dir string) ([]Entry, error) {
	des, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	entries := make([]Entry, 0, len(des))
	for _, d := range des {
		if d.Name() == ".git" {
			continue
		}
		entries = append(entries, Entry{Name: d.Name(), Path: filepath.Join(dir, d.Name()), IsDir: d.IsDir()})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
	return entries, nil
}
