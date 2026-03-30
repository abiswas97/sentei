## ADDED Requirements

### Requirement: Confirmation view displays resolved options as key-value summary
The TUI SHALL provide a reusable confirmation view component that renders a title and a list of key-value pairs representing the resolved options for a command.

#### Scenario: Create worktree confirmation
- **WHEN** the confirmation view is shown for a create operation with Branch="feature/foo", Base="main", Ecosystems=["go", "pnpm"], MergeBase=true, CopyEnvFiles=false
- **THEN** the view SHALL display "Create Worktree" as the title and list each option as a labeled row

#### Scenario: Cleanup confirmation
- **WHEN** the confirmation view is shown for a cleanup operation with Mode="aggressive"
- **THEN** the view SHALL display "Cleanup Branches" as the title with Mode="aggressive" as a labeled row

### Requirement: Confirmation view shows equivalent CLI command
The confirmation view SHALL display the equivalent CLI command below the options summary. The command string SHALL be generated from the options struct so it stays in sync with the displayed values.

#### Scenario: CLI command for create
- **WHEN** the confirmation view shows Branch="feature/foo", Base="main", MergeBase=true
- **THEN** the CLI command line SHALL read: `sentei create --branch feature/foo --base main --merge-base`

#### Scenario: CLI command for cleanup
- **WHEN** the confirmation view shows Mode="safe"
- **THEN** the CLI command line SHALL read: `sentei cleanup --mode safe`

#### Scenario: CLI command includes --non-interactive hint
- **WHEN** any confirmation view is rendered
- **THEN** the displayed command SHALL NOT include `--non-interactive` (the user is already in TUI mode; the command shows what they'd type to skip the TUI next time, and they can append `--non-interactive` themselves)

### Requirement: Confirmation view accepts enter to proceed or esc to go back
The confirmation view SHALL transition to the execution phase when the user presses Enter, and return to the previous step (or the first step of the flow if entered via flags) when the user presses Escape.

#### Scenario: User confirms
- **WHEN** the user presses Enter on the confirmation view
- **THEN** the TUI SHALL transition to the progress/execution view

#### Scenario: User goes back from flag-based entry
- **WHEN** the user entered the confirmation view via flags (all required flags provided) and presses Escape
- **THEN** the TUI SHALL navigate to the first step of the full flow, allowing the user to walk through all screens

#### Scenario: User goes back from normal flow
- **WHEN** the user reached the confirmation view through the normal multi-step TUI flow and presses Escape
- **THEN** the TUI SHALL navigate back to the previous step in the flow

#### Scenario: User quits
- **WHEN** the user presses 'q' or Ctrl+C on the confirmation view
- **THEN** the TUI SHALL exit cleanly
