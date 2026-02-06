
### Requirement: Setup creates a bare repository with worktrees
The system SHALL create a bare git repository at `/tmp/sentei-playground/repo.git` and add worktrees representing all states the TUI handles.

#### Scenario: Fresh setup
- **WHEN** `Setup()` is called and no playground directory exists
- **THEN** it SHALL create a bare repo and return the repo path and a cleanup function

#### Scenario: Idempotent setup
- **WHEN** `Setup()` is called and a playground directory already exists
- **THEN** it SHALL remove the existing directory, create a fresh playground, and return without error

#### Scenario: Setup returns cleanup function
- **WHEN** `Setup()` returns successfully
- **THEN** the cleanup function SHALL remove the entire `/tmp/sentei-playground/` directory when called

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
- **WHEN** the user runs `sentei --playground`
- **THEN** the system SHALL set up the playground, run the TUI against it, and remove the playground directory after the TUI exits

#### Scenario: Playground keep flag
- **WHEN** the user runs `sentei --playground --playground-keep`
- **THEN** the system SHALL set up the playground and run the TUI, but NOT remove the playground directory after exit

#### Scenario: Playground flag with repo path
- **WHEN** the user runs `sentei --playground /some/path`
- **THEN** the `--playground` flag SHALL take precedence and the path argument SHALL be ignored

### Requirement: Playground uses delayed runner for TUI
The system SHALL wrap the git CommandRunner with a DelayRunner in playground mode so that deletion operations take visible time for progress UI testing.

#### Scenario: DelayRunner wraps real runner
- **WHEN** `--playground` flag is set
- **THEN** the runner passed to the TUI model SHALL be a `DelayRunner` wrapping the real `GitRunner` with an 800ms delay per operation

#### Scenario: Enrichment uses fast runner
- **WHEN** `--playground` flag is set
- **THEN** worktree enrichment (which runs before TUI launch) SHALL use the unwrapped `GitRunner` with no delay

#### Scenario: Non-playground mode unaffected
- **WHEN** `--playground` flag is NOT set
- **THEN** the TUI model SHALL receive the real `GitRunner` with no wrapping

### Requirement: DelayRunner implements CommandRunner
The system SHALL provide a `DelayRunner` struct in `internal/git/` that implements `CommandRunner` by sleeping for a configurable duration then delegating to an inner runner.

#### Scenario: DelayRunner adds sleep before delegation
- **WHEN** `DelayRunner.Run()` is called with a 800ms delay
- **THEN** it SHALL sleep for 800ms then call the inner runner's `Run()` with the same arguments

#### Scenario: DelayRunner preserves inner runner results
- **WHEN** the inner runner returns output and an error
- **THEN** `DelayRunner` SHALL return the same output and error unchanged
