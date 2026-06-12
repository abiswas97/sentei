## Context

`progress-package` consolidated the vocabulary; this change makes it honest. The audit's three measured failures (✦ melting, Teardown's fake total, the 0%-dead bar) share one root cause: `total = len(steps discovered so far)`. The plan for every lying flow is actually known before work starts (staged integrations x worktrees; artifact count; selected worktrees). User decision from the design FAQ: plan-in-stream (declaration is events), with a typed builder for ergonomics; checkpoints per flow as declared data rather than a global started-equals-half heuristic.

## Goals / Non-Goals

**Goals**
- Totals come from declaration; discovery semantics remain only for genuinely undiscoverable work.
- ✦/green/collapse mean "closed and complete", restoring the "anything still is settled" promise.
- The bar moves when real sub-stage boundaries (checkpoints) are crossed, including step starts.

**Non-Goals**
- No timing/hold/settle changes (`progress-completion-settle`).
- No per-checkpoint weighting (equal weights; revisit on a demonstrated need, rule of three).
- No continuous progress (e.g. clone object-transfer percentage) — the checkpoint model leaves room; not used yet.

## Decisions

**D1: Plan-in-stream with a typed compiler.**
`Declare(plan, emit)` emits one Pending event per planned step (creating steps and totals in the fold with no fold rework — first-mention-creates-step flips from bug to feature) and a close marker per closed phase. The stream remains the single source of truth; replay/golden testing keeps working. Alternative — a Plan struct joined with events at fold time — rejected: two artifacts that must agree reintroduce the consistency problem being fixed.

**D2: Close is an explicit event; declared phases default to closed.**
`Declare` closes each phase unless `Open: true` (scan-style phases that append steps after declaration close later via an explicit close emit). Alternative — implicit close when the burst arrives — is the default behavior; the explicit marker exists only for Open phases, keeping the common case zero-ceremony.

**D3: Checkpoints are fields on the existing Event, not a new event type.**
`Event{Checkpoint: k, Of: n}` on a Running event means "reached sub-stage k of n". A step's declared checkpoint count lives in the Pending burst (`Of` on the Pending event). Atomic steps declare 1 checkpoint; start/finish steps declare 2 (start = checkpoint 1). The fold treats a step's resolved status as reaching its final checkpoint. Alternative — a separate CheckpointReached event kind — rejected as a second vocabulary for the same fact.

**D4: Headers count steps; the bar counts checkpoints.**
Human-facing counts (`2/3`) stay in human units. The bar's fill = checkpoints reached / checkpoints declared across all phases, equal weights. This is the resolution split that lets parallel removals move the bar at start without overstating completion in the counts.

**D5: ✦ gate is `closed && done == total` in one derivation.**
`PhaseState.Settled()` is the single predicate used by collapse, ✦ rendering, and green styling, so the three can never disagree (mirrors the cleanup truth-chain pattern).

## Risks / Trade-offs

- [Open phases forget to close, ✦ never appears] → invariant test: every flow's plan either closes all phases at declaration or the flow's completion path emits the close; E2E asserts terminal views show settled phases.
- [A flow emits a step not in its declaration] → fold accepts it (total grows; monotonicity preserved) but a debug-mode invariant flags it in tests, keeping production forgiving and tests strict.
- [Bar regressions from miscounted checkpoints] → fold clamps: reached never exceeds declared; property test drives random event interleavings and asserts monotonic non-decreasing fill.
- [Checkpoint declarations drift from emitter reality] → declarations and emissions live in the same flow package, adjacent by construction; the table-driven flow tests assert final reached == declared.

## Migration Plan

1. Land Plan/Declare + Event fields + fold semantics + invariants in `internal/progress` (no flow changes; undeclared flows behave exactly as today).
2. Seed plans flow by flow, each its own commit with its table-driven test: apply, teardown, removal, repo flows.
3. Switch `ProgressLayout` bar fill to checkpoint counts and ✦/collapse to `Settled()`; re-record the playground tapes and frame-verify (last progress frame of each flow shows monotonic bar, single ✦ per phase).

Rollback: flows revert to undeclared independently; the fold's discovery path is the unchanged base case, not a legacy branch.

## Open Questions

- Whether removal's per-worktree step should declare 2 checkpoints (start/finish) or 3 (start, removed, teardown-of-artifacts) — decide when seeding, based on what the deleter can honestly observe.
