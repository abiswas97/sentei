# Bubbles/help Adoption

## Why

Key/label knowledge currently lives in three independent sources: `keys.go` carries `key.WithHelp` text that nothing renders, ~20 `viewKeyHints` call sites re-spell key strings as literals, and `help.go` re-spells them a third time for the F1 portal, plus six render sites bypass `viewKeyHints` entirely with inline hint strings. Rebinding a key means grep-and-update across ~10 files, and hints can silently lie about bindings. The spec's own rule ("bindings are referenced, never redefined locally") is violated by every footer. Decided 2026-06-11: adopt `bubbles/help` fed from `keys.go` as the single source.

## What Changes

- `keys.go` becomes the single source for per-view key presentation: each view declares named key sections (`[]keySection`) whose bindings reuse shared key constructors with view-appropriate descriptions (the same physical key may mean "delete" in one view and "continue" in another, declared once per view, never at render sites).
- Footers render through `bubbles/help` (`ShortHelpView`) styled to the existing chrome contract: dim, ` · ` separators, 2-space pad; long hint rows truncate gracefully at narrow widths.
- `viewKeyHints`, the `KeyHint` type, and the hand-written tables in `help.go` are deleted; the F1 portal renders its named, grouped sections from the same `keySection` data.
- The six stray inline hint strings (`list.go` status bar, `repo_options.go`, `create_options.go`, `repo_summary.go` x2, `migrate_summary.go`) route through the new footer.
- The worktree list status bar finally advertises `?` and `F1` (the portal was undiscoverable from the one view that has details).
- `·` escapes disappear with their call sites.
- Visual output is unchanged except for the added `?`/`F1` hints on the list view.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `tui-design-system`: key-binding single-source requirement extends to per-view presentation data; render sites carry no raw key or label strings.
- `tui-chrome`: key-hints rendering requirement changes from the `viewKeyHints(KeyHint...)` API to binding-derived footers with identical visual format.
- `tui-list-view`: status bar requirement gains `?`/`F1` discoverability and derives hints from bindings.
- `tui-help`: contextual help content requirement derives from the same per-view key sections as the footers.

## Impact

- **Code**: `keys.go` (per-view sections), `chrome.go` (footer via bubbles/help), `help.go` (portal content from sections), every view file's footer line (~20 one-line call swaps plus the six stray sites).
- **Dependencies**: `charm.land/bubbles/v2/help` (already in the module).
- **Tests**: chrome/footers covered by existing render assertions updated to the new API; new table-driven tests for keymap-derived footers and portal sections; full-coverage rule applies to new helpers.
- **Out of scope**: the `viewFrame` chrome composer (separate refactor concern, queued).
