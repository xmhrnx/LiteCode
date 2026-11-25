package lsp

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// searchDirs are where language servers and their installers usually
// land besides the PATH. An app started from the desktop often doesn't
// get the PATH a terminal has, so `go install` and `npm i -g` results
// would be missed.
func searchDirs(root string) []string {
	home, _ := os.UserHomeDir()
	var dirs []string
	add := func(parts ...string) {
		if len(parts) > 0 && parts[0] != "" {
			dirs = append(dirs, filepath.Join(parts...))
		}
	}
	if root != "" {
		add(root, "node_modules", ".bin")
	}
	add(os.Getenv("GOBIN"))
	if gp := os.Getenv("GOPATH"); gp != "" {
		for _, p := range filepath.SplitList(gp) {
			add(p, "bin")
		}
	}
	add(home, "go", "bin")
	add(home, ".cargo", "bin")
	add(home, ".local", "bin")
	add(home, ".npm-global", "bin")
	add(home, ".bun", "bin")
	switch runtime.GOOS {
	case "windows":
		add(os.Getenv("APPDATA"), "npm")
		add(os.Getenv("LOCALAPPDATA"), "Programs", "Python", "Scripts")
		add(os.Getenv("ProgramFiles"), "Go", "bin")
		add(os.Getenv("ProgramFiles"), "nodejs")
		add(os.Getenv("ProgramFiles"), "LLVM", "bin")
	case "darwin":
		add("/opt/homebrew/bin")
		add("/usr/local/bin")
		add("/usr/local/go/bin")
	default:
		add("/usr/local/bin")
		add("/usr/local/go/bin")
		add("/snap/bin")
	}
	return dirs
}

// look finds a program on the PATH or in the usual install folders, and
// returns its full path.
func look(name, root string) string {
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	exts := []string{""}
	if runtime.GOOS == "windows" {
		exts = []string{".exe", ".cmd", ".bat", ""}
	}
	for _, d := range searchDirs(root) {
		for _, ext := range exts {
			p := filepath.Join(d, name+ext)
			if info, err := os.Stat(p); err == nil && !info.IsDir() {
				return p
			}
		}
	}
	return ""
}

// Find returns the first command whose program is installed, with the
// program's full path, or nil. root is the project folder, whose
// node_modules/.bin counts too.
func Find(root string, commands ...[]string) []string {
	for _, c := range commands {
		if len(c) == 0 {
			continue
		}
		if p := look(c[0], root); p != "" {
			return append([]string{p}, c[1:]...)
		}
	}
	return nil
}

// FindAll returns every command whose program is installed, in order,
// each with the program's full path.
func FindAll(root string, commands ...[]string) [][]string {
	var out [][]string
	for _, c := range commands {
		if len(c) == 0 {
			continue
		}
		if p := look(c[0], root); p != "" {
			out = append(out, append([]string{p}, c[1:]...))
		}
	}
	return out
}
