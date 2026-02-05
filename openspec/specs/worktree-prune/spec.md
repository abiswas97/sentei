
### Requirement: Prune orphaned worktree metadata
The system SHALL prune orphaned worktree metadata by executing `git worktree prune` against the repository after all worktree deletions complete.

#### Scenario: Successful prune
- **WHEN** `git worktree prune` succeeds
- **THEN** the function SHALL return nil

#### Scenario: Failed prune
- **WHEN** `git worktree prune` fails
- **THEN** the function SHALL return the error from the git command

### Requirement: PruneWorktrees public function
The system SHALL expose a function `PruneWorktrees(runner CommandRunner, repoPath string) error` that runs `git worktree prune` and returns the result.

#### Scenario: Prune invocation
- **WHEN** PruneWorktrees is called with a valid runner and repo path
- **THEN** it SHALL execute `git -C <repoPath> worktree prune` and return nil on success or the error on failure
