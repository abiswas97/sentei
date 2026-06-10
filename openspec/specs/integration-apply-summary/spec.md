# integration-apply-summary Specification

## Purpose
TBD - created by archiving change flow-state-correctness. Update Purpose after archive.
## Requirements
### Requirement: Apply ends in an outcome summary
The TUI SHALL transition from the integration progress view to an integration summary view when the apply finishes, showing per-worktree outcomes: succeeded steps with the done indicator and failed steps with the failed indicator and their error text, plus overall counts.

#### Scenario: Fully successful apply
- **WHEN** an apply completes with all steps succeeding across 3 worktrees
- **THEN** the summary SHALL show each worktree with its completed steps and a success headline, and offer a key hint to return to the integration list

#### Scenario: Partially failed apply
- **WHEN** an apply completes with failed steps in 2 of 7 worktrees
- **THEN** the summary SHALL show the failed steps with their error text under their worktrees, and the headline SHALL state how many steps failed

#### Scenario: State persistence failure
- **WHEN** the apply finishes but persisting the integration state to disk fails
- **THEN** the summary SHALL display the save error prominently and indicate that the shown integration set was not persisted

### Requirement: Staged markers are reconciled from persisted state after apply
The TUI SHALL rebuild the integration list's active/staged markers from persisted state when the user dismisses the apply summary, on success and failure alike, so pending markers (`[+]`/`[-]`) never survive an apply.

#### Scenario: List after successful apply
- **WHEN** the user dismisses the summary after a successful apply that enabled two integrations
- **THEN** the integration list SHALL show both integrations as active (`[x]`) with no pending markers and no pending-changes counter

#### Scenario: List after failed persistence
- **WHEN** the user dismisses the summary after an apply whose state save failed
- **THEN** the integration list SHALL reflect what is actually on disk, not the attempted set

### Requirement: Migrate flow hand-off is unchanged
The integration apply summary SHALL NOT be inserted into the migrate flow; when the apply was launched from migration, the existing transition to the migrate flow's own summary remains.

#### Scenario: Apply during migration
- **WHEN** an integration apply completes as part of the migrate flow
- **THEN** the TUI SHALL proceed to the migrate summary as it does today

