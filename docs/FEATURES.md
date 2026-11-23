# Research: what a full IDE offers, and what LiteCode keeps

Reference: IntelliJ Platform "New UI" (2024.x–2025.x, default Dark theme), default
keymap for Windows/Linux ("Windows" / "Default for KDE/GNOME" share these bindings).

## 1. Full feature inventory (JetBrains)

### Window / frame
- Custom-drawn frame ("merged main menu and toolbar"): hamburger main menu, **project
  widget** (project name + dropdown with recent projects, new/open/clone), **VCS widget**
  (current branch + branches popup), run configuration selector, Run/Debug/Stop buttons,
  Search Everywhere, Settings gear, window buttons.
- Compact mode, presentation mode, distraction-free / zen mode, full screen.
- Multiple project windows, "Attach" project to an open window.

### Tool windows
- Stripes on the left, right and bottom with icon buttons; each tool window can be moved
  to **Left / Right / Bottom** (top or bottom slot of the side stripe), resized by
  dragging its edge, hidden (Shift+Esc), maximized (Ctrl+Shift+'), hide all (Ctrl+Shift+F12).
- View modes: docked pinned, docked unpinned, undocked, float, window.
- Standard windows: Project (Alt+1), Bookmarks (Alt+2), Find (Alt+3), Run (Alt+4),
  Debug (Alt+5), Problems (Alt+6), Structure (Alt+7), Services (Alt+8), Git (Alt+9),
  Commit (Alt+0), Terminal (Alt+F12), Notifications, TODO, Database, Gradle/Maven/npm.

### Editor
- Tabs (pin, close others, close to the right, middle-click close, limit, placement),
  split right/down, drag tabs between splits, "Recent files" (Ctrl+E), "Recent locations"
  (Ctrl+Shift+E), Switcher (Ctrl+Tab).
- Gutter: line numbers, VCS change markers (click → inline diff/rollback), breakpoints,
  run icons, folding, Git blame (annotate).
- Syntax highlighting, semantic highlighting, inspections/quick-fixes (Alt+Enter),
  completion, parameter info, quick doc, go to declaration (Ctrl+B), find usages (Alt+F7),
  rename (Shift+F6), refactorings, live templates, multi-caret, column selection,
  extend selection (Ctrl+W), code folding, sticky lines, breadcrumbs, inlay hints,
  reformat (Ctrl+Alt+L), optimize imports, Prettier/ESLint integration (per project).
- Line ops: duplicate (Ctrl+D), delete (Ctrl+Y), move line (Alt+Shift+↑/↓), comment
  (Ctrl+/), block comment (Ctrl+Shift+/), join lines, go to line (Ctrl+G).
- Local history, clipboard history (Ctrl+Shift+V), scratch files.

### Navigation / search
- Search Everywhere (double Shift) with tabs All / Classes / Files / Symbols / Actions /
  Text; Go to File (Ctrl+Shift+N), Class (Ctrl+N), Symbol (Ctrl+Alt+Shift+N), Action
  (Ctrl+Shift+A); Find in Files (Ctrl+Shift+F) popup with live results + preview, Replace
  in Files (Ctrl+Shift+R); in-file Find (Ctrl+F) / Replace (Ctrl+R) bar; Select in
  (Alt+F1); Bookmarks.

### Git
- **Commit tool window**: change tree grouped by directory with checkboxes, Unversioned
  Files node, changelists, commit message editor, Amend, Commit / Commit and Push,
  pre-commit checks; double-click → diff.
- **Git tool window (Log)**: branches tree (HEAD, Local, Remote) with filter, commit list
  with graph lines and ref labels, filters (branch/user/date/path), commit details pane
  (message, hash, author, changed files tree), context actions (checkout revision,
  cherry-pick, revert, reset, new branch/tag here, interactive rebase).
- **Branches popup** (title-bar widget / Ctrl+Shift+`): search, New Branch, Checkout Tag
  or Revision, Recent, Local, Remote; per-branch Checkout / Merge / Rebase / Push /
  Rename / Delete / Compare.
- **Diff viewer**: side-by-side and unified, word-level highlighting, editable right side,
  navigate changes (F7 / Shift+F7), ignore whitespace.
- Push dialog (Ctrl+Shift+K) listing outgoing commits; Update Project (Ctrl+T); fetch;
  stash/shelve; three-way merge conflict resolver; annotate/blame; file history; rollback
  (Ctrl+Alt+Z); GitHub/GitLab PR integration.

### Other
- Welcome screen: recent projects list (searchable, groups, remove from list), New
  Project, Open, Get from VCS (clone), plugins, customize.
- Settings (Ctrl+Alt+S): Appearance (theme, UI font), Keymap (search by action or
  shortcut, assign), Editor (font, size, line spacing, color scheme), Plugins
  (marketplace), Version Control, Languages & Frameworks, Tools.
- Run configurations, debugger, profiler, test runner, coverage, database tools, HTTP
  client, remote dev, code-with-me, AI assistant, indexing (the main reason it is heavy).

## 2. What LiteCode keeps (the main ones)

LiteCode keeps the general feel (dockable panels, keyboard-first, strong Git UI) with its own naming and visual design: Files / Changes / History panels, Search Anything, Run Action. Picked from your list (look, Git UI, movable terminal, font, colours, keybindings, recent
projects menu, multiple open files), plus what those depend on and the two search
features you mentioned.

| Area | Kept | Dropped / later |
| --- | --- | --- |
| Frame | Custom title bar: main menu (☰), project widget with **recent projects**, branch widget, Search, Settings, window buttons | Run configs, run/debug (later) |
| Look | LiteCode's own design (not a copy): cool dark neutrals, violet accent, hairline separators, Lucide icons tinted per language, Inter 13 for UI, JetBrains Mono 13 for code | Light theme (palette is ready, colours later) |
| Tool windows | Left / Right / Bottom docks, stripes, **Move to → Left/Right/Bottom** on any tool window incl. Terminal, drag-to-resize, Shift+Esc hide, Ctrl+Shift+F12 hide all, layout saved per user | Float/window modes |
| Editor | LiteCode's own code editor: syntax highlighting for the built-in languages (plugins add keywords), line-number gutter, current line, indent guides, selection by mouse (click, drag, double-click word, triple-click or gutter click line), smart Home, word moves, auto-closing brackets and quotes, smart Enter and Backspace, Tab/Shift+Tab indent, undo/redo with typing merged into steps, copy/cut a whole line with nothing selected, input-method composition. Multiple files in tabs with file icons, Ctrl+Tab / Alt+←/→, Ctrl+E recent files, line ops (Ctrl+D, Ctrl+Y, Ctrl+/, Alt+Shift+↑/↓), Ctrl+G, Ctrl+S, status bar | Find/replace in file, completion, folding, multi-caret, VCS gutter markers, horizontal wheel scrolling |
| Formatting | **Global** "format on save" with Prettier (project `node_modules/.bin/prettier`, else PATH) and per-language formatters (gofmt, …) from plugins — no per-project setup | ESLint |
| Navigation | Search Everywhere (double Shift), Go to File (Ctrl+Shift+N), Find in Files (Ctrl+Shift+F) with results + preview, Find Action (Ctrl+Shift+A) | Symbols/classes (needs language servers) |
| Git | Commit window (tree, checkboxes, unversioned, message, amend, Commit / Commit and Push, rollback), Git log window (branches tree, commit list with ref labels, details + changed files), side-by-side diff, branches popup (new, checkout, merge, delete, push, update/fetch), Ctrl+K / Ctrl+Shift+K / Ctrl+T | Graph lines, interactive rebase, merge resolver, blame (later) |
| Keymap | Ctrl-based defaults familiar from IntelliJ-style editors, user overrides in `keymap.json`, editable in Settings | Alternate keymaps (Mac, VS Code) — later as files |
| Projects | Welcome screen with searchable recent projects, Open Folder, Clone | New-project wizards |
| Plugins | Declarative language plugins (`plugin.json`): extensions, comment tokens, indent, formatter command, LSP command (reserved), run command | Marketplace |
| Terminal | The user's shell in the project folder (Giode Terminal: colors, alternate screen, scrollback, selection, mouse), themed to match, restart button, movable like any other panel | Several sessions in tabs |

## 3. Why it should be light

- Native Go binary on Gio: no JVM, no indexer, no background inspections.
- Nothing runs until used: the file tree loads a directory when it's expanded; git runs
  `git` on demand in a goroutine; search walks files only when asked.
- Frames are drawn only when something changes (Gio is event driven), so an idle IDE uses
  no CPU.
