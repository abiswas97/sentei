## MODIFIED Requirements

### Requirement: Pure event fold
`internal/progress` SHALL provide a pure fold (`Snapshot`) from an ordered event slice to per-phase display state (phase name, ordered steps with statuses, done/total/failed counts, closed flag, and per-step reached/declared checkpoint counts), deterministic for a given input, preserving phase first-appearance order and step first-appearance order within a phase, and counting Done, Skipped, and Failed steps as resolved.

#### Scenario: Fold is deterministic and order-preserving
- **WHEN** the same event slice is folded twice
- **THEN** both results are deeply equal, phases appear in first-mention order, and steps appear in first-mention order within their phase

#### Scenario: Skipped counts as resolved
- **WHEN** a phase's events resolve one step Done, one Skipped, and one Failed
- **THEN** the phase reports done=3 of total=3 with failed=1, so a phase containing a best-effort skip still reaches completion

#### Scenario: Later status supersedes earlier
- **WHEN** a step emits Running and subsequently Done
- **THEN** the folded step's status is Done and the step is not duplicated

#### Scenario: Fold carries declaration state
- **WHEN** a stream contains a declaration burst, a close marker, and checkpointed Running events
- **THEN** the folded phase state reports the declared totals, the closed flag, and per-step checkpoint progress alongside the step statuses
