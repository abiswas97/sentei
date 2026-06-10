## MODIFIED Requirements

### Requirement: Display deletion progress
The TUI SHALL display a progress view during deletion showing a progress bar, percentage, and per-worktree status, updating in real-time as each deletion event is received from the deletion channel. The percentage and per-worktree statuses SHALL be computed exclusively from the current run's outcomes and total; outcomes from any previous run in the same session SHALL NOT contribute.

#### Scenario: Progress bar updates incrementally
- **WHEN** a worktree deletion completes (success or failure) and 2 out of 5 total have now finished
- **THEN** the progress bar SHALL immediately re-render showing 40% completion with "2/5" label

#### Scenario: Per-worktree status on deletion start
- **WHEN** a `DeletionStarted` event is received for a worktree
- **THEN** the TUI SHALL show "removing..." next to that worktree

#### Scenario: Per-worktree status on success
- **WHEN** a `DeletionCompleted` event is received for a worktree
- **THEN** the TUI SHALL show a checkmark and "removed" next to that worktree

#### Scenario: Per-worktree status on failure
- **WHEN** a `DeletionFailed` event is received for a worktree
- **THEN** the TUI SHALL show an error indicator and "failed" next to that worktree

#### Scenario: Events forwarded as Bubble Tea messages
- **WHEN** the deletion goroutine sends events over the progress channel
- **THEN** each event SHALL be delivered to the Bubble Tea Update loop as an individual message, not batched

#### Scenario: Percentage never exceeds 100
- **WHEN** a deletion run is in progress, regardless of how many runs preceded it in the session
- **THEN** the displayed percentage SHALL be between 0 and 100 inclusive, derived from current-run outcomes over the current-run total

#### Scenario: Teardown phase visible while running
- **WHEN** the selected worktrees have integration artifacts and the pre-deletion teardown phase is executing
- **THEN** the progress view SHALL show a Teardown phase with the active indicator instead of appearing idle on pending deletion rows

### Requirement: Transition to summary after all deletions
The TUI SHALL run `git worktree prune` via a Cmd after all deletions complete, then run cleanup, then fire an eager worktree reload alongside the transition to summary view. Completion detection SHALL be satisfied exactly when every worktree selected for the current run has an outcome, independent of any previous run's outcomes.

#### Scenario: All deletions complete triggers prune
- **WHEN** every selected worktree has either succeeded or failed
- **THEN** the TUI SHALL execute the prune operation as a Bubble Tea Cmd before transitioning to summary

#### Scenario: Prune completes successfully
- **WHEN** the prune Cmd returns with no error
- **THEN** the TUI SHALL store a nil prune error on the model and transition to the summary view

#### Scenario: Prune fails
- **WHEN** the prune Cmd returns with an error
- **THEN** the TUI SHALL store the error on the model and transition to the summary view

#### Scenario: Cleanup complete fires eager reload
- **WHEN** the cleanup Cmd completes (the final step before summary transition)
- **THEN** the handler SHALL increment `worktreeGeneration`, fire `loadWorktreeContext` via `tea.Batch` alongside the `holdOrAdvance` command, and NOT set `stateStale`

#### Scenario: Second run completes with fewer worktrees than the first
- **WHEN** a first run deleted 2 worktrees and a second run deletes 1
- **THEN** the second run SHALL transition to summary when its single outcome arrives
