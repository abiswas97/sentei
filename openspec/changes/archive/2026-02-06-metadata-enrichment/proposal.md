## Why

The worktree discovery layer (F1) returns basic structural data from `git worktree list --porcelain`, but the TUI (F3) needs richer metadata to render status indicators (🟢🟡🔴🔒), sort by staleness, and warn users before deletion. Without enrichment, there is no bridge between raw discovery and an informed cleanup experience. This is PRD feature F2.

## What Changes

- Extend the `Worktree` struct with enrichment fields: `LastCommitDate` (time.Time), `LastCommitSubject` (string), `HasUncommittedChanges` (bool), `HasUntrackedFiles` (bool), `IsEnriched` (bool)
- Add an `EnrichWorktrees` function that fetches metadata for all worktrees in parallel using the existing `CommandRunner` interface
- Individual git commands per worktree: `git log -1 --format=%ai`, `git log -1 --format=%s`, `git status --porcelain`
- Skip enrichment for bare repo entries and prunable/missing worktrees (mark with enrichment-unavailable state)
- Collect per-worktree errors without failing the batch — partial enrichment is acceptable

## Capabilities

### New Capabilities
- `worktree-enrichment`: Parallel metadata fetching for discovered worktrees — commit date, commit subject, dirty/untracked status, with edge case handling for bare/prunable/missing entries

### Modified Capabilities
- `worktree-discovery`: Add enrichment fields to the Worktree struct (additive, non-breaking)

## Impact

- **Code**: `internal/git/worktree.go` (struct extension), new `internal/worktree/enricher.go` (enrichment logic)
- **APIs**: `Worktree` struct gains new fields (backward compatible — zero values are safe defaults)
- **Dependencies**: None new — uses existing `CommandRunner` and `sync` stdlib
- **Performance**: Must complete in <5s for 50 worktrees via parallel execution
