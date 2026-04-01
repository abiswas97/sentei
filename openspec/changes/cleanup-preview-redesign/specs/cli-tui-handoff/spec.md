## MODIFIED Requirements

### Requirement: --yes flag skips CLI confirmation
CLI subcommands that show confirmation dialogs SHALL accept a `--yes` / `-y` flag to skip the confirmation and proceed immediately.

#### Scenario: Cleanup with --yes
- **WHEN** the user runs `sentei cleanup --mode safe --yes`
- **THEN** the system SHALL skip the confirmation dialog and run cleanup immediately

#### Scenario: Remove with --yes
- **WHEN** the user runs `sentei remove --merged --yes`
- **THEN** the system SHALL skip the standard confirmation and proceed to removal

#### Scenario: --yes does NOT skip dirty worktree confirmation
- **WHEN** the user runs `sentei remove --all --yes` and the selection includes dirty worktrees
- **THEN** the system SHALL still show the dirty/unpushed confirmation gate (data loss protection cannot be bypassed with --yes)

#### Scenario: --yes with aggressive cleanup
- **WHEN** the user runs `sentei cleanup --mode aggressive --yes`
- **THEN** the system SHALL skip confirmation and run aggressive cleanup immediately (the user explicitly chose aggressive mode via flag)

#### Scenario: --yes without required flags
- **WHEN** the user runs `sentei cleanup --yes` (no --mode flag, defaults to safe)
- **THEN** the system SHALL proceed with safe mode cleanup without confirmation
