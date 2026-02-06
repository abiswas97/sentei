## MODIFIED Requirements

### Requirement: Selection behavior with filter
Selections SHALL persist across filter changes. The `a` (select all) key SHALL operate only on currently visible (filtered) worktrees. Protected worktrees SHALL be excluded from select-all regardless of filter state.

#### Scenario: Hidden selections persist
- **WHEN** the user selects worktree A, then applies a filter that hides worktree A
- **THEN** worktree A SHALL remain selected (visible when filter is cleared or in the confirmation dialog)

#### Scenario: Select all with active filter
- **WHEN** a filter is active showing 3 of 10 worktrees (1 protected) and the user presses `a`
- **THEN** only the 2 visible non-protected worktrees SHALL be toggled, not all 10

#### Scenario: Selected count includes hidden
- **WHEN** 2 worktrees are selected but only 1 is visible due to filter
- **THEN** the status bar SHALL show "2 selected" (total, not just visible)

#### Scenario: Confirmation shows all selected
- **WHEN** the user presses enter to confirm deletion with a filter active
- **THEN** the confirmation dialog SHALL show ALL selected worktrees, including those hidden by the filter
