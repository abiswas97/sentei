## MODIFIED Requirements

### Requirement: Transition to summary after all deletions
The TUI SHALL run `git worktree prune` via a Cmd after all deletions complete, then transition to the summary view with the prune result.

#### Scenario: All deletions complete triggers prune
- **WHEN** every selected worktree has either succeeded or failed
- **THEN** the TUI SHALL execute the prune operation as a Bubble Tea Cmd before transitioning to summary

#### Scenario: Prune completes successfully
- **WHEN** the prune Cmd returns with no error
- **THEN** the TUI SHALL store a nil prune error on the model and transition to the summary view

#### Scenario: Prune fails
- **WHEN** the prune Cmd returns with an error
- **THEN** the TUI SHALL store the error on the model and transition to the summary view

### Requirement: Post-deletion summary
The TUI SHALL display a summary showing the count of successfully removed worktrees, the count of failures, details for any failures, and the prune result.

#### Scenario: All successful with prune success
- **WHEN** all 3 selected worktrees were deleted successfully and prune succeeded
- **THEN** the summary SHALL show "3 worktrees removed successfully", "Pruned orphaned worktree metadata", and no error section

#### Scenario: Mixed results with prune success
- **WHEN** 2 worktrees succeeded, 1 failed, and prune succeeded
- **THEN** the summary SHALL show "2 removed, 1 failed", list the failed worktree with its error, and "Pruned orphaned worktree metadata"

#### Scenario: Prune failed
- **WHEN** deletions are complete and prune failed
- **THEN** the summary SHALL show "Warning: failed to prune worktree metadata" with the error, so the user can run it manually
