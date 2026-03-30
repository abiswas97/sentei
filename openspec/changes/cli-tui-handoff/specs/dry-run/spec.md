## MODIFIED Requirements

### Requirement: CLI accepts --dry-run flag
The CLI SHALL accept a `--dry-run` boolean flag that enables non-interactive preview mode. The `--dry-run` flag SHALL be composable with `--non-interactive` and command-specific flags. When used with a decision command, `--dry-run` SHALL show what would happen without executing.

#### Scenario: Flag is recognized
- **WHEN** user runs `sentei --dry-run`
- **THEN** the program enters dry-run mode (no TUI launched)

#### Scenario: Combines with --playground
- **WHEN** user runs `sentei --playground --dry-run`
- **THEN** a playground repo is created, dry-run output is printed, and playground is cleaned up

#### Scenario: Combines with repo path
- **WHEN** user runs `sentei --dry-run /path/to/repo`
- **THEN** dry-run output is printed for the specified repo

#### Scenario: Dry-run with decision command
- **WHEN** user runs `sentei cleanup --mode aggressive --dry-run --non-interactive`
- **THEN** the system SHALL print what branches would be cleaned up without actually deleting them

#### Scenario: Dry-run with remove and filter
- **WHEN** user runs `sentei remove --merged --dry-run --non-interactive`
- **THEN** the system SHALL print which worktrees would be removed without deleting them and exit with code 0
