# tui-design-system Delta

## ADDED Requirements

### Requirement: Selection and danger weight
Cursor rows SHALL use the `▸` marker with selected entry text carrying the accent in menu-style views; destructive confirmation footers SHALL render their destructive hint in the warning style; worktrees SHALL be named by one canonical label across all views (short HEAD hash for detached); the confirm screen SHALL use the list's badge vocabulary; action hints SHALL NOT advertise impossible actions (no delete hint at zero selected).

#### Scenario: Danger hint weighted
- **WHEN** the deletion confirm screen renders
- **THEN** `y delete` SHALL carry the warning style while `n go back` stays dim

#### Scenario: Canonical detached label
- **WHEN** a detached worktree appears in the list and on the confirm screen
- **THEN** both SHALL show the same short HEAD hash

#### Scenario: No dead-end hints
- **WHEN** the list has zero selections
- **THEN** the status bar SHALL NOT offer `enter delete`
