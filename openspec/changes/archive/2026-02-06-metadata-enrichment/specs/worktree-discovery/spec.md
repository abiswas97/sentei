## MODIFIED Requirements

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
