## ADDED Requirements

### Requirement: Single progress vocabulary
The system SHALL define exactly one progress status vocabulary (`StepStatus`: Pending, Running, Done, Failed, Skipped) and exactly one progress event type (`Event` with Phase, Step, Status, Message, Error), both in `internal/progress`, and every multi-phase flow (worktree removal, integration apply, teardown, worktree create, repo create, repo clone, repo migrate, cleanup) SHALL emit its progress exclusively through this vocabulary.

#### Scenario: No parallel progress vocabularies exist
- **WHEN** the codebase is searched for progress event or status type definitions
- **THEN** `internal/progress` contains the only `StepStatus` enum and the only progress event struct, and `internal/pipeline`, `integration.ManagerEvent`, and `worktree.DeletionEvent` do not exist

#### Scenario: Integration apply emits the shared vocabulary
- **WHEN** the integration manager applies changes to a worktree
- **THEN** it emits `progress.Event` values whose Phase is the worktree's display name, preserving the per-worktree phase grouping rendered today

#### Scenario: Worktree removal emits the shared vocabulary
- **WHEN** a worktree deletion starts, completes, or fails
- **THEN** the deleter emits a `progress.Event` with Running, Done, or Failed status respectively, carrying the worktree path as the step and any error on the Failed event

### Requirement: Pure event fold
`internal/progress` SHALL provide a pure fold (`Snapshot`) from an ordered event slice to per-phase display state (phase name, ordered steps with statuses, done/total/failed counts), deterministic for a given input, preserving phase first-appearance order and step first-appearance order within a phase, and counting Done, Skipped, and Failed steps as resolved.

#### Scenario: Fold is deterministic and order-preserving
- **WHEN** the same event slice is folded twice
- **THEN** both results are deeply equal, phases appear in first-mention order, and steps appear in first-mention order within their phase

#### Scenario: Skipped counts as resolved
- **WHEN** a phase's events resolve one step Done, one Skipped, and one Failed
- **THEN** the phase reports done=3 of total=3 with failed=1, so a phase containing a best-effort skip still reaches completion

#### Scenario: Later status supersedes earlier
- **WHEN** a step emits Running and subsequently Done
- **THEN** the folded step's status is Done and the step is not duplicated

### Requirement: Canonical phase placeholders
The fold SHALL support reordering display state onto a canonical phase sequence, inserting an empty pending phase for any canonical phase that has not emitted events, while phases outside the canonical list retain discovery order after the canonical ones.

#### Scenario: Unstarted canonical phase renders pending
- **WHEN** a flow declares the canonical sequence Teardown, Removing worktrees, Prune & cleanup and only Removing worktrees has emitted events
- **THEN** the display state contains all three phases in canonical order with Teardown and Prune & cleanup empty (pending, zero totals)

### Requirement: Consolidation is behavior-neutral
Replacing the previous progress dialects SHALL NOT change any rendered output: golden chrome tests SHALL pass byte-identical without regeneration, and all flow E2E tests SHALL pass unchanged.

#### Scenario: Golden views unchanged
- **WHEN** the golden chrome test suite runs against the consolidated code without `-update`
- **THEN** every pinned view matches byte-for-byte, ANSI included

#### Scenario: Removal flow behavior preserved
- **WHEN** the removal E2E suite exercises deletion including failure and locked-worktree edge cases
- **THEN** all assertions pass without modification to the tests
