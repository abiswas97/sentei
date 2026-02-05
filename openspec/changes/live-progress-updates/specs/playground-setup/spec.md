## ADDED Requirements

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
