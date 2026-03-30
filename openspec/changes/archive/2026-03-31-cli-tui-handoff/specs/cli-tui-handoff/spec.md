## ADDED Requirements

### Requirement: Command taxonomy classifies commands as output or decision
The system SHALL classify each command as either "output" (read-only, always non-interactive) or "decision" (requires user choices, defaults to TUI). Output commands SHALL always produce stdout and exit. Decision commands SHALL launch the TUI unless `--non-interactive` is provided.

#### Scenario: Output command runs without TUI
- **WHEN** the user runs `sentei ecosystems`
- **THEN** the system SHALL print output to stdout and exit without launching Bubble Tea

#### Scenario: Decision command launches TUI by default
- **WHEN** the user runs `sentei create` without `--non-interactive`
- **THEN** the system SHALL launch the TUI for the create worktree flow

#### Scenario: Decision command with --non-interactive skips TUI
- **WHEN** the user runs `sentei create --branch foo --base main --non-interactive`
- **THEN** the system SHALL execute the create operation without launching Bubble Tea and print results to stdout

### Requirement: --non-interactive flag forces headless execution
The system SHALL accept a `--non-interactive` flag on all decision commands. When provided, the command SHALL execute without launching the TUI. If required flags are missing, the system SHALL print an error listing the missing flags and exit with a non-zero exit code.

#### Scenario: All required flags provided
- **WHEN** the user runs `sentei cleanup --mode safe --non-interactive`
- **THEN** the system SHALL execute cleanup in safe mode and print results to stdout

#### Scenario: Missing required flags
- **WHEN** the user runs `sentei cleanup --non-interactive` without `--mode`
- **THEN** the system SHALL print an error: "missing required flag: --mode (safe|aggressive)" and exit with code 1

#### Scenario: --non-interactive on output command
- **WHEN** the user runs `sentei ecosystems --non-interactive`
- **THEN** the flag SHALL be accepted but have no effect (output commands are already non-interactive)

### Requirement: --force flag required for destructive --non-interactive operations
The system SHALL require `--force` alongside `--non-interactive` for commands that delete worktrees, branches, or perform migrations. Without `--force`, the system SHALL print a warning and exit with a non-zero exit code.

#### Scenario: Destructive operation without --force
- **WHEN** the user runs `sentei remove --merged --non-interactive` without `--force`
- **THEN** the system SHALL print: "destructive operation requires --force with --non-interactive" and exit with code 1

#### Scenario: Destructive operation with --force
- **WHEN** the user runs `sentei remove --merged --force --non-interactive`
- **THEN** the system SHALL execute the removal without TUI interaction

#### Scenario: Non-destructive operation without --force
- **WHEN** the user runs `sentei create --branch foo --base main --non-interactive`
- **THEN** the system SHALL execute without requiring `--force` (create is not destructive)

### Requirement: Flags build options struct for decision commands
Each decision command SHALL define a typed options struct. Flags SHALL populate this struct via a `ParseFlags` function. Both TUI and `--non-interactive` paths SHALL consume the same options struct for execution.

#### Scenario: Flags parsed into options struct
- **WHEN** the user runs `sentei create --branch feature/foo --base main`
- **THEN** the system SHALL parse flags into a CreateOptions struct with Branch="feature/foo" and Base="main"

#### Scenario: Invalid flag value
- **WHEN** the user runs `sentei cleanup --mode invalid`
- **THEN** the system SHALL print an error: "invalid value for --mode: must be 'safe' or 'aggressive'" and exit with code 1

### Requirement: Partial flags enter TUI at first missing required field
When a decision command receives some but not all required flags (without `--non-interactive`), the TUI SHALL launch at the first screen whose corresponding flag was not provided. Screens for provided flags SHALL be skipped.

#### Scenario: Branch provided, base missing
- **WHEN** the user runs `sentei create --branch feature/foo`
- **THEN** the TUI SHALL open at the base branch selection step, skipping the branch name input

#### Scenario: No flags provided
- **WHEN** the user runs `sentei create`
- **THEN** the TUI SHALL open at the first step of the create flow (branch name input)

#### Scenario: All required flags provided
- **WHEN** the user runs `sentei create --branch feature/foo --base main`
- **THEN** the TUI SHALL open at the confirmation view, skipping all input steps

### Requirement: Command registry routes commands
The system SHALL use a command registry to dispatch commands instead of an ad-hoc if/else chain. Each registered command SHALL declare its name, type (output/decision), flag parser, CLI executor, and TUI builder.

#### Scenario: Known command dispatched
- **WHEN** the user runs `sentei cleanup --mode safe`
- **THEN** the registry SHALL look up "cleanup", classify it as a decision command, parse its flags, and route to the appropriate handler

#### Scenario: Unknown command
- **WHEN** the user runs `sentei foobar`
- **THEN** the system SHALL print "unknown command: foobar" with usage help and exit with code 1

#### Scenario: No command (root)
- **WHEN** the user runs `sentei` with no arguments
- **THEN** the system SHALL launch the TUI at the main menu
