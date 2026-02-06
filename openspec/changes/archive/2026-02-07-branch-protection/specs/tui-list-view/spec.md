## MODIFIED Requirements

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
