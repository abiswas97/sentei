## Why

The worktree list is currently displayed in the order returned by `git worktree list --porcelain`, which is arbitrary. Users cannot sort by age to find the stalest worktrees first, nor can they filter a long list by branch name to quickly find specific worktrees. For repos with 10-30 worktrees, this makes the tool slower to use than it needs to be. PRD F7 calls for sorting and filtering as the first post-MVP feature.

## What Changes

- Add sorting of the worktree list by last activity date (default, oldest first) or branch name
- Cycle through sort fields with a key binding (`s`); reverse sort order with `S`
- Display current sort field and direction in the status bar
- Add inline text filtering activated by `/` (standard TUI convention, matches Vim and `bubbles/list`)
- Filter matches branch names using substring match (case-insensitive)
- While filtering, show a text input bar replacing the status bar; `esc` clears and exits filter mode
- Filtered-out worktrees are hidden from the list but not deselected (selections persist across filter changes)
- Navigation, selection, and deletion operate only on the visible (filtered + sorted) subset
- `a` (select all) toggles only visible worktrees when a filter is active
- Show filter state and match count in the status bar when a filter is applied

## Capabilities

### New Capabilities
- `worktree-sorting`: Sort the worktree list by configurable fields (age, branch) with toggleable direction
- `worktree-filtering`: Inline text filter to narrow the displayed worktree list by branch name

### Modified Capabilities
- `tui-list-view`: The list view must operate on a sorted/filtered subset rather than the raw worktree slice. Status bar gains sort indicator and filter state. Key bindings expand with `/`, `s`, `S`, `esc` (context-dependent).

## Impact

- **internal/tui/list.go**: Major changes — list rendering and navigation must use a filtered/sorted index mapping instead of direct `m.worktrees` indices. Status bar updated.
- **internal/tui/model.go**: New state fields for sort config, filter text, filter mode, and the computed visible index slice.
- **internal/tui/keys.go**: New key bindings for sort (`s`/`S`) and filter (`/`).
- **internal/tui/styles.go**: Possible new styles for filter input bar and sort indicator.
- **No new dependencies**: `bubbles/textinput` is already available via the bubbles dependency. Standard library `sort` and `strings` suffice for the rest.
- **Selection model**: `m.selected` currently maps `int` indices into `m.worktrees`. This must remain stable across sort/filter changes — selections should reference worktrees by identity (path), not by display position.
