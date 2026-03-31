## ADDED Requirements

### Requirement: Confirmation gate for dirty/unpushed worktrees
The system SHALL display a confirmation view when the user proceeds with worktree removal and the selection includes worktrees with uncommitted changes, untracked files, or commits not pushed to remote.

#### Scenario: All clean and pushed — no gate
- **WHEN** the user confirms removal of 5 worktrees, all clean with branches pushed to remote
- **THEN** the system SHALL proceed directly to removal progress with no confirmation gate

#### Scenario: Dirty worktree triggers gate
- **WHEN** the user confirms removal and 1 of 3 selected worktrees has HasUncommittedChanges=true
- **THEN** the system SHALL display a confirmation view listing all selected worktrees with their risk status

#### Scenario: Unpushed worktree triggers gate
- **WHEN** the user confirms removal and 1 worktree has commits not pushed to remote
- **THEN** the system SHALL display a confirmation view highlighting the unpushed worktree

#### Scenario: Untracked files trigger gate
- **WHEN** the user confirms removal and 1 worktree has HasUntrackedFiles=true
- **THEN** the system SHALL display a confirmation view highlighting the worktree with untracked files

### Requirement: Confirmation view shows at-risk worktrees
The confirmation gate SHALL display each selected worktree with its risk status and a summary of potential data loss.

#### Scenario: Mixed risk levels
- **WHEN** the confirmation gate is shown with 1 clean, 1 dirty, and 1 unpushed worktree
- **THEN** the view SHALL list: `● feature/auth  clean`, `⚠ feature/dashboard  dirty` with "2 uncommitted changes", and `⚠ fix/hotfix  not pushed` with "3 commits ahead of remote"

#### Scenario: Summary warning
- **WHEN** 2 of 5 selected worktrees are at risk
- **THEN** the view SHALL display "2 worktrees have potential data loss" below the list

#### Scenario: Confirm proceeds to removal
- **WHEN** the user presses Enter on the confirmation gate
- **THEN** the system SHALL proceed to removal progress

#### Scenario: Back returns to selection
- **WHEN** the user presses Esc on the confirmation gate
- **THEN** the system SHALL return to the worktree list with selections preserved

### Requirement: Pushed-to-remote detection
The system SHALL detect whether a worktree's branch has been pushed to its remote tracking branch.

#### Scenario: Branch with remote tracking and up to date
- **WHEN** a worktree's branch has a remote tracking branch and `git rev-list HEAD..@{upstream}` returns 0 and `git rev-list @{upstream}..HEAD` returns 0
- **THEN** the worktree SHALL be considered "pushed"

#### Scenario: Branch with commits ahead of remote
- **WHEN** `git rev-list @{upstream}..HEAD` returns N > 0
- **THEN** the worktree SHALL be considered "not pushed" with "N commits ahead of remote"

#### Scenario: Branch with no remote tracking
- **WHEN** the branch has no upstream configured
- **THEN** the worktree SHALL be considered "not pushed" with "no remote tracking branch"
