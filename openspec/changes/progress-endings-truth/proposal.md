# Progress Endings Tell the Truth

## Why

Audit P0s: the create flow exits at a 33% bar with `Dependencies`/`Integrations` reading "pending" forever (no-work phases count as outstanding and no final spring sync fires at completion), and quitting mid-progress leaves zero terminal trace of what was created. The summary's `cd` line, the flow's one actionable artifact, truncates into uselessness.

## What Changes

- `ProgressLayout` gains completion semantics: when a flow's result exists, phases that never discovered work render as dim `– skipped` and stop counting as outstanding, so the bar's final target is genuinely 100%.
- Final spring syncs at `createCompleteMsg`, `repoDoneMsg`, and `integrationFinalizedMsg` (the class fixed for removal in #64).
- Quitting from a live progress view prints a one-line warning to stderr naming the interrupted operation.
- The create summary prints the full worktree path, wrapped, never truncated.
- Deferred (recorded): pre-listing repo-flow steps with fixed denominators needs a pipeline plan-event type; follow-up noted in the design.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `tui-chrome`: progress-layout requirement gains the skipped-phase state and the completion guarantee.

## Impact
progress_layout.go, the three flow completion branches, model.go (InterruptedFlow), main.go, create_summary.go.
