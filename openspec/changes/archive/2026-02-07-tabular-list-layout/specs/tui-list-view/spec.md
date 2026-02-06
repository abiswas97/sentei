## MODIFIED Requirements

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

## ADDED Requirements

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
