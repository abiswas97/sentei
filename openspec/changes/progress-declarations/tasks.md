## 1. Vocabulary extensions

- [ ] 1.1 Add `Checkpoint`/`Of` fields to `progress.Event` and the phase-close marker; extend the fold to track closed flags and per-step reached/declared checkpoints (clamped, monotonic, resolution reaches final checkpoint)
- [ ] 1.2 Add `Plan`/`PlannedPhase`/`PlannedStep` and `Declare(plan, emit)` compiling to the Pending burst + close markers (Open phases close later via explicit emit)
- [ ] 1.3 Add `PhaseState.Settled()` (closed && all declared steps resolved) as the single done predicate
- [ ] 1.4 Invariant + property tests: totals monotonic, done <= total, checkpoints never regress, undeclared phase never completes, event-after-close flagged in tests, random interleavings yield monotonic non-decreasing overall fill

## 2. Flow plan seeding (one commit each, table-driven test each)

- [ ] 2.1 Integration apply declares staged-changes x worktrees; verify phases never reopen across a 2x2 apply (success, failure, skip permutations)
- [ ] 2.2 Teardown declares its artifact-removal count; verify the displayed total is real from first frame
- [ ] 2.3 Worktree removal declares per-worktree steps with start/finish checkpoints (resolve the 2-vs-3 checkpoint question against what the deleter honestly observes); verify nonzero checkpoint progress with all steps Running
- [ ] 2.4 Repo create/clone/migrate declare known step lists; genuinely undiscoverable phases stay Open and close on completion paths

## 3. Rendering integration

- [ ] 3.1 `ProgressLayout` overall bar fill derives from checkpoint counts; phase headers keep step counts
- [ ] 3.2 ✦/collapse/green derive from `Settled()`; remove any `done == total` styling paths
- [ ] 3.3 E2E: terminal frames of every flow show each phase settling exactly once; no settled phase re-renders in-progress

## 4. Verification

- [ ] 4.1 Full gauntlet; golden tests regenerated only where honesty changes rendering (bar fill, ✦ timing) with frame-by-frame review of re-recorded playground tapes
- [ ] 4.2 Re-record final_motion-style real-repo tape; frame-verify monotonic bar and single-✦-per-phase via the frame-hash segmentation method
- [ ] 4.3 Update `.impeccable.md` decision log (declarations, checkpoint bar, settled predicate)
