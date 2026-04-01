## Why

After a mutation (worktree removal, creation, integration apply), the first keypress on the main menu is swallowed because a lazy `stateStale` gate in `updateMenu` consumes the entire Bubble Tea update cycle to fire an async worktree reload, instead of processing the user's input. The worktree count also appears stale until the reload completes. This creates a perceptible input lag that feels broken.

## What Changes

- Fire `loadWorktreeContext` eagerly at mutation completion time (like React Query's `onSuccess`), not lazily on menu re-entry
- Handle `worktreeContextMsg` globally in `Update()` before view-specific dispatch, so the response is processed regardless of which view is active
- Add a generation token to `worktreeContextMsg` to guard against out-of-order responses overwriting fresher data
- Remove the `stateStale` field and the lazy-reload gate from `updateMenu`
- Remove the `worktreeContextMsg` handler from `updateMenu` (moved to global scope, would become dead code)
- Only fire eager reload from mutation sites that return to the menu (removal, creation, integration apply); skip flows that exit with `tea.Quit` (repo create/clone, standalone cleanup)

## Capabilities

### New Capabilities

- `eager-state-refresh`: Mechanism for eagerly refreshing worktree state on mutation success, with generation-token guards for safe out-of-order message handling

### Modified Capabilities

- `tui-progress`: The progress/summary flow now fires a reload command at mutation completion instead of setting a stale flag

## Impact

- `internal/tui/model.go`: Remove `stateStale` field, add `worktreeGeneration` field, add global `worktreeContextMsg` handler in `Update()`
- `internal/tui/menu.go`: Remove `stateStale` gate and `worktreeContextMsg` case from `updateMenu`, add generation to `worktreeContextMsg` and `loadWorktreeContext`
- `internal/tui/progress.go`: Replace `stateStale = true` with eager reload via `tea.Batch`
- `internal/tui/create_progress.go`: Same
- `internal/tui/integration_progress.go`: Same (conditional on `returnView != migrateNextView`)
- `internal/tui/repo_progress.go`: Remove `stateStale = true` (flow exits with `tea.Quit`)
- `internal/tui/cleanup_result.go`: Remove `stateStale = true` (flow exits with `tea.Quit`)
- Existing tests for progress, menu, and summary flows will need updates
