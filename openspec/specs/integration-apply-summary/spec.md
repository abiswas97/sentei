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

### Requirement: Presence-based tool detection
The default integration detect probe SHALL test PATH presence (`command -v <binary>`) rather than invoking the tool with a flag it may not implement. An integration MAY still declare an explicit `Detect.Command` that overrides the default. A tool already installed by any manager (pipx, uv, brew, manual) SHALL be detected and its install steps skipped.

#### Scenario: Existing install is detected regardless of CLI flags
- **WHEN** the integration's binary is on PATH but does not implement `--version`
- **THEN** detection SHALL succeed and the install and dependency steps SHALL be skipped

#### Scenario: Missing tool still installs
- **WHEN** the binary is not on PATH
- **THEN** detection SHALL fail and the install step SHALL run

### Requirement: Bounded failure rendering
A failed step's error SHALL render as a bounded peek, never raw: in summaries, at most three lines — the error's first line (dim), its last non-empty line (error color), and a dim `… N more — ? for full output` marker when lines were elided — each truncated to the view width. In live progress rows the error SHALL clamp to a single truncated line. The complete untrimmed output SHALL be available in the detail portal.

#### Scenario: Multi-hundred-line tool output stays bounded
- **WHEN** a failed step's error contains an installer's full output
- **THEN** the summary SHALL show at most three lines for that step and the chrome SHALL remain intact

#### Scenario: Full output preserved
- **WHEN** the user opens the detail portal from a summary with failures
- **THEN** the failed step's complete error output SHALL be readable there

### Requirement: Installs are functionally complete
An integration's install command SHALL produce a tool whose default setup command can run: for cocoindex-code that means installing the `embeddings-local` extra so the default local-embedding `ccc index` has its runtime dependencies.

#### Scenario: Fresh ccc install can index
- **WHEN** sentei installs cocoindex-code on a machine without it
- **THEN** the install SHALL include the embeddings-local extra so `ccc index` does not fail on missing modules

### Requirement: Install resolution is environment-independent
Integration install commands SHALL NOT depend on ambient interpreter pins: the ccc install pins its Python so a `.python-version` in the working directory cannot make resolution unsatisfiable.

#### Scenario: Caller directory pins an old Python
- **WHEN** sentei applies the ccc integration from a directory whose `.python-version` is below 3.11
- **THEN** the install SHALL still resolve using the pinned interpreter

