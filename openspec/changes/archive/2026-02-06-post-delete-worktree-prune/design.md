## Context

After worktree deletions, `$GIT_DIR/worktrees` can retain orphaned metadata entries. The current TUI summary displays a text tip suggesting users run `git worktree prune` manually. Since sentei already has a `CommandRunner` and the repo path, it can automate this step.

The prune operation is a single `git worktree prune` call against the bare repo — fast, idempotent, and safe. It only removes metadata for worktrees whose directories no longer exist.

## Goals / Non-Goals

**Goals:**
- Automatically run `git worktree prune` after all deletions complete
- Show prune outcome in the summary (success or failure)
- Keep the prune step non-blocking — a prune failure does not prevent the user from seeing deletion results or exiting

**Non-Goals:**
- Running `git gc` or other maintenance commands — prune is sufficient for worktree cleanup
- Adding `--verbose` or `--dry-run` flags for the prune step — it runs silently in the background
- Making prune optional or configurable — it should always run after deletions

## Decisions

### D1: Prune as a synchronous step between progress and summary

Run prune after `allDeletionsCompleteMsg` arrives but before transitioning to `summaryView`. The prune call is fast (typically <100ms) so there's no need for async handling or a separate progress indicator. The flow becomes:

`allDeletionsCompleteMsg` → run prune via Cmd → `pruneCompleteMsg` → `summaryView`

**Alternative considered**: Run prune inside `DeleteWorktrees` after all goroutines finish. Rejected because it couples pruning to the deletion function and makes the deletion channel protocol more complex. Keeping it in the TUI layer gives the view control over when to show the result.

### D2: New `PruneWorktrees` function in `internal/worktree/`

A standalone function `PruneWorktrees(runner CommandRunner, repoPath string) error` that runs `git worktree prune`. Returns nil on success or the error. Keeps the worktree package as the place for all worktree lifecycle operations.

### D3: Prune result stored on Model, displayed in summary

Add a `pruneErr error` field to `Model`. The summary view checks this field:
- `nil` → show "Pruned orphaned worktree metadata" (replaces the old tip)
- non-nil → show "Warning: prune failed: <error>" so the user knows to run it manually

### D4: Bubble Tea Cmd pattern for prune

Use a Cmd that runs the prune and returns a `pruneCompleteMsg{err error}`. This keeps the prune off the main goroutine and follows the Elm architecture pattern already established.

## Risks / Trade-offs

- [Prune fails silently in edge cases] → Mitigated by surfacing the error in summary view. User can still run prune manually.
- [Adds ~100ms to the transition to summary] → Acceptable. The user just waited for deletions; a brief prune is imperceptible.
- [Prune runs even if all deletions failed] → This is correct behavior. Even failed `git worktree remove --force` can leave partial metadata. Prune is idempotent and safe regardless.
