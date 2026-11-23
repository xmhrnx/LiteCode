// Package config keeps LiteCode's own files (settings, recent projects, layouts,
// per-project sessions) in the user's config directory, so projects are never
// polluted with IDE files.
package config

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// AppDirName is the folder LiteCode uses under the user's config directory.
const AppDirName = "litecode"

// Dir returns LiteCode's config directory, creating it when missing.
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, AppDirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// Path returns the path of a file inside the config directory.
func Path(elem ...string) (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(append([]string{dir}, elem...)...), nil
}

// ProjectKey turns a project path into a stable file-name-safe key.
func ProjectKey(projectPath string) string {
	sum := sha1.Sum([]byte(filepath.Clean(projectPath)))
	return filepath.Base(projectPath) + "-" + hex.EncodeToString(sum[:6])
}

// LoadJSON reads path into v. A missing file leaves v untouched and is not an error.
func LoadJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// SaveJSON writes v to path atomically (write to a temp file, then rename).
func SaveJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
