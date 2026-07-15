## 1. Vocabulary extensions

- [x] 1.1 Add `Checkpoint`/`Of` fields to `progress.Event` and the phase-close marker; extend the fold to track closed flags and per-step reached/declared checkpoints (clamped, monotonic, resolution reaches final checkpoint)
- [x] 1.2 Add `Plan`/`PlannedPhase`/`PlannedStep` and `Declare(plan, emit)` compiling to the Pending burst + close markers (Open phases close later via explicit emit)
- [x] 1.3 Add `PhaseState.Settled()` (closed && all declared steps resolved) as the single done predicate
- [x] 1.4 Invariant + property tests: totals monotonic, done <= total, checkpoints never regress, undeclared phase never completes, event-after-close flagged in tests, random interleavings yield monotonic non-decreasing overall fill

## 2. Flow plan seeding (one commit each, table-driven test each)

- [x] 2.1 Integration apply declares staged-changes x worktrees; verify phases never reopen across a 2x2 apply (success, failure, skip permutations)
- [x] 2.2 Teardown declares its artifact-removal count; verify the displayed total is real from first frame
- [x] 2.3 Worktree removal declares per-worktree steps with start/finish checkpoints (resolved: 2 checkpoints — the deleter observes nothing finer inside `git worktree remove`); verify nonzero checkpoint progress with all steps Running
- [x] 2.4 Repo create/clone/migrate declare known step lists; genuinely undiscoverable phases stay Open and close on completion paths

## 3. Rendering integration

- [x] 3.1 `ProgressLayout` overall bar fill derives from checkpoint counts; phase headers keep step counts
- [x] 3.2 ✦/collapse/green derive from `Settled()`; remove any `done == total` styling paths
- [x] 3.3 E2E: terminal frames of every flow show each phase settling exactly once; no settled phase re-renders in-progress

## 4. Verification

- [x] 4.1 Full gauntlet; golden tests regenerated only where honesty changes rendering (bar fill, ✦ timing) with frame-by-frame review of re-recorded playground tapes
- [x] 4.2 Re-record final_motion-style real-repo tape; frame-verify monotonic bar and single-✦-per-phase via the frame-hash segmentation method
- [x] 4.3 Update `.impeccable.md` decision log (declarations, checkpoint bar, settled predicate)
