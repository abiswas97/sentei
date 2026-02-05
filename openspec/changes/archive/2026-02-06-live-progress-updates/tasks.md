## 1. DelayRunner

- [x] 1.1 Add `DelayRunner` struct to `internal/git/commands.go` implementing `CommandRunner` with `Inner CommandRunner` and `Delay time.Duration` fields
- [x] 1.2 Write test for `DelayRunner` verifying it delegates to inner runner and preserves output/error

## 2. Live Progress Event Forwarding

- [x] 2.1 Add `progressCh <-chan worktree.DeletionEvent` field to `Model` in `model.go`
- [x] 2.2 Rewrite `startDeletion()` in `progress.go` to create the channel, spawn the deletion goroutine, store channel on model, and return a `waitForDeletionEvent` Cmd
- [x] 2.3 Implement `waitForDeletionEvent` function that reads one event from the channel and returns the corresponding `tea.Msg` (or `allDeletionsCompleteMsg` on channel close)
- [x] 2.4 Update `updateProgress` to handle `worktreeDeleteStartedMsg`, `worktreeDeletedMsg`, and `worktreeDeleteFailedMsg` — update `deletionStatuses` and `deletionDone`, then return next `waitForDeletionEvent` Cmd

## 3. Playground Integration

- [x] 3.1 In `main.go`, when `--playground` is set, wrap runner with `DelayRunner` (800ms) for the TUI model while keeping the unwrapped runner for enrichment

## 4. Verification

- [x] 4.1 Run `go build` and `go vet ./...` to verify compilation
- [x] 4.2 Run `go test ./...` to verify all existing tests pass
- [ ] 4.3 Manual test with `go run . --playground` — confirm progress bar increments and per-worktree statuses transition live (requires interactive TUI)
