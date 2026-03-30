## ADDED Requirements

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
