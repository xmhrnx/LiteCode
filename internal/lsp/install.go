package lsp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"litecode/internal/proc"
)

// Installer installs a language's server with a tool the user has.
type Installer struct {
	Server string   // the server's name, for messages
	Tool   string   // the program that installs it: go, npm, rustup, pip
	Args   []string // its arguments
}

// Command is the install command as it would be typed.
func (in Installer) Command() string {
	return in.Tool + " " + strings.Join(in.Args, " ")
}

// installers are the ways to install each language's server, best first.
var installers = map[string][]Installer{
	"go":         {{"gopls", "go", []string{"install", "golang.org/x/tools/gopls@latest"}}},
	"gomod":      {{"gopls", "go", []string{"install", "golang.org/x/tools/gopls@latest"}}},
	// TypeScript 7 serves the language itself (tsc --lsp).
	"typescript": {{"TypeScript", "npm", []string{"install", "-g", "typescript@latest"}}},
	"javascript": {{"TypeScript", "npm", []string{"install", "-g", "typescript@latest"}}},
	"python": {
		{"pyright", "npm", []string{"install", "-g", "pyright"}},
		{"pylsp", "pip", []string{"install", "--user", "python-lsp-server"}},
		{"pylsp", "pip3", []string{"install", "--user", "python-lsp-server"}},
	},
	"rust":       {{"rust-analyzer", "rustup", []string{"component", "add", "rust-analyzer"}}},
	"css":        {{"vscode-css-language-server", "npm", []string{"install", "-g", "vscode-langservers-extracted"}}},
	"html":       {{"vscode-html-language-server", "npm", []string{"install", "-g", "vscode-langservers-extracted"}}},
	"json":       {{"vscode-json-language-server", "npm", []string{"install", "-g", "vscode-langservers-extracted"}}},
	"yaml":       {{"yaml-language-server", "npm", []string{"install", "-g", "yaml-language-server"}}},
	"shell":      {{"bash-language-server", "npm", []string{"install", "-g", "bash-language-server"}}},
	"dockerfile": {{"docker-langserver", "npm", []string{"install", "-g", "dockerfile-language-server-nodejs"}}},
}

// InstallerFor returns how to install the language's server with a tool
// that's installed, or false when there's none.
func InstallerFor(languageID, root string) (Installer, bool) {
	for _, in := range installers[languageID] {
		if p := look(in.Tool, root); p != "" {
			in.Tool = p
			return in, true
		}
	}
	return Installer{}, false
}

// Install runs an installer, returning its output when it fails.
func Install(ctx context.Context, in Installer) error {
	cmd := proc.Command(ctx, in.Tool, in.Args...)
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(out.String())
		if len(msg) > 400 {
			msg = "…" + msg[len(msg)-400:]
		}
		if msg == "" {
			return fmt.Errorf("installing %s failed: %w", in.Server, err)
		}
		return errors.New("installing " + in.Server + " failed: " + msg)
	}
	return nil
}
