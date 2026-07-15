# Progress Arc Correctness and Live Region Design

**Date:** 2026-07-14
**Status:** Approved

## Problem

The progress arc has a sound event vocabulary, but several producers and the
terminal renderer do not yet uphold it end to end:

1. Conditional integration work is discovered after execution starts, so the
   denominator grows and visible progress can move backward.
2. Early failures can close a phase while declared downstream steps remain
   pending forever.
3. Repository create, clone, and migrate still discover their known steps as
   they execute instead of declaring their plans first.
4. Removal and integration rebuild display state in the TUI rather than using
   the shared `progress.Snapshot` fold, creating multiple sources of truth.
5. The expanding phase tree exceeds an 80x24 terminal and animates pending
   phases as if they were running.
6. The pre-commit identity guard checks repository configuration rather than
   Git's effective author and committer identities.

The result is a progress display that can be numerically dishonest, visually
unstable, and contradictory after failure even though the underlying package
boundaries are otherwise clean.

## Goals

- Freeze an accurate denominator before destructive or long-running work.
- Make every declared step resolve to done, failed, or skipped.
- Keep one event stream and one fold as the progress source of truth.
- Fit all live progress views at 80x24 without clipping the bar or footer.
- Use motion only for work that is actively executing.
- Preserve detailed, auditable failure and skip information.
- Demonstrate the finished behavior with deterministic VHS recordings.

## Non-goals

- Adding cancellation or joining background commands when `q` is pressed.
- Fixing the positional repository argument documented in the README.
- Changing the removal summary's pre-existing `q` behavior.
- Introducing a new TUI or animation dependency.

## Considered Approaches

### Declare the maximum possible plan

Declare every possible conditional step and skip work that is not needed.
This guarantees a stable denominator but counts hypothetical work and creates
noisy skip traces.

### Preflight and freeze the exact plan

Perform read-only discovery first, compile the exact work into the event
stream, then execute that immutable plan. This keeps totals accurate and
retains step-level detail. This is the selected approach.

### Count only coarse phases

Treat each integration, worktree, or repository phase as one progress unit and
show sub-operations as status text. This is robust but discards useful detail
and weakens the existing progress architecture.

## Design

### 1. Exact plans precede execution

Each flow separates planning from execution:

1. Preflight performs read-only discovery and returns an immutable operation
   plan.
2. `progress.Declare` compiles that plan into Pending events and close markers.
3. Execution consumes the same plan without adding step names.
4. Every execution path resolves every declared step.

The existing `progress.Plan`, `PlannedPhase`, and `PlannedStep` remain the
domain contract. Plans continue to ride the event stream so replay, tests, and
the TUI do not need a second artifact to join with events.

The plan owns execution, not only declaration. A concurrency-safe execution
object validates the plan, emits the complete declaration prefix, runs or
resolves named steps, and terminalizes any untouched steps during `Finish`.
This makes undeclared work and terminal-state mutation contract violations
instead of states the renderer must repair.

Step identity is separate from presentation. Each planned step has a stable ID
that is unique within its phase and a display label. Events fold by ID and show
the label. This prevents repeated labels such as `Copy index from main` from
collapsing into one step.

Integration preflight determines, per worktree and integration:

- whether the tool is already available;
- which dependencies require installation;
- whether the tool installation step will run;
- the setup, teardown, and artifact-removal steps that will run.

Detection results are carried in the plan and reused by execution. Execution
must not repeat detection and produce a different plan.

Because the supported integration installers are global tools, their dependency
and install work appears once in a `Prerequisites` phase. Per-worktree phases
contain setup, teardown, and artifact-removal work. This also prevents the first
worktree's installation from invalidating identical plans for later worktrees.
The TUI shows an indeterminate `Preparing plan...` state while the read-only
probes run; percentages begin only after the exact plan exists.

Repository create, clone, and migrate declare the steps already determined by
their options. A step whose applicability depends on an earlier result may be
declared conservatively and resolved as skipped with a reason, but no new step
may be introduced after its phase closes.

### 2. Terminal-state guarantee

A closed phase must be settled when its driver completes. If a step fails,
later dependent steps are emitted as `StepSkipped` with a concise reason such
as `blocked by installation` or `blocked by worktree creation`.

Independent cleanup work may continue after a failure when that is already the
flow's safety behavior. Only dependency relationships cause skips.

The progress package gains a small helper for resolving the untouched suffix
of a declared plan as skipped, plus a final execution safety net that resolves
any remaining pending step. Producer tests validate the stream with
`ValidateStream` and assert that every phase is settled after completion.

### 3. One fold for every progress view

`progress.Snapshot` remains the only event-to-display fold. `StepState` retains
the stable ID, display label, and structured error needed by the TUI so
integration progress no longer reimplements folding to decorate errors.

Removal emits declarations, checkpoints, terminal events, and close markers
for teardown, worktree removal, prune, and cleanup. Its TUI-specific state may
still control concurrency and summaries, but `buildRemovalPhases` is replaced
by folding the emitted event history.

Integration progress similarly folds its event history directly, then applies
only presentation-safe phase labels such as worktree basenames. Presentation
must not recompute status, counts, or closure.

### 4. Pinned live region

`ProgressLayout` receives the raw terminal height and divides it into fixed
vertical regions:

1. title and optional subtitle;
2. separator;
3. compact completed history;
4. one active or failed phase detail region;
5. a static waiting summary;
6. separator, progress bar, and footer.

The pure viewport projection returns bounded history, one focus phase, and a
queued summary. The bar and footer remain at stable rows during a run by padding
unused live-region rows. Completed phases collapse to one line. The active
region receives the remaining line budget and windows its steps, prioritizing
the current running step, the latest failure, recent completions, and the next
pending work in that order. If even those exceed the budget, one stat line
reports the omitted counts. Pending phases render as a static summary and never
receive an active glyph or shimmer.

At 80x24, history and the waiting summary compact before the active region,
bar, or footer can be displaced. At smaller heights, the layout preserves the
title, active/failure line, bar, and footer, and replaces omitted detail with a
single stat line.

The projection has three resize tiers: normal at 18 or more rows, compact at
12-17 rows, and minimal below 12 rows. Width pressure removes elapsed time
before shrinking the useful bar, then falls back to percentage-only at truly
small widths. Full failed-step output and omitted history remain available in
the existing `bubbles/viewport` details portal.

### 5. Motion and Charm primitives

The existing Charm stack remains sufficient:

- `bubbles/progress` continues to spring between monotonic targets;
- Lip Gloss measures and truncates cells and composes the fixed regions;
- the existing deterministic motion clock drives the active star and shimmer;
- `bubbles/help` continues to render footer bindings.

Motion follows one rule: anything moving is being worked on; anything still is
waiting or settled. Only the currently executing phase and step twinkle and
shimmer. Phase collapse is an instant state change because terminal row motion
would compete with the pinned layout. The completion beat remains the single
hero moment: the bar settles, changes to the success palette, and the active
star crystallizes.

No new decorative spinner or transition is added. Tests use static motion
fixtures, and non-interactive or reduced-motion contexts retain meaningful
static glyphs and an immediately correct final frame. `SENTEI_MOTION=off` and
`TERM=dumb` disable star, shimmer, and spring ticks without disabling elapsed
time or delaying settle; a settings UI is deferred.

### 6. Effective Git identity guard

The pre-commit hook reads `git var GIT_AUTHOR_IDENT` and
`git var GIT_COMMITTER_IDENT`, which include environment overrides and therefore
match the identities Git will actually write. It rejects known test markers
such as the `.invalid` domain and `sentei-test`, while allowing legitimate
contributors and automation. It never prints or requires a hard-coded personal
email.

Hook tests cover configured identities, environment overrides, author/committer
differences, and test-identity rejection.

## Verification

### Domain and producer contracts

- Phase IDs and step IDs are nonempty and unique within their scope.
- The declaration burst is the complete stream prefix before work starts.
- Denominator and checkpoint totals are fully established before the first
  Running event.
- Totals never increase during execution and reached checkpoints never regress.
- No producer introduces an undeclared step or mutates a terminal step.
- All closed phases are settled at driver completion.
- Failure resolves dependent suffixes as skipped with a reason.
- Empty and already-installed integration plans remain valid and truthful.
- Concurrent event interleavings preserve totals and terminal states.
- Equal display labels with different stable IDs remain distinct.

### TUI behavior

- Rendered output never exceeds the supplied height at 80x24 and narrower
  fixtures.
- The bar and footer stay present at fixed rows throughout a run.
- Exactly one active phase animates; pending phases are static.
- Failures and their blocked steps remain visible.
- Long skip reasons and error previews stay within the supplied width.
- Resizing between 80x24, 120x40, and 50x16 preserves state and controls.
- Static-motion mode schedules no motion or spring frames and settles normally.
- Success and failure final frames are asserted with `teatest`.
- Existing golden views and the race-enabled full test suite remain clean.

### Demonstrations

VHS tapes run only against playground data or isolated temporary repositories.
Recordings cover:

1. multi-worktree progress at 80x24;
2. successful worktree removal;
3. integration installation failure with the dependent setup step skipped;
4. a concise before/after comparison of the expanding tree and pinned region.

Each tape fixes terminal dimensions, typing delays, and operation timing.
`stty cols 80 rows 24` fixes terminal cells independently of VHS pixel size.
PATH shims under an isolated temporary home provide deterministic command
latency and integration failure without touching a user repository or network.
GIFs and selected PNG frames are inspected after rendering for clipping,
unreadable transient frames, bar regression, and a coherent final frame.

## Risks and Mitigations

- **Preflight result becomes stale before execution.** The interval is short;
  execution treats an already-satisfied install as success or skip rather than
  mutating the plan. Tests cover this idempotent boundary.
- **Plans duplicate producer control flow.** Plan builders own step selection;
  execution iterates plan operations rather than reconstructing conditions.
- **Preflight feels like stalled progress.** A non-percent preparing state gives
  immediate feedback; the determinate bar begins only when its denominator is
  truthful.
- **Pinned history hides useful detail.** Failures stay expanded, while full
  output remains available through the existing details portal and summaries.
- **Small terminals cannot show every row.** The layout preserves active work,
  failures, progress, and controls, then reports omitted counts explicitly.

## Rollout

Land the work as small commits: progress helpers and contracts, integration
planning, creator/repository planning, removal event unification, pinned layout
and motion, identity guard, then demonstrations. Run focused tests after each
commit and the full race/lint/build gauntlet before the final review.
