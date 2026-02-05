### Requirement: Worktree data model
The system SHALL represent each worktree as a struct with fields: Path (string), HEAD (string), Branch (string), IsBare (bool), IsLocked (bool), LockReason (string), IsPrunable (bool), PruneReason (string), IsDetached (bool), LastCommitDate (time.Time), LastCommitSubject (string), HasUncommittedChanges (bool), HasUntrackedFiles (bool), IsEnriched (bool), EnrichmentError (string).

#### Scenario: Normal worktree
- **WHEN** a porcelain block contains `worktree`, `HEAD`, and `branch` lines
- **THEN** the struct SHALL have Path, HEAD, and Branch populated, with IsBare=false, IsDetached=false

#### Scenario: Bare repository entry
- **WHEN** a porcelain block contains `worktree` and `bare` lines
- **THEN** the struct SHALL have IsBare=true, with HEAD and Branch empty

#### Scenario: Detached HEAD worktree
- **WHEN** a porcelain block contains `worktree`, `HEAD`, and `detached` lines but no `branch` line
- **THEN** the struct SHALL have IsDetached=true and Branch as empty string

#### Scenario: Newly parsed worktree before enrichment
- **WHEN** a worktree is freshly parsed from porcelain output
- **THEN** enrichment fields SHALL be at zero values: LastCommitDate is zero time, LastCommitSubject is empty, HasUncommittedChanges is false, HasUntrackedFiles is false, IsEnriched is false, EnrichmentError is empty

### Requirement: Parse porcelain output
The system SHALL parse the full output of `git worktree list --porcelain` into a slice of Worktree structs. Blocks are separated by blank lines. Each block starts with a `worktree <path>` line.

#### Scenario: Multiple worktrees
- **WHEN** the porcelain output contains N worktree blocks separated by blank lines
- **THEN** the parser SHALL return exactly N Worktree structs in the same order

#### Scenario: Empty input
- **WHEN** the porcelain output is an empty string
- **THEN** the parser SHALL return an empty slice and no error

#### Scenario: Locked worktree without reason
- **WHEN** a porcelain block contains a line that is exactly `locked`
- **THEN** the struct SHALL have IsLocked=true and LockReason as empty string

#### Scenario: Locked worktree with reason
- **WHEN** a porcelain block contains a line `locked <reason>`
- **THEN** the struct SHALL have IsLocked=true and LockReason set to `<reason>`

#### Scenario: Prunable worktree without reason
- **WHEN** a porcelain block contains a line that is exactly `prunable`
- **THEN** the struct SHALL have IsPrunable=true and PruneReason as empty string

#### Scenario: Prunable worktree with reason
- **WHEN** a porcelain block contains a line `prunable <reason>`
- **THEN** the struct SHALL have IsPrunable=true and PruneReason set to `<reason>`

#### Scenario: Branch ref parsing
- **WHEN** a porcelain block contains `branch refs/heads/feature-x`
- **THEN** the Branch field SHALL be set to the full ref `refs/heads/feature-x`

### Requirement: Git command execution abstraction
The system SHALL define a CommandRunner interface with a method `Run(dir string, args ...string) (string, error)` and provide a concrete implementation that executes git commands via `os/exec`.

#### Scenario: Successful command
- **WHEN** a git command executes successfully
- **THEN** Run SHALL return the trimmed stdout and nil error

#### Scenario: Failed command
- **WHEN** a git command exits with non-zero status
- **THEN** Run SHALL return an error containing the stderr output

### Requirement: Repository validation
The system SHALL validate that a given path is a git repository before attempting to list worktrees, by running `git -C <path> rev-parse --git-dir`.

#### Scenario: Valid git repository
- **WHEN** `rev-parse --git-dir` succeeds for the given path
- **THEN** the system SHALL proceed to list worktrees

#### Scenario: Not a git repository
- **WHEN** `rev-parse --git-dir` fails for the given path
- **THEN** the system SHALL return an error with the message "not a git repository" (or containing that phrase)

#### Scenario: Path does not exist
- **WHEN** the given path does not exist on disk
- **THEN** the system SHALL return an error indicating the path is invalid

### Requirement: ListWorktrees public API
The system SHALL expose a function `ListWorktrees(repoPath string) ([]Worktree, error)` that validates the repository, executes `git worktree list --porcelain`, and returns parsed results.

#### Scenario: End-to-end listing
- **WHEN** ListWorktrees is called with a valid git repository path
- **THEN** it SHALL return a slice of Worktree structs representing all worktrees in the repository

#### Scenario: Repository validation failure
- **WHEN** ListWorktrees is called with a path that is not a git repository
- **THEN** it SHALL return nil and an error without executing `git worktree list`
