## Why

Live playground testing (2026-06-10) surfaced four correctness bugs in flows the upcoming UI work builds on: a second removal run shows a 200% counter and hangs forever, the create flow reopens with the previous run's branch name still in the input (and will happily create the concatenated branch), integration apply ends with no summary so a failed apply is indistinguishable from one that never ran, and concurrent playground sessions clobber each other because they share one fixed temp directory. Polishing the TUI chrome on top of flows that hang or lie about their state is wasted effort, so these fixes land first.

## What Changes

- Reset all removal-flow run state (`deletionResult`, `deletionStatuses`, `selected`, teardown/prune/cleanup results) when a new deletion starts, fixing the 200% counter, the never-completing progress view, the stale "2 selected" footer, and the prematurely green "Prune & cleanup" phase
- Reset create-flow inputs and staged options when entering the flow from the menu, so every entry starts pristine
- Add an integration apply outcome summary: after apply, show what succeeded and what failed (with step errors) before returning to the list, and refresh staged markers from persisted state on both success and failure so `[+]`/`[-]` pending markers never survive an apply
- Give each playground session a unique temp directory (`MkdirTemp`) instead of the fixed shared `sentei-playground` path, so concurrent sessions (parallel e2e agents, two terminals) cannot interfere

## Capabilities

### New Capabilities
- `flow-state-reset`: Entering any flow from the menu, or starting a new run of a flow, yields pristine flow state — no inputs, selections, counters, or results carried over from a previous run
- `integration-apply-summary`: Post-apply outcome reporting — a summary of applied and failed integration steps shown before returning to the list, with staged markers reconciled against persisted state

### Modified Capabilities
- `tui-progress`: Removal progress percentage and completion detection are defined against the current run only; carried-over outcomes must not inflate the counter or block completion
- `playground-setup`: Playground directory is unique per session instead of a fixed shared path

## Impact

- `internal/tui/model.go` — flow-state reset helpers on `Model`; menu transitions call them
- `internal/tui/menu.go` — create/remove entry transitions reset their sub-state
- `internal/tui/confirm.go` — deletion start resets run state before seeding statuses
- `internal/tui/progress.go` — no rendering change required once state resets correctly (covered by regression tests)
- `internal/tui/integration_progress.go`, new `internal/tui/integration_summary.go` — apply summary view and truthful staged reconciliation
- `internal/playground/setup.go` — `MkdirTemp`-based unique directory; callers updated
- No new dependencies, no CLI flag changes
