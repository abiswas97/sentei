## ADDED Requirements

### Requirement: Activate filter mode with slash key
The TUI SHALL enter filter mode when the user presses `/` from the list view. In filter mode, a text input bar SHALL appear replacing the status bar, and keyboard input SHALL be directed to the text input.

#### Scenario: Enter filter mode
- **WHEN** the user presses `/` on the list view
- **THEN** a text input bar SHALL appear with a prompt (e.g. `filter: `) and the cursor SHALL be in the text input

#### Scenario: Slash key disabled during filter input
- **WHEN** the user is already in filter mode
- **THEN** pressing `/` SHALL type the character into the filter input

### Requirement: Filter worktrees by branch name
The TUI SHALL filter the displayed worktree list by the text entered in the filter input. The filter SHALL match against branch names using case-insensitive substring matching. Only matching worktrees SHALL be displayed.

#### Scenario: Substring match
- **WHEN** the filter text is "feat" and worktrees have branches "feature/auth", "feature/login", "bugfix/nav"
- **THEN** only "feature/auth" and "feature/login" SHALL be displayed

#### Scenario: Case insensitive match
- **WHEN** the filter text is "Feature"
- **THEN** worktrees with branch "feature/auth" SHALL match

#### Scenario: Empty filter shows all
- **WHEN** the filter text is empty
- **THEN** all worktrees SHALL be displayed (same as no filter)

#### Scenario: No matches
- **WHEN** the filter text matches no branch names
- **THEN** the list SHALL be empty and a message such as "no matches" SHALL be displayed

#### Scenario: Real-time filtering
- **WHEN** the user types characters in the filter input
- **THEN** the list SHALL update after each keystroke

### Requirement: Exit filter mode
The TUI SHALL exit filter mode when the user presses `esc` or `enter` while the filter input is focused.

#### Scenario: Exit with esc clears filter
- **WHEN** the user presses `esc` while in filter mode
- **THEN** the filter text SHALL be cleared, all worktrees SHALL be shown, and focus SHALL return to list navigation

#### Scenario: Exit with enter applies filter
- **WHEN** the user presses `enter` while in filter mode
- **THEN** the filter SHALL remain applied with the current text, and focus SHALL return to list navigation

#### Scenario: Clear applied filter with esc
- **WHEN** a filter is applied (non-empty text, not in filter mode) and the user presses `esc`
- **THEN** the filter SHALL be cleared and all worktrees SHALL be shown

#### Scenario: Quit requires esc after filter clear
- **WHEN** the user presses `esc` to clear an applied filter
- **THEN** the application SHALL NOT quit; a subsequent `esc` or `q` SHALL quit

### Requirement: Filter state display
The status bar SHALL indicate when a filter is applied and how many worktrees match.

#### Scenario: Filter applied indicator
- **WHEN** a filter is applied with text "feat" matching 3 of 10 worktrees
- **THEN** the status bar SHALL display the filter text and match count, e.g. `filter: "feat" (3/10)`

#### Scenario: Filter mode active indicator
- **WHEN** the user is actively typing in the filter input
- **THEN** the filter input bar SHALL replace the normal status bar and the legend line SHALL be replaced with contextual key hints showing `enter: apply | esc: cancel`

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

### Requirement: Cursor clamping after filter change
The cursor SHALL be clamped to valid bounds whenever the visible list changes due to filter updates.

#### Scenario: Cursor beyond filtered list
- **WHEN** the cursor is at position 8 and a filter reduces the visible list to 3 items
- **THEN** the cursor SHALL move to position 2 (last item in filtered list)

#### Scenario: Filter removes all items
- **WHEN** the filter matches zero worktrees
- **THEN** the cursor SHALL be at position 0
