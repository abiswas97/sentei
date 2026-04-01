## Context

Changes 1 and 2 provide the foundation: shared chrome, progress layout, animation buffer, detail portal, and contextual `?` key. This change uses all of those to redesign the cleanup flow and add a safety gate to worktree removal.

The current cleanup flow is: menu → confirmation (blind) → run → result. The new flow is: menu → dry-run scan (with loading) → preview (safe results + aggressive offer) → run → result.

## Goals / Non-Goals

**Goals:**
- Cleanup preview shows exactly what will happen before any mutation
- TUI users can upgrade to aggressive cleanup when the scan reveals cleanable branches
- Full branch details available via `?` portal without cluttering the summary
- Dirty/unpushed worktree removal requires explicit confirmation
- `--yes` flag for CI/scripted usage

**Non-Goals:**
- Changing cleanup business logic (safe vs aggressive rules stay the same)
- Adding new cleanup operations (e.g., cleaning up tags, remotes)
- Interactive branch selection within aggressive cleanup (it's all-or-nothing per mode)

## Decisions

### D1: Dry-run scan as a separate API

**Decision**: Add a `cleanup.DryRun(runner, repoPath, opts) DryRunResult` function that runs the same logic as `cleanup.Run` but only counts/collects what would happen. `DryRunResult` contains both safe and aggressive results so the preview can show both.

**Alternatives considered**:
- *Run cleanup with DryRun=true flag*: The current `Options.DryRun` field already exists but doesn't return structured results — it just skips mutations. We need structured data for the preview.
- *Two separate dry-runs (safe then aggressive)*: Wasteful — scanning the repo twice. Better to scan once and return results for both modes.

**Why separate function**: Clean separation. The existing `Run` function stays unchanged. `DryRun` returns a structured result optimized for display. Both share the same underlying scan logic (extract into shared helpers).

### D2: Cleanup preview as a new view state, not modification of confirmation

**Decision**: Add a new `cleanupPreviewView` state and `cleanup_preview.go` file. The existing `cleanupConfirmView` continues to serve the CLI path (`sentei cleanup --mode=X`).

**Alternatives considered**:
- *Modify existing confirmation to conditionally show preview*: Mixes two concerns (CLI confirmation vs TUI preview) in one view, making both harder to test and maintain.

**Why new view**: Each view has one clear responsibility. CLI confirmation shows "you asked for X, confirm?" TUI preview shows "here's what's available, what do you want?"

### D3: Dirty/unpushed detection at confirmation time

**Decision**: When the user presses Enter on the removal selection, check if any selected worktree has `HasUncommittedChanges`, `HasUntrackedFiles`, or is not pushed to remote. If so, show a confirmation view listing the at-risk worktrees. If all clean and pushed, proceed directly to removal.

**Alternatives considered**:
- *Always confirm removal*: Adds friction to the common case (clean worktrees) without safety benefit.
- *Warn inline in the list view*: The list already shows status indicators, but users might not notice them when bulk-selecting.

**Why gate at confirmation**: It's the last chance before irreversible action. The gate shows a focused view of specifically what's at risk, not just status indicators in a long list.

### D4: `--yes` flag on CLI subcommands

**Decision**: Add `--yes` / `-y` flag to `sentei cleanup` and `sentei remove` subcommands. When set, skip confirmation and proceed immediately. Does NOT skip the TUI confirmation gate for dirty worktrees — `--yes` means "I've reviewed the flags I passed," not "delete my uncommitted work."

**Alternatives considered**:
- *`--non-interactive`*: Too broad — this would imply the entire TUI is skipped, which is a different concept.
- *`--force`*: Implies overriding safety checks. `--yes` only skips the "are you sure?" prompt.

**Why `--yes`**: Standard Unix convention (apt, yum, etc.). Clear intent — "yes, proceed with what I asked for."

## Risks / Trade-offs

**[Risk] Dry-run scan adds latency to cleanup flow** → Mitigated by the animation buffer (`MinProgressDisplay`). The `◐ Scanning repository…` loading state makes the wait transparent. For most repos this takes <1 second.

**[Risk] Aggressive branch list could be very long** → Mitigated by inline preview (first 2-3 names) + detail portal for the full list. The summary stays compact regardless of branch count.

**[Trade-off] Two cleanup view paths (TUI preview vs CLI confirmation)** → Accepted. Each path serves a different user intent. The code overhead is small since both share the dry-run scan and chrome helpers.
