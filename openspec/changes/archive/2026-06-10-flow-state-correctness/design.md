## Context

The Bubble Tea `Model` is the single root of truth for all flows (Elm architecture). Sub-flow state lives in long-lived structs on `Model` (`removeState`, create/integration state) that are constructed once at startup and mutated in place. Nothing resets them between flow entries or runs, so state leaks across lifetimes:

- `removeState` mixes long-lived data (worktree list, sort preferences) with per-run data (`selected`, `deletionResult`, `deletionStatuses`, `teardownResults`, `pruneErr`, `cleanupResult`). `confirm.go` seeds a new `deletionTotal` but `deletionResult.Outcomes` accumulates forever; `progress.go` computes `done := len(Outcomes)` and completes only when `done == deletionTotal`, so a second run reads 200% and never finishes.
- The menu transition to `createBranchView` (`menu.go`) starts the input cursor blinking but never clears `branchInput`, so the previous run's text persists and typing appends to it.
- `integrationFinalizedMsg` returns straight to the list. On success the staged markers are reconciled in memory; on save failure (or per-step `✗` failures) the user is dropped back on a list that still shows `[+] N changes pending` with no indication anything ran.
- `playground.PlaygroundDir` is a package-level fixed path (`$TMPDIR/sentei-playground`) and `Setup` begins with `RemoveAll(PlaygroundDir)`, so two concurrent playground sessions destroy each other's repos.

## Goals / Non-Goals

**Goals:**
- Per-run flow state is provably pristine at the start of every run; per-entry flow state is pristine on every menu entry
- Removal progress percentage and completion semantics are correct for any sequence of runs in one session
- Integration apply always ends in an explicit outcome (summary view), and staged markers always match persisted state afterwards
- Concurrent playground sessions are fully isolated

**Non-Goals:**
- Visual/chrome changes (ui-chrome-unification owns those; the integration summary uses today's summary conventions and gets re-skinned there)
- Changing deletion, creation, or integration business logic
- A generalized state-machine framework — targeted lifetime fixes only

## Decisions

### D1: Split per-run removal state into its own struct, recreated at run start

**Decision**: Extract the per-run fields out of `removeState` into a `removalRun` struct (`deletionResult`, `deletionStatuses`, `deletionTotal`, `progressCh`, `teardownResults`, `pruneErr`, `cleanupResult`). Starting a deletion (`confirm.go` enter handler) assigns `m.remove.run = newRemovalRun(selected)`. Selection (`selected` map) is per-entry, cleared when the flow is entered from the menu and after a completed run.

**Alternatives considered**:
- *Reset helper that zeroes fields in place (`resetRemovalRun()`)*: Works, but every future field added to `removeState` silently re-introduces the leak unless the author remembers to extend the helper. The struct boundary makes the lifetime explicit — you cannot forget to reset a field that lives inside the recreated struct.
- *Recreate all of `removeState`*: Throws away the enriched worktree list and the user's sort/filter preferences, forcing a rescan and losing UX state that should survive runs.

**Why struct split**: It encodes the two lifetimes (session-long vs run-long) in the type system instead of in discipline. The compiler keeps future fields honest.

### D2: Reset-on-entry for the create flow at the menu transition

**Decision**: The menu's `case "Create new worktree"` calls a `resetCreateFlow()` method (clear `branchInput`, restore `baseInput` default, clear staged options) before switching views.

**Alternatives considered**:
- *Reset on exit (summary → menu)*: Leaks if the flow is abandoned via esc at any intermediate step — every exit path must remember to reset.
- *Recreate the whole create sub-struct*: Equivalent outcome; a small targeted reset is clearer here because most create fields (e.g. detected default branch) are session-long.

**Why reset-on-entry**: There is exactly one entry point (the menu case) and many exit points. Resetting at the single entry edge is the smallest honest fix.

### D3: Integration apply ends in a summary view, and staged state is reloaded from disk

**Decision**: `integrationFinalizedMsg` transitions to a new `integrationSummaryView` instead of returning to the list. The summary renders from the already-collected `m.integ.events`: per-worktree outcomes with done/failed indicators, failed steps with their error text, plus the save error prominently when `state.Save` failed. Dismissing the summary (enter) returns to the integration list via `loadIntegrationState()`, so staged markers are rebuilt from persisted state on success and failure alike — pending markers can never survive an apply.

**Alternatives considered**:
- *Inline result banner on the list*: Cheaper, but a partially failed apply across 7 worktrees needs per-step detail, which a banner cannot hold; and it leaves the reconciliation bug (staged map mutated only on the success path) in place.
- *Reuse `migrate_summary.go` machinery*: That view is migrate-specific; forcing integration outcomes through it couples two flows that only superficially rhyme.

**Why summary view**: Matches every other mutating flow in the app (create, remove, cleanup all end in a summary), fixes the silent-failure path structurally, and gives the chrome-unification change one consistent surface to re-skin. The migrate path (`returnView == migrateNextView`) keeps its existing direct hand-off, since migrate has its own summary.

### D4: Playground directory becomes unique per session

**Decision**: `playground.Setup` creates its root via `os.MkdirTemp(os.TempDir(), "sentei-playground-*")` and returns the path; the startup `RemoveAll` of a shared path is deleted. `main.go` removes the directory on clean exit (best effort). `PlaygroundDir` as a fixed exported var is removed.

**Alternatives considered**:
- *Keep fixed path, add a lock file*: Serializes sessions instead of isolating them; a crashed session leaves a stale lock; parallel e2e agents (a real workflow here) would block each other.
- *Fixed path + PID suffix*: Unique in practice but leaks accumulate with no OS contract; `MkdirTemp` is the idiomatic primitive and gets the same `$TMPDIR` hygiene the OS already manages.

**Why MkdirTemp**: Parallel playground sessions are a first-class use case (multi-agent visual testing drove this change). Isolation beats coordination.

## Risks / Trade-offs

- **[Risk] A reset edge is missed and a leak remains** → Mitigation: regression E2E tests drive each flow twice in one session (create → create, remove → remove, apply → apply) and assert pristine second-run state; these tests fail today and pin the fix.
- **[Risk] Removing `PlaygroundDir` breaks tests that referenced the fixed path** → Mitigation: playground tests already create repos under `t.TempDir()`-style isolation or can take the returned path; compile errors surface every caller.
- **[Trade-off] Integration summary adds one keypress to the happy path** → Accepted: apply mutates up to N worktrees and runs external installers; an explicit outcome is worth one enter, and it matches every other mutating flow.
