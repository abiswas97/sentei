## Why

After deleting worktrees, orphaned metadata can remain in `$GIT_DIR/worktrees`. Currently, the summary view only *suggests* the user run `git worktree prune` manually. This is friction that should be automated — the tool already knows deletions just happened and has a `CommandRunner` available.

## What Changes

- Automatically run `git worktree prune` after all deletions complete, before displaying the summary
- Replace the "Tip: run `git worktree prune`" text with a confirmation that pruning was performed (or a warning if it failed)
- Add a new `PruneWorktrees` function to the worktree package that wraps the git command
- The prune step runs once (not per-worktree), is non-blocking to the exit flow, and its result is shown in the summary

## Capabilities

### New Capabilities
- `worktree-prune`: Automatic post-deletion pruning of orphaned worktree metadata via `git worktree prune`

### Modified Capabilities
- `tui-progress`: The transition from progress → summary now includes a prune step; summary view shows prune result instead of a manual tip

## Impact

- `internal/worktree/` — new `PruneWorktrees` function
- `internal/tui/progress.go` — trigger prune after `allDeletionsCompleteMsg`
- `internal/tui/summary.go` — display prune result instead of manual tip
- `internal/tui/model.go` — add prune result field to `Model`
