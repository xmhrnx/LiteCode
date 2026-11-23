package toolwin

import (
	"litecode/internal/keymap"
	"litecode/internal/layout"
)

// panelActions are the actions that toggle each panel, for tooltips.
var panelActions = map[layout.WindowID]keymap.Action{
	"files":     keymap.ProjectWindow,
	"changes":   keymap.CommitWindow,
	"history":   keymap.GitWindow,
	"terminal":  keymap.TerminalWindow,
	"assistant": keymap.AssistantWindow,
}
