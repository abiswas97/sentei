## Context

Four progress representations coexist: `internal/pipeline` (Event stream + RunStep/PhaseRecorder, used by repo/creator flows and folded by `internal/tui/phase_display.go`), `integration.ManagerEvent` (apply + migrate-integrations, pumped through channels into tui messages), `worktree.DeletionEvent` plus bespoke tui messages (removal), and `internal/progress.Tracker` (dead, zero importers). The June 2026 UX audit's honesty fixes (upfront totals, checkpoints, close-gated ✦) need a single place to land. This change creates that place and is deliberately behavior-neutral: it must not change a single rendered byte.

Constraints: the removal flow is the destructive path; golden chrome tests pin five stable views byte-exact; teatest E2E covers the flows end to end; the project forbids shims and dead-path flags (dialect types are replaced at the emitter, never wrapped).

## Goals / Non-Goals

**Goals**
- One status vocabulary, one `Event` type, one exported fold, one runner-helper set, in `internal/progress`.
- Zero behavior change: golden tests pass without `-update`; E2E unchanged.
- Dead `Tracker` and `internal/pipeline` deleted; `ManagerEvent`/`DeletionEvent` deleted.

**Non-Goals**
- No plan declarations, checkpoints, phase-close events, or ✦-gating changes (that is `progress-declarations`).
- No timing/settle changes (`progress-completion-settle`).
- No copy or layout changes.

## Decisions

**D1: Consolidate under `internal/progress`, delete `internal/pipeline`.**
The package names the domain (progress of multi-phase operations), not one mechanism (a pipeline). The runner helpers move in unchanged. Alternative considered: consolidate under `pipeline` and delete `progress` — rejected because the domain term outlives the mechanism and the next change adds declaration types that are not "pipeline" concepts.

**D2: Replace dialects at the emitter; no adapters.**
`integration.Manager` emits `progress.Event` directly, mapping worktree name to `Event.Phase` at emit time (the fold already groups by phase, so today's per-worktree phase rendering is preserved). `worktree` deletion emits `progress.Event` with phase `Removing worktrees` and step = path; the tui translates events to its messages in one place instead of three message types. Alternative: adapter functions wrapping old types — rejected per the no-shims rule; adapters would preserve the duplicated vocabulary the change exists to delete.

**D3: The fold moves to the package as `Snapshot(events) []PhaseState`.**
`buildPhaseDisplays` and `withPendingPhases` move from `internal/tui/phase_display.go` to `internal/progress`, exported, with their tests. `PhaseState`/`StepState` are the exported forms of `phaseDisplay`/`stepDisplay`. The tui keeps zero folding logic; `ProgressLayout` consumes `[]PhaseState`. Alternative: keep the fold in tui and export only types — rejected; the fold is the contract's semantics and belongs beside its invariant tests.

**D4: One `StepStatus` enum, values preserved.**
`pipeline.StepStatus` becomes `progress.StepStatus` with identical ordering (Pending, Running, Done, Failed, Skipped) so any persisted or switch-based logic is a mechanical rename. The dead tracker's enum is deleted with it.

**D5: Event field names follow the existing `pipeline.Event` exactly** (`Phase, Step, Status, Message, Error`), minimizing the rename diff. New fields wait for `progress-declarations`.

## Risks / Trade-offs

- [Removal flow regression while rewiring its three message types] → the byte-identical golden criterion plus the existing removal E2E suite are the acceptance gate; the rewire is a separate task from the mechanical renames so it reviews in isolation.
- [ManagerEvent's worktree→phase mapping changes grouping subtly] → integration progress already renders per-worktree phases; a table-driven test asserts the fold of mapped events reproduces today's `buildIntegrationPhases` output for the same scenario.
- [Large mechanical diff hides a semantic edit] → commits separate "rename imports" (no logic) from "replace emitters" (logic); reviewers diff the latter only.
- [Two `StepStatus` enums diverge silently today] → consolidation removes the risk class; an invariant test pins the enum ordering.

## Migration Plan

1. Create `internal/progress` with vocabulary + fold + runner helpers (copied from pipeline + tui, tests moved along); delete `tracker.go`.
2. Mechanical import rename across ~20 files; delete `internal/pipeline`.
3. Replace `ManagerEvent` at the emitter and consumers.
4. Replace `DeletionEvent`/bespoke removal messages at emitter and consumers.
5. Gauntlet (gofmt, go vet, go test -race, golangci-lint) + golden tests without `-update` at every step.

Rollback: each step is a commit; the change is internal-only, so revert is clean at any boundary.

## Open Questions

- None blocking. The exported `PhaseState` field set should anticipate `progress-declarations` (closed flag, checkpoint counts) but not include them; reviewer should confirm the field names leave room (e.g. avoid claiming `Total` for anything other than step count).
