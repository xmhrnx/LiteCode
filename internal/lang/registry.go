package lang

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// Plugin is what a plugin folder's plugin.json holds.
type Plugin struct {
	Name        string      `json:"name"`
	Version     string      `json:"version"`
	Description string      `json:"description"`
	Languages   []*Language `json:"languages"`
	Dir         string      `json:"-"`
}

// Registry is every language, plugins first so they can override built-ins.
type Registry struct {
	langs   []*Language
	Plugins []Plugin
	// Errors are plugins that couldn't be read; the rest still load.
	Errors []error
}

// NewRegistry loads the built-in languages and the plugins in pluginDir. A
// missing pluginDir is fine.
func NewRegistry(pluginDir string) *Registry {
	r := &Registry{}
	if pluginDir != "" {
		r.loadPlugins(pluginDir)
	}
	r.langs = append(r.langs, Builtin()...)
	return r
}

func (r *Registry) loadPlugins(dir string) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return
	}
	if err != nil {
		r.Errors = append(r.Errors, err)
		return
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p, err := readPlugin(filepath.Join(dir, e.Name()))
		if err != nil {
			r.Errors = append(r.Errors, err)
			continue
		}
		r.Plugins = append(r.Plugins, p)
		r.langs = append(r.langs, p.Languages...)
	}
}

func readPlugin(dir string) (Plugin, error) {
	var p Plugin
	data, err := os.ReadFile(filepath.Join(dir, "plugin.json"))
	if err != nil {
		return p, fmt.Errorf("plugin %s: %w", filepath.Base(dir), err)
	}
	if err := json.Unmarshal(data, &p); err != nil {
		return p, fmt.Errorf("plugin %s: plugin.json: %w", filepath.Base(dir), err)
	}
	if p.Name == "" {
		p.Name = filepath.Base(dir)
	}
	p.Dir = dir
	for _, l := range p.Languages {
		if l.ID == "" || l.Name == "" {
			return p, fmt.Errorf("plugin %s: every language needs an id and a name", p.Name)
		}
	}
	return p, nil
}

// ForPath returns the language of a file, PlainText when none claims it.
func (r *Registry) ForPath(path string) *Language {
	for _, l := range r.langs {
		if l.Matches(path) {
			return l
		}
	}
	return PlainText
}

// Languages lists every language, plugin ones first.
func (r *Registry) Languages() []*Language {
	return r.langs
}
