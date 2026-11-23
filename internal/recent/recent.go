// Package recent keeps the list of recently opened projects shown on the
// welcome screen and in the title bar's project widget.
package recent

import (
	"path/filepath"
	"slices"
	"strings"
	"time"

	"litecode/internal/config"
)

// MaxProjects is how many projects the list keeps.
const MaxProjects = 30

// Project is one entry of the recent projects list.
type Project struct {
	Path     string    `json:"path"`
	Name     string    `json:"name"`
	OpenedAt time.Time `json:"openedAt"`
}

// List is the recent projects, most recently opened first.
type List struct {
	Projects []Project `json:"projects"`
	path     string
}

// Load reads the list from the config directory.
func Load() (*List, error) {
	path, err := config.Path("recent-projects.json")
	if err != nil {
		return &List{}, err
	}
	l := &List{path: path}
	err = config.LoadJSON(path, l)
	return l, err
}

// Touch moves (or adds) a project to the top of the list and saves it.
func (l *List) Touch(path string, now time.Time) error {
	path = filepath.Clean(path)
	l.Projects = slices.DeleteFunc(l.Projects, func(p Project) bool { return p.Path == path })
	l.Projects = slices.Insert(l.Projects, 0, Project{Path: path, Name: filepath.Base(path), OpenedAt: now})
	if len(l.Projects) > MaxProjects {
		l.Projects = l.Projects[:MaxProjects]
	}
	return l.save()
}

// Remove drops a project from the list and saves it.
func (l *List) Remove(path string) error {
	l.Projects = slices.DeleteFunc(l.Projects, func(p Project) bool { return p.Path == path })
	return l.save()
}

// Filter returns the projects whose name or path contains query, case-insensitively.
func (l *List) Filter(query string) []Project {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return l.Projects
	}
	var out []Project
	for _, p := range l.Projects {
		if strings.Contains(strings.ToLower(p.Name), q) || strings.Contains(strings.ToLower(p.Path), q) {
			out = append(out, p)
		}
	}
	return out
}

func (l *List) save() error {
	if l.path == "" {
		return nil
	}
	return config.SaveJSON(l.path, l)
}
