## 1. Selection Model Migration

- [x] 1.1 Change `m.selected` from `map[int]bool` to `map[string]bool` (keyed by worktree path)
- [x] 1.2 Update toggle (spacebar) handler to use worktree path as key
- [x] 1.3 Update select-all (`a`) handler to use worktree paths
- [x] 1.4 Update `selectedWorktrees()` to look up by path
- [x] 1.5 Update `viewList()` checkbox rendering to check path-based selection
- [x] 1.6 Update `viewConfirm()` if it references `m.selected` by index
- [x] 1.7 Verify existing tests pass after migration

## 2. Visible Indices and Reindex

- [x] 2.1 Add `visibleIndices []int` field to `Model`
- [x] 2.2 Add sort state fields: `sortField SortField`, `sortAscending bool` (with `SortField` enum: `SortByAge`, `SortByBranch`)
- [x] 2.3 Add filter state fields: `filterText string`, `filterActive bool`, `filterInput textinput.Model`
- [x] 2.4 Implement `reindex()` method: filter by branch substring (case-insensitive), then sort by current field/direction, store in `visibleIndices`, clamp cursor/offset
- [x] 2.5 Call `reindex()` in `NewModel` to set initial sorted order
- [x] 2.6 Update `viewList()` to iterate `m.visibleIndices` instead of `m.worktrees` directly
- [x] 2.7 Update cursor bounds checks to use `len(m.visibleIndices)`

## 3. Sorting

- [x] 3.1 Add `Sort` and `ReverseSort` key bindings to `keys.go` (`s` and `S`)
- [x] 3.2 Handle `s` in `updateList`: cycle sort field (age -> branch -> path -> age), call `reindex()`
- [x] 3.3 Handle `S` in `updateList`: toggle `sortAscending`, call `reindex()`
- [x] 3.4 Sort zero-value `LastCommitDate` to end of list regardless of direction
- [x] 3.5 Add sort indicator to status bar (e.g. `sort: age ^` or `sort: branch v`)

## 4. Filtering

- [x] 4.1 Initialize `textinput.Model` in `NewModel` with prompt `filter: ` and no placeholder
- [x] 4.2 Add `Filter` key binding to `keys.go` (`/`)
- [x] 4.3 Handle `/` in `updateList`: set `filterActive = true`, focus textinput
- [x] 4.4 Route key events to `textinput.Update` when `filterActive` is true
- [x] 4.5 Call `reindex()` after each textinput update (real-time filtering)
- [x] 4.6 Handle `esc` in filter mode: clear filter text, blur textinput, set `filterActive = false`, call `reindex()`
- [x] 4.7 Handle `enter` in filter mode: keep filter text applied, blur textinput, set `filterActive = false`
- [x] 4.8 Handle `esc` in list mode with applied filter: clear `filterText`, call `reindex()` (do not quit)
- [x] 4.9 Render filter input bar in place of status bar when `filterActive` is true
- [x] 4.10 Show filter state in status bar when filter is applied but not active (e.g. `filter: "feat" (3/10)`)
- [x] 4.11 Display "no matches" message when `visibleIndices` is empty due to filter

## 5. Status Bar Updates

- [x] 5.1 Add sort indicator to the status bar text
- [x] 5.2 Add key hints for new bindings: `/: filter`, `s: sort`
- [x] 5.3 Update `viewStatusBar()` to conditionally show filter info when filter is applied
- [x] 5.4 Ensure the `q` key always quits (unconditional, even with active filter)

## 6. Tests

- [x] 6.1 Test `reindex()` with various sort fields and directions
- [x] 6.2 Test `reindex()` with filter text (substring, case-insensitive, empty, no matches)
- [x] 6.3 Test combined sort + filter produces correct `visibleIndices`
- [x] 6.4 Test selection persistence across sort changes (path-based)
- [x] 6.5 Test select-all with active filter only toggles visible worktrees
- [x] 6.6 Test cursor clamping when filter reduces visible list
- [x] 6.7 Test zero-value `LastCommitDate` sorts to end
