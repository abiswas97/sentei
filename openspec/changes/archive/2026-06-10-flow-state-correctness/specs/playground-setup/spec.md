## MODIFIED Requirements

### Requirement: Setup creates a bare repository with worktrees
The system SHALL create a bare git repository inside a unique per-session playground directory (created via `os.MkdirTemp` under the system temp directory) and add worktrees representing all states the TUI handles. Two concurrent playground sessions SHALL NOT share or affect each other's directories.

#### Scenario: Fresh setup
- **WHEN** `Setup()` is called
- **THEN** it SHALL create a new unique playground directory containing a bare repo and return the repo path and a cleanup function

#### Scenario: Concurrent sessions are isolated
- **WHEN** two playground sessions run at the same time
- **THEN** each SHALL operate in its own directory, and worktrees created or removed in one session SHALL never appear in or disappear from the other

#### Scenario: Setup returns cleanup function
- **WHEN** `Setup()` returns successfully
- **THEN** the cleanup function SHALL remove that session's entire playground directory when called
