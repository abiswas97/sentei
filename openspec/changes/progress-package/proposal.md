## Why

Progress reporting has fragmented into four parallel vocabularies: `internal/pipeline` (repo/creator flows), `integration.ManagerEvent` (apply), `worktree.DeletionEvent` plus bespoke tui messages (removal), and `internal/progress.Tracker`, which has zero importers and is dead code. Two packages define identically-named `StepStatus` enums. Every honesty fix the June 2026 UX audit demands (upfront totals, checkpoint progress, phase-close gating) would otherwise have to be implemented three times; consolidating first means implementing them once.

## What Changes

- Rebirth `internal/progress` as the single package owning the progress contract: status vocabulary, `Event`, the event fold (today's `buildPhaseDisplays` in `internal/tui`), and the runner helpers (`RunStep`, `PhaseRecorder`) absorbed from `internal/pipeline`.
- Delete the dead `progress.Tracker` (zero importers).
- Delete `internal/pipeline`; all importers (~20 files across `internal/repo`, `internal/creator`, `internal/tui`, `cmd`) move to `internal/progress` mechanically.
- Replace `integration.ManagerEvent` with `progress.Event` at the emitter (worktree name maps to phase at emit time); the dialect type is removed, not adapted.
- Replace `worktree.DeletionEvent` and the bespoke removal tui messages (`worktreeDeleteStartedMsg`, `worktreeDeletedMsg`, `worktreeDeleteFailedMsg`) with `progress.Event` at the emitter; the dialect types are removed.
- Move `buildPhaseDisplays`/`withPendingPhases` from `internal/tui` into `internal/progress` as the exported fold (`Snapshot`); `internal/tui` consumes the exported display state.
- **Behavior-neutral**: rendered output is unchanged. Golden chrome tests must pass byte-identical without regeneration; all E2E tests unchanged.

## Capabilities

### New Capabilities
- `progress-events`: the consolidated progress contract — one status vocabulary, one `Event` type emitted by every multi-phase flow (removal, apply, teardown, create, clone, migrate, cleanup), and one pure fold from event stream to per-phase display state. Defines the determinism and ordering guarantees later changes build on (declarations, checkpoints, close gating arrive in `progress-declarations`, not here).

### Modified Capabilities

<!-- none: this change is behavior-neutral; tui-progress requirements (windowing, real-time updates, Cmd-chained consumption) hold unchanged and their spec text does not name the dialect types being replaced -->

## Impact

- Affected code: `internal/progress` (rewritten), `internal/pipeline` (deleted), `internal/integration` (ManagerEvent removed), `internal/worktree` (DeletionEvent removed), `internal/tui` (phase_display.go moves out; progress.go, integration_progress.go, migrate_integrations.go, model.go rewired), `internal/repo`, `internal/creator`, `cmd` (import renames).
- No CLI surface change, no rendered-output change, no new dependencies.
- Risk concentrated in the removal flow (destructive path); mitigated by the byte-identical golden criterion and the existing E2E suite (PRD 9.1 edge cases: missing worktree directories, locked worktrees, permission failures all flow through the replaced events and must keep their behavior).
- Unblocks: `progress-declarations` (plan bursts, checkpoints, ✦ close gating), `progress-truth-polish`.
