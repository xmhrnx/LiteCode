// Package app is the root of the UI: the welcome screen or the open project's
// workbench, and the wiring between their parts.
package app

import (
	"litecode/internal/layout"
	"litecode/ui/assistant"
	"litecode/ui/gitui"
	"litecode/ui/projectview"
	"litecode/ui/terminal"
)

// defaultPlacements is where tool windows start, as in the IDE: Project and
// Commit on the left, Terminal and Git at the bottom.
var defaultPlacements = map[layout.WindowID]layout.Placement{
	projectview.WindowID: {Side: layout.Left, Order: 0},
	gitui.CommitWindowID: {Side: layout.Left, Order: 1},
	terminal.WindowID:    {Side: layout.Bottom, Order: 0},
	gitui.LogWindowID:    {Side: layout.Bottom, Order: 1},
	assistant.WindowID:   {Side: layout.Right, Order: 0},
}
