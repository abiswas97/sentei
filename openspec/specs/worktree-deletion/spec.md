## Purpose
Covers parallel worktree deletion via `git worktree remove --force`, progress reporting through channels, and the public DeleteWorktrees API.

## Requirements

### Requirement: Delete worktrees via git worktree remove
The system SHALL delete worktrees by executing `git worktree remove --force <path>` for each selected worktree.

#### Scenario: Successful deletion
- **WHEN** `git worktree remove --force <path>` succeeds for a worktree
- **THEN** the worktree SHALL be marked as successfully deleted

#### Scenario: Failed deletion
- **WHEN** `git worktree remove --force <path>` fails for a worktree
- **THEN** the system SHALL capture the error message, mark the worktree as failed, and continue deleting remaining worktrees

### Requirement: Parallel deletion execution
The system SHALL execute deletions concurrently, bounded by a configurable maximum concurrency limit (default: 5).

#### Scenario: Concurrent deletion within bounds
- **WHEN** deleting N worktrees with max concurrency M
- **THEN** the system SHALL run at most M deletion operations simultaneously

#### Scenario: All deletions complete
- **WHEN** all deletions have finished (success or failure)
- **THEN** the system SHALL signal completion with a summary of results

### Requirement: Progress reporting via channel
The system SHALL report progress for each deletion via a channel, sending a message when each worktree deletion starts, succeeds, or fails.

#### Scenario: Progress message on start
- **WHEN** a worktree deletion begins
- **THEN** the system SHALL send a "started" message with the worktree path

#### Scenario: Progress message on completion
- **WHEN** a worktree deletion succeeds
- **THEN** the system SHALL send a "completed" message with the worktree path

#### Scenario: Progress message on failure
- **WHEN** a worktree deletion fails
- **THEN** the system SHALL send a "failed" message with the worktree path and error

### Requirement: DeleteWorktrees public function
The system SHALL expose a function `DeleteWorktrees(runner CommandRunner, repoPath string, worktrees []Worktree, maxConcurrency int, progress chan<- DeletionEvent) DeletionResult` that deletes all given worktrees and returns a summary.

#### Scenario: Full deletion pipeline
- **WHEN** DeleteWorktrees is called with a slice of selected worktrees
- **THEN** it SHALL delete each worktree in parallel, send progress events, and return a DeletionResult containing counts and per-worktree outcomes

#### Scenario: Empty worktree list
- **WHEN** DeleteWorktrees is called with an empty slice
- **THEN** it SHALL return immediately with zero counts and close the progress channel

### Requirement: Non-interactive worktree removal via flags
The `remove` command SHALL support non-interactive execution when `--non-interactive` and `--force` are provided alongside filter flags (`--stale`, `--merged`, or `--all`).

#### Scenario: Remove merged worktrees non-interactively
- **WHEN** the user runs `sentei remove --merged --force --non-interactive`
- **THEN** the system SHALL resolve all merged-branch worktrees, delete them using `git worktree remove --force <path>`, and print a summary to stdout

#### Scenario: Remove stale worktrees non-interactively
- **WHEN** the user runs `sentei remove --stale 30d --force --non-interactive`
- **THEN** the system SHALL resolve all worktrees with last commit older than 30 days, delete them, and print a summary

#### Scenario: Remove all non-interactively
- **WHEN** the user runs `sentei remove --all --force --non-interactive`
- **THEN** the system SHALL delete all non-protected worktrees and print a summary

#### Scenario: No worktrees match filter
- **WHEN** the user runs `sentei remove --stale 30d --force --non-interactive` and no worktrees are older than 30 days
- **THEN** the system SHALL print "No worktrees match the filter criteria" and exit with code 0

#### Scenario: Protected worktrees excluded
- **WHEN** the user runs `sentei remove --all --force --non-interactive` and "main" is a protected branch
- **THEN** the "main" worktree SHALL NOT be deleted

#### Scenario: Non-interactive summary format
- **WHEN** worktree removal completes in non-interactive mode
- **THEN** the system SHALL print a summary showing: count deleted, count failed (with error messages), and count skipped (protected)
