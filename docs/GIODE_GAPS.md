# What Giode is missing for LiteCode

Rechecked against v1.0.27-alpha.1: none of the items below is fixed in that release.

Found while building LiteCode on Giode v1.0.26-alpha.1. Per your rule, none of
these are worked around in LiteCode; each says what LiteCode does today and what it
needs. Ordered by impact.

## Blocking (core IDE features wait on these)

### 1. CodeEditor component
`Input.Text().Multiline(true)` is the editor today. An IDE needs a dedicated
code editor:
- **Styled spans per range** (syntax highlighting, search-hit and diff
  backgrounds), applied without re-shaping the whole document.
- **Gutter**: line numbers, VCS change markers, fold arrows, click targets,
  scrolling in lockstep with the text.
- **No soft wrap** option plus horizontal scrolling (code lines must not wrap).
- **Tab rendering**: tab characters currently draw as a missing-glyph box.
  `Styles.TabWidth` only affects `Text`/`RichText`, not `Input`. (Every Go file
  is affected.)
- Caret/selection **change callback** (`OnCaretMove`) and line/column queries,
  so the status bar doesn't recompute from the whole text each frame.
- Edit API by range (`Replace(start, end, text)`) that's one undo step.
  LiteCode uses `SetSelection` + `InsertText` for this.
- Multi-caret / column selection, current-line highlight, matching-bracket
  highlight, scroll-to-line with centering, visible-line range.
- Large-file performance (rope/piece table instead of one string).
- Hooks for completion popups and inline hints anchored at a text position
  (`PositionOf(offset) image.Point`).

### 2. Terminal component
The Terminal panel docks and moves to any side, but has nothing inside:
- PTY (Unix `openpty`/`forkpty`, Windows ConPTY), shell launch with env/cwd.
- VT100/xterm escape parser (colors, cursor movement, alternate screen, mouse).
- Cell-grid renderer in a monospace font, scrollback, selection and copy,
  paste, resize → `TIOCSWINSZ`.

### 3. Key observer
`giode.OnKey` and `Shortcuts` only get keys nothing else took, so the app
can't see keys a focused text field consumed. Double-Shift (Search Anything) needs to
know whether any other key was pressed between the two Shifts. LiteCode
approximates by watching text changes (`ide.TypingRev`). Needed: an
`OnKeyObserve(fn)` that sees every key event without consuming it.

## Visual fidelity

4. **SVG `fill-rule="evenodd"`** is ignored by the icon renderer. LiteCode's
   Lucide icons are stroke-based and unaffected, but fill-based sets with
   holes (Material Symbols, many brand icons) render with their holes filled.
5. **Menu border and shadow**: `Menu` only takes Background, BorderRadius,
   Padding and Color. LiteCode's popups have a 1px border and a shadow; menus
   can't, so over a same-colored panel they disappear and LiteCode gives them a
   lighter background instead.
6. **Resize cursors**: `Styles.Cursor` has no `col-resize` / `row-resize`
   (DataGrid uses Gio's internally). Splitters show the grab cursor.
7. **`RichText` ignores `MaxLines` / `TextOverflow`**: long Find-in-Files
   lines wrap; LiteCode clips the row.
8. **`Text` ignores `Width` inside some rows** (settings labels): LiteCode wraps
   them in a fixed-width `Box`.
9. **DocTabs**: no per-tab icon (file-type icons in editor tabs), no pinned
   tabs, no style for the selected tab's underline color.
10. **Dialog** draws no border and accepts no top offset (command-style
    popups like Search Anything sit high, not centered). LiteCode built its own popup layer
    for the top-anchored popups.

## API ergonomics

11. **Palette is fixed**: an IDE needs ~40 named colors (diff, VCS status,
    selection, …). Custom `theme.*` tokens (`SetToken("diff.inserted", …)`)
    would let them follow light/dark like the built-in ones.
12. **Handle types not re-exported** at the root: `Tooltip`, `Scroll`,
    `Dialog`, `Input`, `Checkbox`, `Select`, `TitleBar`, `IconButton`, plus
    the typed style values (`Width`, `Height`, `FontSize`, `LineHeight`,
    `Top`, `BorderRadius`). LiteCode imports `components/*` and
    `styles/properties` for them.
13. **`Pressable.OnClick` fires for every mouse button**: a right-click on a
    stripe button both toggled the window and opened its menu. An
    `OnPrimaryClick` (or a button filter) would help; LiteCode filters by
    `e.Button`.
14. **`Menu.Trigger` around a clickable element** never opens: the inner
    pressable takes the click. LiteCode opens menus with `OpenAtClick` instead.
15. **Menus opened with `OpenAtClick` must also be laid out** somewhere to
    paint; easy to miss for menus built on demand. LiteCode has a `MenuSlot`.
16. **File-name build tags**: a component file named `tool_windows.go` is
    Windows-only to Go. Worth a line in the one-component-per-file guidance.

## Nice to have
17. A `SplitPane` component (LiteCode's splitter is a `Pressable` with drag
    events, which works fine).
18. Window-focus callback (`OnFocusChange`) instead of polling `App.Focused()`
    each frame.

## Found while adding the AI Assistant (v1.0.27)

19. **First click in a Dialog opened from a keyboard shortcut is lost** while a text
    field (the editor) had focus: the press seems to go to moving focus out of the
    field. Opened with the mouse, the first click works. Seen with Settings
    (Ctrl+Alt+S); root cause not confirmed.
20. **Key observer** (item 3) matters more now: every text field has to report its edits
    (`widgets.TextField` / `widgets.Typing`) so double Shift doesn't fire while
    typing capitals. A `Combobox`'s inner field can't be wrapped that way.
