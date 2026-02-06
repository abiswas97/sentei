## ADDED Requirements

### Requirement: CLI accepts --dry-run flag
The CLI SHALL accept a `--dry-run` boolean flag that enables non-interactive preview mode.

#### Scenario: Flag is recognized
- **WHEN** user runs `sentei --dry-run`
- **THEN** the program enters dry-run mode (no TUI launched)

#### Scenario: Combines with --playground
- **WHEN** user runs `sentei --playground --dry-run`
- **THEN** a playground repo is created, dry-run output is printed, and playground is cleaned up

#### Scenario: Combines with repo path
- **WHEN** user runs `sentei --dry-run /path/to/repo`
- **THEN** dry-run output is printed for the specified repo

### Requirement: Dry-run prints worktree table to stdout
The system SHALL print a formatted table of all non-bare worktrees to stdout showing: status indicator, branch name, age, and last commit subject.

#### Scenario: Output format
- **WHEN** dry-run mode is active with enriched worktrees
- **THEN** output is a text table with columns: Status (`[ok]`/`[~]`/`[!]`/`[L]`), Branch, Age, Subject
- **AND** output contains no ANSI color codes
- **AND** worktrees are sorted by age ascending (oldest first, matching TUI default)

#### Scenario: No worktrees found
- **WHEN** dry-run mode is active and no non-bare worktrees exist
- **THEN** the program prints "No worktrees found (only the main working tree exists)." and exits with code 0

### Requirement: Dry-run exits without user interaction
The system SHALL exit immediately after printing the worktree table. No confirmation, selection, or deletion SHALL occur.

#### Scenario: Immediate exit
- **WHEN** dry-run output is printed
- **THEN** the program exits with code 0
- **AND** no Bubble Tea program is started
- **AND** no worktrees are deleted

### Requirement: Dry-run output is pipe-friendly
The output SHALL be plain text suitable for piping to other tools.

#### Scenario: Piped output
- **WHEN** user runs `sentei --dry-run | grep "\[~\]"`
- **THEN** only lines with dirty worktrees are shown (no ANSI escape codes interfere with grep)
