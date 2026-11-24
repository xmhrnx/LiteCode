package theme

import (
	"path"
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"

	"litecode/assets"
)

// fileKind is how a file type is drawn: which icon, in which color.
type fileKind struct {
	icon  string
	color string
}

// LoadIcons registers the embedded icons by file name: "search", "branch", …
func LoadIcons() error {
	return giode.RegisterIconsFS(assets.Icons, "icons")
}

// Icon draws an icon at 16dp in one color.
func Icon(name, color string) giode.Element {
	return IconSized(name, color, 16)
}

// IconSized draws an icon at size dp.
func IconSized(name, color string, size int) giode.Element {
	return giode.Icon(name, giode.Styles{Width: properties.Width(size), Height: properties.Height(size), Color: color})
}

// Language colors for file icons.
const (
	tintCode   = "#74B3FF"
	tintGo     = "#4FD1E8"
	tintJS     = "#F5D05B"
	tintTS     = "#5C9DFF"
	tintData   = "#F5A65B"
	tintStyle  = "#C38BFF"
	tintMarkup = "#F2845C"
	tintDocs   = "#9FB4D6"
	tintConfig = "#979BA7"
	tintShell  = "#5CCF8F"
	tintImage  = "#4FD1B8"
	tintRust   = "#E8936B"
	tintPython = "#7FB2F0"
)

// fileKinds maps extensions to how they're drawn.
var fileKinds = map[string]fileKind{
	".go": {"file-code", tintGo}, ".mod": {"file-config", tintGo}, ".sum": {"file-lock", tintConfig},
	".js": {"file-code", tintJS}, ".jsx": {"file-code", tintJS}, ".mjs": {"file-code", tintJS}, ".cjs": {"file-code", tintJS},
	".ts": {"file-code", tintTS}, ".tsx": {"file-code", tintTS}, ".vue": {"file-code", tintShell}, ".svelte": {"file-code", tintMarkup},
	".json": {"file-data", tintData}, ".jsonc": {"file-data", tintData}, ".yml": {"file-config", tintStyle}, ".yaml": {"file-config", tintStyle},
	".toml": {"file-config", tintData}, ".xml": {"file-code", tintMarkup}, ".csv": {"file-data", tintShell}, ".sql": {"file-data", tintTS},
	".css": {"file-code", tintStyle}, ".scss": {"file-code", tintStyle}, ".less": {"file-code", tintStyle},
	".html": {"file-code", tintMarkup}, ".htm": {"file-code", tintMarkup},
	".md": {"file-text", tintDocs}, ".txt": {"file-text", tintConfig}, ".rst": {"file-text", tintDocs},
	".sh": {"file-shell", tintShell}, ".bash": {"file-shell", tintShell}, ".zsh": {"file-shell", tintShell}, ".ps1": {"file-shell", tintTS},
	".py": {"file-code", tintPython}, ".rs": {"file-code", tintRust}, ".java": {"file-code", tintRust}, ".kt": {"file-code", tintStyle},
	".c": {"file-code", tintCode}, ".h": {"file-code", tintStyle}, ".cpp": {"file-code", tintCode}, ".hpp": {"file-code", tintStyle},
	".png": {"file-image", tintImage}, ".jpg": {"file-image", tintImage}, ".jpeg": {"file-image", tintImage}, ".gif": {"file-image", tintImage},
	".svg": {"file-image", tintImage}, ".webp": {"file-image", tintImage}, ".ico": {"file-image", tintImage},
	".zip": {"file-archive", tintData}, ".gz": {"file-archive", tintData}, ".tar": {"file-archive", tintData},
	".ttf": {"file-font", tintConfig}, ".otf": {"file-font", tintConfig}, ".woff": {"file-font", tintConfig}, ".woff2": {"file-font", tintConfig},
	".lock": {"file-lock", tintConfig}, ".env": {"file-config", tintConfig}, ".ini": {"file-config", tintConfig}, ".conf": {"file-config", tintConfig},
}

// FileIcon draws the icon of a file by its name, or a folder.
func FileIcon(name string, dir bool) giode.Element {
	if dir {
		return Icon("folder", Current().TextMuted)
	}
	lower := strings.ToLower(name)
	switch {
	case lower == "dockerfile" || strings.HasPrefix(lower, "dockerfile.") || strings.HasPrefix(lower, "docker-compose"):
		return Icon("file-docker", tintTS)
	case lower == "makefile":
		return Icon("file-shell", tintData)
	case strings.HasPrefix(lower, "."):
		return Icon("file-config", tintConfig)
	}
	if k, ok := fileKinds[path.Ext(lower)]; ok {
		return Icon(k.icon, k.color)
	}
	return Icon("file", tintConfig)
}
