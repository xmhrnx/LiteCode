package workspace

import "litecode/internal/config"

// Session is what was open when a project was closed.
type Session struct {
	OpenFiles []string `json:"openFiles"`
	Selected  string   `json:"selected"`
	Pinned    []string `json:"pinned,omitempty"`
}

// LoadSession reads a project's session.
func LoadSession(projectPath string) (Session, error) {
	var s Session
	path, err := config.Path("sessions", config.ProjectKey(projectPath)+".json")
	if err != nil {
		return s, err
	}
	err = config.LoadJSON(path, &s)
	return s, err
}

// SaveSession writes a project's session.
func SaveSession(projectPath string, s Session) error {
	path, err := config.Path("sessions", config.ProjectKey(projectPath)+".json")
	if err != nil {
		return err
	}
	return config.SaveJSON(path, s)
}
