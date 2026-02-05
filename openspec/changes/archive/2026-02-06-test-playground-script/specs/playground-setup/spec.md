## ADDED Requirements

### Requirement: Setup creates a bare repository with worktrees
The system SHALL create a bare git repository at `/tmp/wt-sweep-playground/repo.git` and add worktrees representing all states the TUI handles.

#### Scenario: Fresh setup
- **WHEN** `Setup()` is called and no playground directory exists
- **THEN** it SHALL create a bare repo and return the repo path and a cleanup function

#### Scenario: Idempotent setup
- **WHEN** `Setup()` is called and a playground directory already exists
- **THEN** it SHALL remove the existing directory, create a fresh playground, and return without error

#### Scenario: Setup returns cleanup function
- **WHEN** `Setup()` returns successfully
- **THEN** the cleanup function SHALL remove the entire `/tmp/wt-sweep-playground/` directory when called

### Requirement: Playground includes a clean worktree
The system SHALL create a worktree on branch `feature/active` with a clean working tree and a recent commit.

#### Scenario: Clean worktree state
- **WHEN** the playground is set up
- **THEN** `feature/active` SHALL have no uncommitted changes, no untracked files, and a commit within the last hour

### Requirement: Playground includes a dirty worktree
The system SHALL create a worktree on branch `feature/wip` with uncommitted changes (modified tracked files).

#### Scenario: Dirty worktree state
- **WHEN** the playground is set up
- **THEN** `feature/wip` SHALL have HasUncommittedChanges=true when enriched

### Requirement: Playground includes a worktree with untracked files
The system SHALL create a worktree on branch `experiment/abandoned` with untracked files but no uncommitted changes to tracked files.

#### Scenario: Untracked files state
- **WHEN** the playground is set up
- **THEN** `experiment/abandoned` SHALL have HasUntrackedFiles=true and HasUncommittedChanges=false when enriched

### Requirement: Playground includes a locked worktree
The system SHALL create a worktree on branch `hotfix/locked` and lock it via `git worktree lock`.

#### Scenario: Locked worktree state
- **WHEN** the playground is set up
- **THEN** `hotfix/locked` SHALL have IsLocked=true when discovered

### Requirement: Playground includes an old worktree
The system SHALL create a worktree on branch `chore/old-deps` with its last commit backdated to at least 90 days ago.

#### Scenario: Old commit date
- **WHEN** the playground is set up and the worktree is enriched
- **THEN** `chore/old-deps` SHALL have a LastCommitDate at least 90 days in the past

### Requirement: Playground includes a detached HEAD worktree
The system SHALL create a worktree in detached HEAD state (checked out to a specific commit, not a branch).

#### Scenario: Detached HEAD state
- **WHEN** the playground is set up
- **THEN** the detached worktree SHALL have IsDetached=true and an empty Branch when discovered

### Requirement: CLI flag integration
The system SHALL accept a `--playground` flag that triggers playground setup, launches the TUI against the playground repo, and cleans up on exit.

#### Scenario: Playground flag launches TUI
- **WHEN** the user runs `wt-sweep --playground`
- **THEN** the system SHALL set up the playground, run the TUI against it, and remove the playground directory after the TUI exits

#### Scenario: Playground keep flag
- **WHEN** the user runs `wt-sweep --playground --playground-keep`
- **THEN** the system SHALL set up the playground and run the TUI, but NOT remove the playground directory after exit

#### Scenario: Playground flag with repo path
- **WHEN** the user runs `wt-sweep --playground /some/path`
- **THEN** the `--playground` flag SHALL take precedence and the path argument SHALL be ignored
