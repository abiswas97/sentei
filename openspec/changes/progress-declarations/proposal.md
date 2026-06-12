## Why

The June 2026 UX audit proved the progress system lies during real runs: ✦ phases reopen and run backward (`✦ 1/1 100%` melts to `1/2 50%` because step totals are discovered per-event though the plan is fully known), Teardown renders a placeholder `0/1` then jumps to `4/4` in one frame, and the overall bar sits at 0% for ~85% of a parallel removal because only completion events exist. Totals derived from discovery instead of declaration is the root cause of all three.

## What Changes

- `internal/progress` gains plan declaration: a typed `Plan` (phases > steps > checkpoint counts) that compiles into the event stream as a Pending burst plus per-phase close markers — the stream stays the single source of truth.
- New event semantics: optional `Checkpoint/Of` fields on Running events for intra-step progress; a phase-close marker meaning "no more steps will be added here".
- Fold invariants, enforced by tests in the package: a phase's total never decreases; done never exceeds total; checkpoints are monotonic per step; no step event may follow its phase's close; an undeclared phase stays pending, never 100%.
- Completion semantics change: phase collapse and the ✦/green done treatment derive from `closed && done == total`, never `done == total` alone.
- The overall progress bar fill derives from checkpoints reached over checkpoints declared; phase headers continue to count steps.
- Flows seed real plans: integration apply declares staged-changes x worktrees upfront (fixes phase reopening); teardown declares its artifact-removal count upfront (fixes the 0/1 lie); worktree removal declares per-worktree steps with start/finish checkpoints (fixes the dead bar during parallel work); repo create/clone/migrate declare the step lists they already know, leaving genuinely undiscoverable phases open.

## Capabilities

### New Capabilities
- `progress-declarations`: plan declaration, checkpoint events, close markers, and the honesty invariants binding them.

### Modified Capabilities
- `progress-events`: the fold gains closed/checkpoint state and the new completion semantics (delta against the spec landed by the `progress-package` change).
- `tui-progress`: the overall bar's fill source changes from completed-steps to declared checkpoints, and done-styling of a phase requires the phase to be closed.

## Impact

- Affected code: `internal/progress` (Plan, Declare, Event fields, fold, invariant tests), `internal/integration` (plan seeding), `internal/creator` (teardown plan), `internal/worktree` (removal checkpoints), `internal/repo` (plan seeding), `internal/tui` (ProgressLayout bar source, ✦ gating).
- Depends on `progress-package` having merged.
- Visual change is honesty-only: bars move earlier and never regress; ✦ appears once per phase, when true.
