## Why

The progress view during worktree deletion does not show live updates. All per-worktree status indicators and the progress bar only update after every deletion has finished, defeating the purpose of the progress UI. For real-world monorepos (e.g. 3GB+ worktrees where `git worktree remove` takes seconds to a minute each), the user stares at a static "removing..." screen with no feedback. The existing `tui-progress` spec already requires live per-worktree updates — the implementation simply doesn't fulfill it.

Additionally, the test playground creates tiny worktrees that delete instantly, making it impossible to observe or test the progress UI during development.

## What Changes

- Fix `startDeletion()` in `progress.go` to forward deletion channel events as individual Bubble Tea messages rather than consuming them silently inside a single Cmd closure
- Store the progress event channel on the Model so `updateProgress` can chain Cmds that each read one event
- Handle `worktreeDeleteStartedMsg`, `worktreeDeletedMsg`, and `worktreeDeleteFailedMsg` in `updateProgress` to incrementally update `deletionStatuses`, `deletionDone`, and the progress bar
- Add a `DelayRunner` CommandRunner wrapper in `internal/git/` that adds a configurable sleep per `Run()` call
- In playground mode, wrap the runner passed to the TUI with `DelayRunner` so deletions take visible time (enrichment still uses the fast runner since it completes before TUI launch)

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `tui-progress`: Implementation does not match spec — progress bar and per-worktree statuses must update live as each deletion event arrives, not only after all deletions complete
- `playground-setup`: Add requirement for simulated slow deletion so the progress UI can be observed during development and demos

## Impact

- `internal/tui/progress.go`: Rewrite `startDeletion` and `updateProgress` to use Bubble Tea Cmd chaining pattern
- `internal/tui/model.go`: Add progress channel field to Model
- `internal/git/commands.go` (or new file): Add `DelayRunner` wrapper
- `main.go`: Wrap runner with `DelayRunner` in playground mode before passing to TUI
