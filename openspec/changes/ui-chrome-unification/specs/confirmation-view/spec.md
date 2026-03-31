## MODIFIED Requirements

### Requirement: Confirmation view displays resolved options as key-value summary
The TUI SHALL provide a reusable confirmation view component that renders using standard view chrome (`viewTitle`, `viewSeparator`, `viewKeyHints`) instead of a bordered dialog box. The view SHALL display a title and a list of key-value pairs representing the resolved options for a command.

#### Scenario: Create worktree confirmation
- **WHEN** the confirmation view is shown for a create operation with Branch="feature/foo", Base="main"
- **THEN** the view SHALL display `viewTitle("Create Worktree")` followed by a separator, then each option as a labeled row, without any border wrapping

#### Scenario: Cleanup confirmation
- **WHEN** the confirmation view is shown for a cleanup operation with Mode="aggressive"
- **THEN** the view SHALL display `viewTitle("Confirm Cleanup")` with Mode="aggressive" as a labeled row, using standard chrome

#### Scenario: No border wrapping
- **WHEN** any confirmation view is rendered
- **THEN** the view SHALL NOT use `styleDialogBox` or any border wrapping — it SHALL use the same full-width layout as all other views

### Requirement: Confirmation view shows equivalent CLI command
The confirmation view SHALL display the equivalent CLI command below the options summary, between the content separator and the key hints.

#### Scenario: CLI command for create
- **WHEN** the confirmation view shows Branch="feature/foo", Base="main", MergeBase=true
- **THEN** the CLI command line SHALL read: `sentei create --branch feature/foo --base main --merge-base`

#### Scenario: CLI command for cleanup
- **WHEN** the confirmation view shows Mode="safe"
- **THEN** the CLI command line SHALL read: `sentei cleanup --mode safe`

### Requirement: Confirmation view accepts enter to proceed or esc to go back
The confirmation view SHALL use `viewKeyHints` to render consistent key hints and transition to the execution phase when the user presses Enter, and return to the previous step when the user presses Escape.

#### Scenario: User confirms
- **WHEN** the user presses Enter on the confirmation view
- **THEN** the TUI SHALL transition to the progress/execution view

#### Scenario: User goes back from flag-based entry
- **WHEN** the user entered the confirmation view via flags and presses Escape
- **THEN** the TUI SHALL navigate to the first step of the full flow

#### Scenario: User goes back from normal flow
- **WHEN** the user reached the confirmation view through the normal multi-step TUI flow and presses Escape
- **THEN** the TUI SHALL navigate back to the previous step in the flow

#### Scenario: User quits
- **WHEN** the user presses 'q' or Ctrl+C on the confirmation view
- **THEN** the TUI SHALL exit cleanly

#### Scenario: Key hints use shared renderer
- **WHEN** the confirmation view renders key hints
- **THEN** it SHALL use `viewKeyHints(KeyHint{"enter", "confirm"}, KeyHint{"esc", "back"}, KeyHint{"q", "quit"})` for consistent formatting
