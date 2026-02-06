## ADDED Requirements

### Requirement: Identify protected branches by convention
The system SHALL consider a worktree protected if its branch name (after stripping `refs/heads/` prefix) exactly matches one of: `main`, `master`, `develop`, `dev`. The match SHALL be case-sensitive. Bare and detached worktrees SHALL NOT be considered protected.

#### Scenario: Standard protected branches
- **WHEN** a worktree has Branch="refs/heads/main"
- **THEN** it SHALL be identified as protected

#### Scenario: All protected names
- **WHEN** worktrees have branches "refs/heads/main", "refs/heads/master", "refs/heads/develop", "refs/heads/dev"
- **THEN** all four SHALL be identified as protected

#### Scenario: Non-protected branch
- **WHEN** a worktree has Branch="refs/heads/feature/dev-tools"
- **THEN** it SHALL NOT be identified as protected (substring match does not count)

#### Scenario: Detached HEAD not protected
- **WHEN** a worktree has IsDetached=true and no branch name
- **THEN** it SHALL NOT be identified as protected regardless of commit

#### Scenario: Case sensitivity
- **WHEN** a worktree has Branch="refs/heads/Main" or "refs/heads/MASTER"
- **THEN** it SHALL NOT be identified as protected (case-sensitive match)

### Requirement: Protected worktrees cannot be selected for deletion
The TUI SHALL prevent selection of protected worktrees. Pressing spacebar on a protected worktree SHALL have no effect. Protected worktrees SHALL NOT be included when toggling select-all.

#### Scenario: Spacebar on protected worktree
- **WHEN** the user presses spacebar on a protected worktree
- **THEN** the selection state SHALL not change

#### Scenario: Select-all skips protected
- **WHEN** the user presses 'a' with 5 visible worktrees, 1 of which is protected
- **THEN** only the 4 non-protected worktrees SHALL be toggled

#### Scenario: Deselect-all with protected present
- **WHEN** the user presses 'a' and all 4 non-protected visible worktrees are selected
- **THEN** all 4 SHALL be deselected; the protected worktree remains unaffected
