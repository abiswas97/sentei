
### Requirement: Display deletion progress
The TUI SHALL display a progress view during deletion showing a progress bar, percentage, and per-worktree status.

#### Scenario: Progress bar updates
- **WHEN** 2 out of 5 worktrees have been deleted
- **THEN** the progress bar SHALL show 40% completion with "2/5" label

#### Scenario: Per-worktree status during deletion
- **WHEN** a worktree deletion is in progress
- **THEN** the TUI SHALL show a spinner or "removing..." indicator next to that worktree

#### Scenario: Per-worktree status on success
- **WHEN** a worktree has been successfully deleted
- **THEN** the TUI SHALL show a checkmark and "removed" next to that worktree

#### Scenario: Per-worktree status on failure
- **WHEN** a worktree deletion has failed
- **THEN** the TUI SHALL show an error indicator and the error message next to that worktree

### Requirement: Transition to summary after all deletions
The TUI SHALL automatically transition to the summary view when all deletions are complete.

#### Scenario: All deletions complete
- **WHEN** every selected worktree has either succeeded or failed
- **THEN** the TUI SHALL transition to the summary view

### Requirement: Post-deletion summary
The TUI SHALL display a summary showing the count of successfully removed worktrees, the count of failures, and details for any failures.

#### Scenario: All successful
- **WHEN** all 3 selected worktrees were deleted successfully
- **THEN** the summary SHALL show "3 worktrees removed successfully" with no error section

#### Scenario: Mixed results
- **WHEN** 2 worktrees succeeded and 1 failed
- **THEN** the summary SHALL show "2 removed, 1 failed" and list the failed worktree with its error message

#### Scenario: Suggest prune
- **WHEN** deletions are complete
- **THEN** the summary SHALL suggest running `git worktree prune` to clean up any orphaned metadata

### Requirement: Exit from summary
The TUI SHALL exit when the user presses 'q', Enter, or Escape from the summary view.

#### Scenario: Quit from summary
- **WHEN** the user presses 'q', Enter, or Escape on the summary view
- **THEN** the application SHALL exit cleanly
