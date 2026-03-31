## Why

The current cleanup flow has two UX problems: (1) the confirmation dialog asks "are you sure?" without showing what will happen, and (2) aggressive cleanup is only accessible via CLI flag — TUI users can't upgrade from safe to aggressive when the tool finds branches to clean. Replacing the blind confirmation with an informed preview (dry-run scan → show results → choose action) makes the cleanup flow transparent and gives TUI users access to aggressive cleanup with full visibility into what will be deleted.

Additionally, worktree removal currently has no safety gate for dirty/unpushed branches — removing a worktree with uncommitted changes is irreversible data loss. Adding a confirmation gate specifically for at-risk worktrees protects against accidental loss while keeping the flow fast for clean worktrees.

## What Changes

- Replace the cleanup confirmation dialog (TUI menu path) with a cleanup preview: dry-run scan with loading state → results grouped by safe/aggressive → choose action
- Add aggressive cleanup upgrade offer in TUI: when dry-run detects branches that aggressive mode would clean, show inline preview (first 2-3 names + "and N more") with `a` key to run aggressive
- Wire `?` key from cleanup preview to open detail portal with full branch list + metadata
- Add `--yes` CLI flag to skip confirmation for scripted/CI usage
- Add confirmation gate for worktree removal when selection includes dirty or unpushed branches
- Move CLI command echo from confirmation to summary views

## Capabilities

### New Capabilities
- `cleanup-preview`: Dry-run scan with loading state, results preview grouped by safe/aggressive, aggressive upgrade offer with inline branch preview, and detail portal for full branch list
- `removal-safety-gate`: Confirmation dialog specifically for worktree removal when selection includes dirty (uncommitted changes) or unpushed (commits not on remote) branches

### Modified Capabilities
- `confirmation-view`: CLI command echo moves from confirmation to summary views; confirmation becomes CLI-only for non-destructive ops
- `cli-tui-handoff`: Add `--yes` flag for non-interactive confirmation skip across CLI subcommands

## Impact

- `internal/tui/cleanup_preview.go` — NEW: replaces cleanup confirmation for TUI menu flow
- `internal/tui/cleanup_confirm.go` — MODIFY: CLI-only path, used when launched via `sentei cleanup`
- `internal/tui/confirm.go` — MODIFY: add dirty/unpushed detection and confirmation gate
- `internal/cleanup/` — MODIFY: expose dry-run scan API that returns what each mode would do
- `internal/tui/confirmation.go` — MODIFY: move CLI command echo to summary views
- `cmd/` — MODIFY: add `--yes` flag to cleanup and remove subcommands
- `.impeccable.md` — expand with cleanup flow patterns
