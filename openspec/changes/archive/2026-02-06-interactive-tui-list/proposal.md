## Why

The discovery and enrichment layers are complete — we can list worktrees and fetch their metadata. But the tool currently just prints to stdout. Users need an interactive TUI to visually browse worktrees, see status at a glance, multi-select stale ones, and trigger deletion — which is the core value proposition of wt-sweep (PRD F3-F6).

## What Changes

- Add a Bubble Tea TUI application with a scrollable worktree list
- Display columns: branch name, last activity (relative time), last commit subject, ASCII status indicators (`[ok]`/`[~]`/`[!]`/`[L]` for clean/dirty/untracked/locked)
- Keyboard navigation: j/k, arrows, page up/down
- Multi-select with spacebar, select-all/deselect-all shortcuts
- Confirmation dialog before deletion, with warnings for dirty/untracked worktrees
- Parallel worktree deletion with real-time progress bar and per-worktree status
- Post-deletion summary showing successes and failures
- Filter out the bare repository entry from the selectable list
- Replace current `main.go` stdout printing with the TUI app
- Add Bubble Tea, Bubbles, and Lip Gloss dependencies

## Capabilities

### New Capabilities
- `tui-list-view`: Interactive scrollable list displaying enriched worktrees with status indicators, keyboard navigation, and multi-select
- `tui-confirmation`: Confirmation dialog showing selected worktrees with safety warnings before deletion
- `worktree-deletion`: Parallel worktree removal via `git worktree remove --force` with progress tracking and error collection
- `tui-progress`: Real-time progress view during deletion with per-worktree status and post-deletion summary

### Modified Capabilities
_None — existing discovery and enrichment specs are unchanged._

## Impact

- **New packages**: `internal/tui/` (model, list, confirm, progress, styles, keys)
- **Modified files**: `main.go` (replace stdout printing with Bubble Tea program)
- **New dependencies**: `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/bubbles`, `github.com/charmbracelet/lipgloss`
- **New package**: `internal/worktree/deleter.go` for parallel deletion logic
- **Consumed APIs**: `git.ListWorktrees`, `worktree.EnrichWorktrees`, `git.CommandRunner`
