## ADDED Requirements

### Requirement: Display confirmation dialog with selected worktrees
The TUI SHALL show a confirmation dialog listing all selected worktrees before deletion, with clear identification of each worktree by branch name and its clean/dirty status.

#### Scenario: Confirmation with clean worktrees only
- **WHEN** the user enters the confirmation view with 3 selected worktrees, all clean
- **THEN** the dialog SHALL list all 3 worktrees with their branch names and "(clean)" label

#### Scenario: Confirmation with dirty worktrees
- **WHEN** at least one selected worktree has HasUncommittedChanges=true
- **THEN** the dialog SHALL display a warning icon next to that worktree with "HAS UNCOMMITTED CHANGES" and a summary warning at the bottom stating the count of worktrees with uncommitted changes

#### Scenario: Confirmation with untracked files
- **WHEN** at least one selected worktree has HasUntrackedFiles=true
- **THEN** the dialog SHALL display a warning indicating untracked files present

### Requirement: Confirm or cancel deletion
The TUI SHALL require the user to press 'y' to confirm deletion or 'n'/Escape to cancel and return to the list view.

#### Scenario: User confirms
- **WHEN** the user presses 'y' on the confirmation dialog
- **THEN** the TUI SHALL transition to the progress view and begin deletion

#### Scenario: User cancels with 'n'
- **WHEN** the user presses 'n' on the confirmation dialog
- **THEN** the TUI SHALL return to the list view with selections preserved

#### Scenario: User cancels with Escape
- **WHEN** the user presses Escape on the confirmation dialog
- **THEN** the TUI SHALL return to the list view with selections preserved

### Requirement: Locked worktree warning in confirmation
The TUI SHALL display a distinct warning for any selected worktree that is locked.

#### Scenario: Locked worktree in selection
- **WHEN** a selected worktree has IsLocked=true
- **THEN** the confirmation dialog SHALL display a lock indicator and a warning that force-removal will be used
