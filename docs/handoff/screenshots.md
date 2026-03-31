# Screenshot Reference

Screenshots from the manual testing session (2026-03-30/31).
Original images are in CleanShot media directory.

## Image 8 — Worktree Removal Progress (STUCK)
- Path: `CleanShot 2026-03-31 at 00.14.45@2x.png`
- Shows: "Removing Worktrees" with purple header, "Removing worktrees 0%", feature/integration-selection listed with pending dot, "Prune & cleanup pending"
- Issue: Stuck at 0%, Ctrl+C didn't work (now fixed), visual style differs from integration progress

## Image 9 — Remove Worktree List (STALE after creation)
- Path: `CleanShot 2026-03-31 at 00.16.53@2x.png`
- Shows: "Remove Worktrees" list with feature/integration-selection selected, but missing the newly created feature/test-creation
- Issue: Worktree list not refreshed after creation (now fixed with dirty flag)

## Image 10 — Remove Worktree List (after restart — correct)
- Path: `CleanShot 2026-03-31 at 00.17.27@2x.png`
- Shows: Correct list with feature/test-creation visible after restart

## Image 11 — Cleanup Confirmation Dialog
- Path: `CleanShot 2026-03-31 at 00.17.44@2x.png`
- Shows: Blue bordered dialog "Confirm Cleanup" with Mode: safe, Dry run: no, CLI command echo
- Question: Should this show in TUI menu flow? User already chose "Cleanup & exit"

## Image 12 — Cleanup Complete Result
- Path: `CleanShot 2026-03-31 at 00.18.08@2x.png`
- Shows: "Cleanup Complete" with bullet points: No stale remote refs, No config duplicates, etc. "Pruned 1 stale worktree" with green dot
- Issue: Could be more visually polished, no progress bar

## Image 13 — Create Worktree Progress
- Path: `CleanShot 2026-03-31 at 00.18.55@2x.png`
- Shows: "Creating Worktree" with Setup 100% ●, Dependencies 100% ●, Integrations 66% with steps listed
- Issue: Green dots without progress bar, inconsistent with integration progress style

## Image 14 — Remove Worktree List (STALE after deletion)
- Path: `image-cache/.../14.png`
- Shows: Deleted worktree still visible in list after returning from deletion
- Issue: Worktree list not refreshed after deletion (now fixed with dirty flag)

## Image 15 — Remove Worktree List (after restart — correct)
- Path: `image-cache/.../15.png`
- Shows: Deleted worktree correctly gone after restart

## Reference: Integration Progress (GOOD — target style)
- The "Applying Integration Changes" view is the visual benchmark
- Grouped by worktree, status indicators, progress bar at bottom with percentage
