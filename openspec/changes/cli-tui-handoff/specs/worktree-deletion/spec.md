## ADDED Requirements

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
