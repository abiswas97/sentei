## ADDED Requirements

### Requirement: Display enriched worktrees in a scrollable list
The TUI SHALL display all non-bare worktrees in a scrollable list. Each row SHALL show: a selection checkbox, a status indicator, the branch name (with `refs/heads/` prefix stripped), relative last activity time, and the last commit subject (truncated if necessary).

#### Scenario: Normal worktree row
- **WHEN** a worktree has Branch="refs/heads/feature-x", LastCommitDate=3 days ago, LastCommitSubject="Add OAuth2 flow", HasUncommittedChanges=false, HasUntrackedFiles=false, IsLocked=false
- **THEN** the row SHALL display: `[ ] [ok] feature-x          3 days ago    Add OAuth2 flow`

#### Scenario: Dirty worktree row
- **WHEN** a worktree has HasUncommittedChanges=true
- **THEN** the status indicator SHALL be `[~]` (dirty)

#### Scenario: Untracked files worktree row
- **WHEN** a worktree has HasUntrackedFiles=true and HasUncommittedChanges=false
- **THEN** the status indicator SHALL be `[!]` (untracked)

#### Scenario: Locked worktree row
- **WHEN** a worktree has IsLocked=true
- **THEN** the status indicator SHALL be `[L]` (locked)

#### Scenario: Bare repository entry excluded
- **WHEN** a worktree has IsBare=true
- **THEN** it SHALL NOT appear in the list

#### Scenario: Prunable worktree row
- **WHEN** a worktree has IsPrunable=true
- **THEN** the row SHALL still appear in the list with a distinct indicator showing it is prunable

#### Scenario: Enrichment error
- **WHEN** a worktree has a non-empty EnrichmentError
- **THEN** the row SHALL display the branch name and an error indicator instead of commit metadata

### Requirement: Keyboard navigation
The TUI SHALL support navigating the worktree list using j/k keys, up/down arrow keys, and page up/page down for larger jumps.

#### Scenario: Move cursor down
- **WHEN** the user presses j or down-arrow and the cursor is not on the last item
- **THEN** the cursor SHALL move to the next item

#### Scenario: Move cursor up
- **WHEN** the user presses k or up-arrow and the cursor is not on the first item
- **THEN** the cursor SHALL move to the previous item

#### Scenario: Cursor at boundary
- **WHEN** the user presses j on the last item or k on the first item
- **THEN** the cursor SHALL not move

#### Scenario: Viewport scrolling
- **WHEN** the cursor moves beyond the visible viewport
- **THEN** the viewport SHALL scroll to keep the cursor visible

### Requirement: Multi-select worktrees
The TUI SHALL allow selecting and deselecting individual worktrees with spacebar, and toggling all worktrees with the 'a' key.

#### Scenario: Toggle single selection
- **WHEN** the user presses spacebar on an unselected worktree
- **THEN** the worktree SHALL become selected and its checkbox SHALL show [x]

#### Scenario: Toggle single deselection
- **WHEN** the user presses spacebar on a selected worktree
- **THEN** the worktree SHALL become deselected and its checkbox SHALL show [ ]

#### Scenario: Select all
- **WHEN** the user presses 'a' and not all worktrees are selected
- **THEN** all worktrees SHALL become selected

#### Scenario: Deselect all
- **WHEN** the user presses 'a' and all worktrees are already selected
- **THEN** all worktrees SHALL become deselected

### Requirement: Initiate deletion from list
The TUI SHALL transition to the confirmation view when the user presses Enter with at least one worktree selected.

#### Scenario: Enter with selections
- **WHEN** the user presses Enter and one or more worktrees are selected
- **THEN** the TUI SHALL transition to the confirmation view

#### Scenario: Enter with no selections
- **WHEN** the user presses Enter and no worktrees are selected
- **THEN** the TUI SHALL remain on the list view (no transition)

### Requirement: Quit the application
The TUI SHALL exit when the user presses 'q' or Ctrl+C from the list view.

#### Scenario: Quit from list view
- **WHEN** the user presses 'q' or Ctrl+C on the list view
- **THEN** the application SHALL exit cleanly

### Requirement: Status bar
The TUI SHALL display a status bar at the bottom showing the count of selected worktrees and available key bindings.

#### Scenario: Status bar content
- **WHEN** 3 worktrees are selected out of 10 total
- **THEN** the status bar SHALL display the selection count and key hints (space: toggle, a: all, enter: delete, q: quit)

### Requirement: Relative time display
The TUI SHALL display last commit dates as human-readable relative time strings.

#### Scenario: Recent activity
- **WHEN** the last commit was 2 hours ago
- **THEN** the display SHALL show "2 hours ago"

#### Scenario: Old activity
- **WHEN** the last commit was 90 days ago
- **THEN** the display SHALL show "3 months ago"

#### Scenario: No commit date
- **WHEN** the LastCommitDate is the zero time value
- **THEN** the display SHALL show "unknown"
