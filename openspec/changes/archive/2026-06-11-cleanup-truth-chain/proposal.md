# Cleanup Truth Chain

## Why

The audit retake found the cleanup flow contradicting itself end to end: the preview's bold headline promised "2 branches would be deleted" while every candidate was unmerged, `a` then opened a warning-styled confirm for "Delete 0 branches?", the result said "Repository is clean" directly above a skipped-branches warning, the removal tip recommended a command that would do nothing, and an unlabeled duplicate command line sat as residue. Counts are honest everywhere else in the app; this chain was the exception.

## What Changes

- The aggressive headline states the EFFECTIVE count ("1 of 3 branches… would be deleted"; "none deletable without --force" when zero), with the `a` hint and gate active only when deletions are possible.
- The result headline never claims clean beside skips ("Nothing cleaned — N branches remain (unmerged)" / "Cleanup complete — N remain").
- The clean preview shows the five checks (what cleanup covers) instead of a bare clean line; result check order matches the preview.
- The command echo gains a `ran:` label; the removal-summary tip notes that unmerged branches need --force; the scanning view gets its bottom chrome and quit hint; the raw `"a"` key literal becomes a keys.go match.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `cleanup-preview`: effective-count headline, deletable-gated aggressive entry, clean-state check visibility.

## Impact
internal/cleanup/dryrun.go (DeletableAggressiveCount), internal/tui/cleanup_preview.go, cleanup_result.go, summary.go.
