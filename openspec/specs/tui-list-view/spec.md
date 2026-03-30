## Purpose
Covers the main TUI list view including worktree display, multi-select, status bar, legend, and quit behavior.

## Requirements

### Requirement: Display enriched worktrees in a scrollable list
The TUI SHALL display worktrees using a `visibleIndices` mapping rather than iterating `m.worktrees` directly. The visible list SHALL reflect the current sort order and any active filter. All column layout, status indicators, and rendering behavior SHALL remain unchanged. Protected worktrees SHALL display `[P]` in the checkbox column instead of `[ ]` or `[x]`.

#### Scenario: Normal worktree row
- **WHEN** a worktree has Branch="refs/heads/feature-x", LastCommitDate=3 days ago, LastCommitSubject="Add OAuth2 flow", HasUncommittedChanges=false, HasUntrackedFiles=false, IsLocked=false
- **THEN** the row SHALL display columns: cursor indicator, `[ ]`, `[ok]`, `feature-x`, `3 days ago`, `Add OAuth2 flow` with columns aligned across all rows

#### Scenario: Protected worktree row
- **WHEN** a worktree has Branch="refs/heads/main" (a protected branch)
- **THEN** the row SHALL display `[P]` in the checkbox column instead of `[ ]` and the worktree SHALL NOT be selectable

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

### Requirement: Multi-select worktrees
The TUI SHALL allow selecting and deselecting individual worktrees with spacebar, and toggling all visible worktrees with the 'a' key. Selections SHALL be stored by worktree path (`map[string]bool`) rather than display index. Protected worktrees SHALL be excluded from selection — spacebar SHALL have no effect on them and select-all SHALL skip them.

#### Scenario: Toggle single selection
- **WHEN** the user presses spacebar on an unselected, non-protected worktree
- **THEN** the worktree SHALL become selected and its checkbox SHALL show [x]

#### Scenario: Toggle single deselection
- **WHEN** the user presses spacebar on a selected worktree
- **THEN** the worktree SHALL become deselected and its checkbox SHALL show [ ]

#### Scenario: Spacebar on protected worktree
- **WHEN** the user presses spacebar on a protected worktree
- **THEN** nothing SHALL happen; the worktree remains unselectable

#### Scenario: Select all visible
- **WHEN** the user presses 'a' and not all non-protected visible worktrees are selected
- **THEN** all currently visible non-protected worktrees SHALL become selected

#### Scenario: Deselect all visible
- **WHEN** the user presses 'a' and all non-protected visible worktrees are already selected
- **THEN** all currently visible non-protected worktrees SHALL become deselected

### Requirement: Status bar
The TUI SHALL display a status bar at the bottom showing the count of selected worktrees, filter state, available key bindings, and a status indicator legend. The sort indicator SHALL appear in the column headers (not the status bar). The legend SHALL appear on a separate line below the keybindings line. When filter mode is active, the legend line SHALL be replaced with contextual key hints (`enter: apply | esc: cancel`).

#### Scenario: Status bar content
- **WHEN** 3 worktrees are selected and no filter is active
- **THEN** the status bar SHALL display the selection count and key hints (space: toggle, a: all, enter: delete, /: filter, s: sort, q: quit)

#### Scenario: Status bar with active filter
- **WHEN** a filter is applied with text "feat" matching 5 of 12 worktrees
- **THEN** the status bar SHALL include filter info such as `filter: "feat" (5/12)`

#### Scenario: Legend line content
- **WHEN** the list view is displayed
- **THEN** a legend line SHALL appear below the keybindings line showing all four status indicators with labels: `[ok] clean  [~] dirty  [!] untracked  [L] locked`

#### Scenario: Legend indicator colors
- **WHEN** the legend line is rendered
- **THEN** each indicator SHALL use the same color style as its corresponding in-table indicator (`[ok]` green, `[~]` orange, `[!]` red, `[L]` gray) and the labels SHALL use a dimmed style

#### Scenario: Legend replaced during filter mode
- **WHEN** the user is actively typing in the filter input
- **THEN** the legend line SHALL be replaced with `enter: apply | esc: cancel` in dimmed style

### Requirement: Quit the application
The TUI SHALL exit when the user presses 'q' or Ctrl+C from the list view. The `esc` key SHALL be context-dependent: it clears an active filter first, then quits on a subsequent press.

#### Scenario: Quit from list view with no filter
- **WHEN** the user presses 'q', 'esc', or Ctrl+C on the list view with no active filter
- **THEN** the application SHALL exit cleanly

#### Scenario: Esc with active filter clears filter first
- **WHEN** the user presses 'esc' on the list view with an applied filter
- **THEN** the filter SHALL be cleared instead of quitting

#### Scenario: Q always quits regardless of filter
- **WHEN** the user presses 'q' on the list view with an applied filter
- **THEN** the application SHALL exit cleanly (q is unconditional quit)

### Requirement: List view accepts pre-selected worktrees from filter flags
The worktree list view SHALL accept an initial selection set derived from filter flags. Pre-selected worktrees SHALL appear with `[x]` checkboxes when the list first renders.

#### Scenario: Pre-selection from --merged flag
- **WHEN** the list view is initialized with a filter selecting merged worktrees
- **THEN** all merged-branch worktrees SHALL have `[x]` checkboxes and all others SHALL have `[ ]`

#### Scenario: Pre-selection from --stale flag
- **WHEN** the list view is initialized with a filter selecting worktrees stale for 30 days
- **THEN** all worktrees whose last commit is older than 30 days SHALL have `[x]` checkboxes

#### Scenario: Pre-selection from --all flag
- **WHEN** the list view is initialized with the --all filter
- **THEN** all non-protected worktrees SHALL have `[x]` checkboxes

#### Scenario: User can modify pre-selection
- **WHEN** the list view has pre-selected worktrees from filter flags
- **THEN** the user SHALL be able to toggle individual items with spacebar and toggle all with 'a', overriding the initial filter selection

#### Scenario: Pre-selection respects protected worktrees
- **WHEN** the --all filter is applied and branch "main" is protected
- **THEN** the "main" worktree SHALL display `[P]` and SHALL NOT be pre-selected

### Requirement: List view displays active filter indicator
When filter flags produced the current pre-selection, the status bar SHALL indicate which filters are active.

#### Scenario: Stale filter active
- **WHEN** the list view was entered with `--stale 30d`
- **THEN** the status bar SHALL include `filter: stale > 30d`

#### Scenario: Merged filter active
- **WHEN** the list view was entered with `--merged`
- **THEN** the status bar SHALL include `filter: merged`

#### Scenario: Multiple filters active
- **WHEN** the list view was entered with `--stale 14d --merged`
- **THEN** the status bar SHALL include `filter: stale > 14d, merged`
