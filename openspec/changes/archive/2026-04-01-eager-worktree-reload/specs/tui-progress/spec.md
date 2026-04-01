## MODIFIED Requirements

### Requirement: Transition to summary after all deletions
The TUI SHALL run `git worktree prune` via a Cmd after all deletions complete, then run cleanup, then fire an eager worktree reload alongside the transition to summary view.

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
