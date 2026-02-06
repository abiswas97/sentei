## Why

Users can accidentally select and delete worktrees for important long-lived branches (main, master, develop, dev). These branches are almost never candidates for cleanup and selecting them is a costly mistake. Convention-based protection eliminates this risk with zero configuration.

## What Changes

- Protected worktrees (matching `main`, `master`, `develop`, `dev`) are visible in the list but cannot be selected for deletion
- Spacebar and select-all (`a`) skip protected worktrees
- Protected worktrees display a `[P]` indicator instead of a checkbox
- No configuration file needed — protection is built-in by convention

## Capabilities

### New Capabilities
- `branch-protection`: Convention-based protection preventing selection/deletion of well-known branches (main, master, develop, dev)

### Modified Capabilities
- `tui-list-view`: Protected worktrees render differently (no checkbox, `[P]` indicator) and cannot be selected
- `worktree-filtering`: Select-all must skip protected worktrees

## Impact

- `internal/tui/list.go`: Rendering changes for protected rows
- `internal/tui/model.go`: Selection logic must check protection status
- `internal/git/worktree.go`: May need `IsProtected` field or a helper function
- Dry run output should indicate protected status
