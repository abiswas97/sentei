## ADDED Requirements

### Requirement: Sort worktrees by field
The TUI SHALL sort the displayed worktree list by a configurable sort field. The available sort fields SHALL be: last activity date (age) and branch name. The default sort field SHALL be age with ascending direction (oldest first).

#### Scenario: Default sort order on startup
- **WHEN** the application starts and displays the worktree list
- **THEN** worktrees SHALL be sorted by last activity date ascending (oldest first)

#### Scenario: Sort by age ascending
- **WHEN** the sort field is age and direction is ascending
- **THEN** worktrees with the oldest last commit date SHALL appear first and the most recent SHALL appear last

#### Scenario: Sort by branch name ascending
- **WHEN** the sort field is branch name and direction is ascending
- **THEN** worktrees SHALL be sorted alphabetically by branch name (with `refs/heads/` prefix stripped) in A-Z order

#### Scenario: Worktrees with zero-value commit date
- **WHEN** sorting by age and a worktree has a zero-value LastCommitDate (e.g. enrichment failed)
- **THEN** the worktree SHALL sort to the end of the list regardless of sort direction

### Requirement: Cycle sort field with key binding
The TUI SHALL cycle through sort fields when the user presses `s` from the list view. The cycle order SHALL be: age -> branch -> age.

#### Scenario: Cycle from age to branch
- **WHEN** the current sort field is age and the user presses `s`
- **THEN** the sort field SHALL change to branch name and the list SHALL re-sort

#### Scenario: Cycle from branch wraps to age
- **WHEN** the current sort field is branch and the user presses `s`
- **THEN** the sort field SHALL change to age and the list SHALL re-sort

#### Scenario: Sort key disabled during filter input
- **WHEN** the user is actively typing in the filter input
- **THEN** pressing `s` SHALL type the character into the filter input instead of changing the sort field

### Requirement: Reverse sort direction with key binding
The TUI SHALL reverse the sort direction when the user presses `S` (shift+s) from the list view.

#### Scenario: Reverse ascending to descending
- **WHEN** the sort direction is ascending and the user presses `S`
- **THEN** the sort direction SHALL change to descending and the list SHALL re-sort

#### Scenario: Reverse descending to ascending
- **WHEN** the sort direction is descending and the user presses `S`
- **THEN** the sort direction SHALL change to ascending and the list SHALL re-sort

### Requirement: Sort indicator in column headers
The table header row SHALL display the current sort field and direction using arrow indicators on the sorted column. The sorted column header SHALL be visually distinct (bold white) from non-sorted columns (dim gray).

#### Scenario: Sort indicator on age column
- **WHEN** the sort field is age and direction is ascending
- **THEN** the Age column header SHALL display `Age ▲` in bold white and other column headers SHALL be dim

#### Scenario: Sort indicator on branch column
- **WHEN** the sort field is branch and direction is descending
- **THEN** the Branch column header SHALL display `Branch ▼` in bold white and other column headers SHALL be dim

### Requirement: Selection stability across sort changes
Sorting SHALL NOT alter which worktrees are selected. Selections SHALL persist by worktree identity (path), not by display position.

#### Scenario: Selection preserved after sort change
- **WHEN** the user selects worktrees A and B, then changes the sort field
- **THEN** worktrees A and B SHALL remain selected at their new display positions

#### Scenario: Cursor position after sort change
- **WHEN** the user changes the sort field
- **THEN** the cursor SHALL reset to the first item in the newly sorted list
