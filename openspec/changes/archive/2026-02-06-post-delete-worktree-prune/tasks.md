## 1. Prune Function

- [x] 1.1 Add `PruneWorktrees(runner CommandRunner, repoPath string) error` to `internal/worktree/deleter.go` that runs `git worktree prune`
- [x] 1.2 Add unit tests for `PruneWorktrees` in `internal/worktree/deleter_test.go` (success and failure cases)

## 2. TUI Integration

- [x] 2.1 Add `pruneErr error` field to `Model` in `internal/tui/model.go`
- [x] 2.2 Add `pruneCompleteMsg` type and `runPrune` Cmd in `internal/tui/progress.go`
- [x] 2.3 Update `allDeletionsCompleteMsg` handler in `updateProgress` to return the prune Cmd instead of transitioning to summary
- [x] 2.4 Add `pruneCompleteMsg` handler in `updateProgress` that stores the error on model and transitions to `summaryView`

## 3. Summary View Update

- [x] 3.1 Update `viewSummary` in `internal/tui/summary.go` to show prune result instead of the manual tip — "Pruned orphaned worktree metadata" on success, or "Warning: failed to prune worktree metadata: <err>" on failure

## 4. Testing

- [x] 4.1 Test prune integration in TUI: verify `allDeletionsCompleteMsg` triggers prune Cmd
- [x] 4.2 Test summary renders prune success and failure messages correctly
