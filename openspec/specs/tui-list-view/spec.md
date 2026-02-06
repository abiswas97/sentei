
### Requirement: Display enriched worktrees in a scrollable list
The TUI SHALL display all non-bare worktrees in a scrollable list using `lipgloss/table` for column layout. Each row SHALL contain six columns: cursor indicator, selection checkbox, status indicator, branch name (with `refs/heads/` prefix stripped), relative last activity time, and last commit subject. The table SHALL render without borders (no top, bottom, left, right, column, header, or row borders). The table width SHALL match the current terminal width.

#### Scenario: Normal worktree row
- **WHEN** a worktree has Branch="refs/heads/feature-x", LastCommitDate=3 days ago, LastCommitSubject="Add OAuth2 flow", HasUncommittedChanges=false, HasUntrackedFiles=false, IsLocked=false
- **THEN** the row SHALL display columns: cursor indicator, `[ ]`, `[ok]`, `feature-x`, `3 days ago`, `Add OAuth2 flow` with columns aligned across all rows

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
The TUI SHALL display a status bar at the bottom showing the count of selected worktrees, available key bindings, and a status indicator legend. The legend SHALL appear on a separate line below the keybindings line.

#### Scenario: Status bar content
- **WHEN** 3 worktrees are selected out of 10 total
- **THEN** the status bar SHALL display the selection count and key hints (space: toggle, a: all, enter: delete, q: quit)

#### Scenario: Legend line content
- **WHEN** the list view is displayed
- **THEN** a legend line SHALL appear below the keybindings line showing all four status indicators with labels: `[ok] clean  [~] dirty  [!] untracked  [L] locked`

#### Scenario: Legend indicator colors
- **WHEN** the legend line is rendered
- **THEN** each indicator SHALL use the same color style as its corresponding in-table indicator (`[ok]` green, `[~]` orange, `[!]` red, `[L]` gray) and the labels SHALL use a dimmed style

#### Scenario: Viewport height adjustment
- **WHEN** the terminal sends a WindowSizeMsg
- **THEN** the visible list height SHALL account for the legend line so the table does not overflow the terminal

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

### Requirement: Responsive column layout with wrapping
The TUI SHALL adapt column widths to the current terminal width. Fixed-content columns (cursor, checkbox, status, age) SHALL have fixed widths. Variable-content columns (branch, subject) SHALL split remaining width proportionally (~50/50). Data columns (branch, age, subject) SHALL have 1-char right padding via `Padding(0, 1)`; prefix columns (cursor, checkbox, status) SHALL include inter-column gap in their fixed width. When content exceeds a column's width, it SHALL wrap within the cell (row expands vertically). Age and subject SHALL remain top-aligned on the first line of wrapped rows.

#### Scenario: Wide terminal
- **WHEN** the terminal width is 120 characters or more
- **THEN** branch names and commit subjects SHALL display without wrapping for typical content lengths

#### Scenario: Narrow terminal
- **WHEN** the terminal width is 60 characters
- **THEN** the table SHALL shrink variable columns proportionally, wrapping content within cells rather than truncating

#### Scenario: Long branch name wrapping
- **WHEN** a branch name exceeds the branch column width
- **THEN** the branch name SHALL wrap within its cell and the row SHALL expand vertically to accommodate

#### Scenario: Terminal resize
- **WHEN** the terminal is resized during operation
- **THEN** column widths SHALL recalculate on the next render to match the new width

#### Scenario: Long commit subject truncation
- **WHEN** a commit subject exceeds the subject column width
- **THEN** the subject SHALL be truncated with `...` rather than wrapping, so that only branch names wrap within their cells

### Requirement: Terminal width tracking
The Model SHALL store the current terminal width and update it from `tea.WindowSizeMsg` events.

#### Scenario: Initial width
- **WHEN** the application starts and receives its first `tea.WindowSizeMsg`
- **THEN** the Model SHALL store the width for use in table rendering

#### Scenario: Width update on resize
- **WHEN** a `tea.WindowSizeMsg` is received with a new width
- **THEN** the Model SHALL update the stored width
