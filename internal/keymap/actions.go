// Package keymap maps IDE actions to key strokes. The default is the
// Ctrl-based default keymap; on macOS "Mod" becomes Command.
// Users override bindings in keymap.json in the config directory.
package keymap

// Action identifies something the IDE can do from a key stroke or a menu.
type Action string

// The actions with default bindings.
const (
	SearchEverywhere   Action = "search.anything"
	GotoFile           Action = "search.file"
	GotoAction         Action = "search.action"
	FindInPath         Action = "search.inFiles"
	RecentFiles        Action = "search.recentFiles"
	GotoLine           Action = "editor.goToLine"
	Switcher           Action = "tabs.switch"
	SwitcherBackward   Action = "tabs.switchBack"
	NextTab            Action = "tabs.next"
	PreviousTab        Action = "tabs.previous"
	CloseEditor        Action = "tabs.close"
	SaveAll            Action = "file.saveAll"
	ReformatCode       Action = "editor.format"
	DuplicateLine      Action = "editor.duplicateLine"
	DeleteLine         Action = "editor.deleteLine"
	CommentLine        Action = "editor.toggleComment"
	MoveLineUp         Action = "editor.moveLineUp"
	MoveLineDown       Action = "editor.moveLineDown"
	ShowSettings       Action = "app.settings"
	OpenProject        Action = "app.openFolder"
	ProjectWindow      Action = "panel.files"
	CommitWindow       Action = "panel.changes"
	GitWindow          Action = "panel.history"
	TerminalWindow     Action = "panel.terminal"
	AssistantWindow    Action = "panel.assistant"
	HideToolWindow     Action = "panel.hide"
	HideAllWindows     Action = "panel.toggleAll"
	CommitAction       Action = "git.commit"
	PushAction         Action = "git.push"
	UpdateProject      Action = "git.pull"
	Branches           Action = "git.branches"
	Rollback           Action = "git.discard"
	NextDifference     Action = "diff.next"
	PreviousDifference Action = "diff.previous"
)

// Names are the labels shown in Settings > Keymap and Run Action.
var Names = map[Action]string{
	SearchEverywhere: "Search Anything", GotoFile: "Go to File…", GotoAction: "Run Action…",
	FindInPath: "Find in Files…", RecentFiles: "Recent Files", GotoLine: "Go to Line…",
	Switcher: "Switch Tab", SwitcherBackward: "Switch Tab Backward", NextTab: "Next Tab",
	PreviousTab: "Previous Tab", CloseEditor: "Close Tab", SaveAll: "Save All",
	ReformatCode: "Format Document", DuplicateLine: "Duplicate Line",
	DeleteLine: "Delete Line", CommentLine: "Toggle Line Comment", MoveLineUp: "Move Line Up",
	MoveLineDown: "Move Line Down", ShowSettings: "Settings…", OpenProject: "Open Folder…",
	ProjectWindow: "Files", CommitWindow: "Changes", GitWindow: "History", TerminalWindow: "Terminal", AssistantWindow: "Assistant",
	HideToolWindow: "Hide Panel", HideAllWindows: "Toggle All Panels",
	CommitAction: "Commit…", PushAction: "Push…", UpdateProject: "Pull",
	Branches: "Branches…", Rollback: "Discard Changes…", NextDifference: "Next Change",
	PreviousDifference: "Previous Change",
}
