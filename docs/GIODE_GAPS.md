# What Giode is missing for LiteCode

Checked against **v1.0.28-alpha.1**. Per the project rule, nothing here is worked
around in LiteCode; each item says what LiteCode does today and what it needs.

## Bugs found with v1.0.28 (new)

### 1. `KeyboardInput` delivers typed text after the keys typed with it
In `components/input/keyboardinput.go`, `update` passes each `key.Event` to `OnKey`
inside the event loop, but typed text reaches `OnText` only from `commit`, after
the loop. When text and a key arrive in the same frame, the key goes first. Typing
`foo(` then Enter quickly gives `fo` ⏎ `o(` in the editor, and `ls --color` ⏎ in
`giode.Terminal` runs `ls --colo` and leaves `r` on the next prompt. Any fast typist
hits this.
Fix: flush the committed text (outside a composition) before handing a key to
`k.key(e)`, e.g. call `k.commit(false)` at the top of the `key.Event` case, so
`OnText` and `OnKey` see input in the order it was typed.

### 2. A key taken by an `OnKey`/`Shortcuts` handler doesn't redraw
`DispatchKeys` runs after layout. When a handler takes the key and changes state
(a keymap action such as Duplicate Line), no new frame is requested, so the change
shows only on the next input event (a mouse move). Fix: invalidate when a
dispatched handler returns true. (LiteCode's editor asks for a frame after its own
`KeyboardInput` handlers, because its status bar is laid out before it; that part
is ordinary app code.)

### 3. `Scroll` doesn't expose its viewport width
`Scroll` keeps `viewportW` but has only `ViewportHeight()`. A code editor in a
`Vertical|Horizontal` scroll needs the width to size its content to at least
the viewport and to scroll the caret into view sideways. LiteCode's editor
scrolls sideways only by following the caret, with no horizontal wheel or
scrollbar. Needed: `ViewportWidth()` (and `ContentWidth()`).

### 4. `Pressable` has no press-down event with a position
Editors place the caret when the button goes down. `OnClick` fires on release,
and `OnDragStart` after 6dp of movement. LiteCode places the caret on release,
and drag selection starts once the drag does, from the press position. Needed:
`OnPress(fn func(ClickEvent))`, and a way to set the drag threshold to 0 for
text selection.

## Resolved in v1.0.28 (now used by LiteCode)

- **Keyboard input for app-drawn content**: `KeyboardInput` is the code editor's
  input. With canvases and `Scroll`, LiteCode has its own editor
  (`ui/editor/text_editor*.go`, `internal/textbuf`, `internal/highlight`).
- **Terminal**: `giode.Terminal()` runs the user's shell in the Terminal panel.
- **Key observer**: `WatchKeys` drives double Shift. Text typed into the editor
  still counts as typing (Giode reports it as text, not keys), so two capitals
  don't trigger Search Anything.
- **Menu border and shadow**: menus now use the popups' edge and shadow.
- **Resize cursors**: splitters show `col-resize` / `row-resize`.
- **DocTab icons**: editor tabs show the file-type icon; diff tabs a diff icon.
- **SVG `fill-rule="evenodd"`**, **`Dialog.Top`**, **custom `theme.*` tokens**:
  available. LiteCode's top-anchored popups and its own color scheme predate them
  and haven't moved over yet.

## Still open from earlier (not rechecked in v1.0.28)

5. **First click in a Dialog opened from a keyboard shortcut is lost** while a text
   field had focus. Opened with the mouse, the first click works.
6. **`RichText` ignores `MaxLines` / `TextOverflow`**: long Find-in-Files lines
   wrap; LiteCode clips the row.
7. **`Text` ignores `Width` inside some rows** (settings labels): LiteCode wraps
   them in a fixed-width `Box`.
8. **DocTabs**: no pinned tabs, no style for the selected tab's underline color.
9. **Handle types not re-exported** at the root: `Tooltip`, `Scroll`, `Dialog`,
   `Input`, `Checkbox`, `Select`, `TitleBar`, `IconButton`.
10. **`Pressable.OnClick` fires for every mouse button**: LiteCode filters by
    `e.Button`. An `OnPrimaryClick` would help.
11. **`Menu.Trigger` around a clickable element** never opens; LiteCode uses
    `OpenAtClick`.
12. **Menus opened with `OpenAtClick` must also be laid out** to paint; LiteCode
    has a `MenuSlot`.
13. **File-name build tags**: a component file named `tool_windows.go` is
    Windows-only to Go. Worth a line in the one-component-per-file guidance.

## Nice to have

14. A `SplitPane` component (LiteCode's splitter is a `Pressable` with drag events,
    which works fine).
15. Window-focus callback (`OnFocusChange`) instead of polling `App.Focused()`.
