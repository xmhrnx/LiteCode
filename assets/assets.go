// Package assets embeds LiteCode's fonts and icons.
package assets

import "embed"

// Fonts holds Inter (UI) and JetBrains Mono (code), both under the SIL OFL.
//
//go:embed fonts/*.ttf
var Fonts embed.FS

// AppIcon is the window and taskbar icon, 256 px.
//
//go:embed icon-256.png
var AppIcon []byte

// Icons holds Lucide icons (ISC license), drawn in one color that follows
// the theme.
//
//go:embed icons/*.svg
var Icons embed.FS
