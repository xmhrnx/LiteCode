package project

import (
	"os"
	"strings"
)

// ShortPath writes a path under the home folder as "~/…".
func ShortPath(path string) string {
	if home, err := os.UserHomeDir(); err == nil && home != "" && strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}
