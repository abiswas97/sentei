# UI Alignment & Progress View Unification

## Context

During manual testing of the CLI/TUI handoff work (PR #22), several UX inconsistencies were identified across sentei's progress views and result screens. The integration management screens are the best reference point — the other views should be brought up to that standard.

## Reference: .impeccable.md Design Principles

- Information hierarchy first — scannable, progressive disclosure
- Color means something — semantic, never decorative
- Keyboard-native — vim-style + standard keys
- Safe by default — destructive actions require confirmation
- Palette: purple accent (62), green (42), yellow (214), red (196), gray (241)

## Issues Found (with screenshots)

### 1. Worktree Removal Progress — Stuck at 0%, Ctrl+C trapped user
**Status: FIXED in PR #22**
- Added Ctrl+C/q handler to all progress views
- But the visual style is still inconsistent with integration progress

### 2. Worktree Removal Progress — Visual mismatch
**Screenshot:** Image 8 — "Removing Worktrees" view
- Shows phase headers + per-worktree status indicators
- No progress bar (unlike integration progress which has ████░░░░ bar)
- Purple header box looks different from other views' `sentei —` title style
- The "Removing worktrees 0%" and "Prune & cleanup pending" text is fine but the visual treatment differs

### 3. Create Worktree Progress — Green dots without progress bar
**Screenshot:** Image 13 — "Creating Worktree" view
- Shows phases (Setup, Dependencies, Integrations) with percentage + green dot
- But no progress bar — just `100% ●` next to each phase
- Inconsistent with integration progress which has the block-character bar
- The green dots (●) feel disconnected without a bar to contextualize them

### 4. Create Summary — Newly created worktree missing from list
**Status: FIXED in PR #22 (dirty flag pattern)**
- After creation, returning to menu now triggers `stateStale` reload

### 5. Cleanup Confirmation View — Should this show in TUI menu flow?
**Screenshot:** Image 11 — "Confirm Cleanup" dialog
- Shows mode, dry-run setting, CLI command echo
- Blue rounded border box looks different from other views
- Question: Is this confirmation useful when the user just selected "Cleanup & exit" from the menu? They already chose to clean up. Maybe only show when launched via `sentei cleanup` (with flags), not from the menu.
- The CLI command echo is great for discoverability but may be noise in the menu flow.

### 6. Cleanup Result Display — Visual improvement needed
**Screenshot:** Image 12 — "Cleanup Complete" view
- Shows bullet points with status indicators (· for nothing happened, ● for pruned)
- Clean and informative but could be more visually polished
- Compare with how integration progress groups results by worktree with status indicators

### 7. Stale state after mutations — Global refresh needed
**Status: FIXED in PR #22 (dirty flag pattern)**
- `stateStale` flag set by all mutations (create, delete, cleanup, integration apply)
- Menu checks flag on entry and reloads lazily
- Analogous to React Query cache invalidation

## Goal: Canonicalize on Integration Progress Style

The integration management screens are the visual benchmark:
- Title: `sentei — Applying Integration Changes`
- Grouped by worktree with indented steps under each
- Status indicators: ● done, ◐ active, ✗ failed
- Progress bar at bottom: `████░░░░ 75%`
- Clean separation between groups

All progress views should converge on this pattern where applicable.

## Proposed Changes

### A. Unified Progress Component
Extract a reusable TUI progress renderer that all views share:
- Title line (`sentei — <Operation>`)
- Groups (worktrees or phases) with steps
- Status indicators (done/active/failed/skipped)
- Progress bar at bottom with percentage
- Consistent keybinding footer

This builds on `internal/progress/tracker.go` which already has the data model.
Add a `View(tracker *progress.Tracker, title string, width int) string` function
that renders the standard progress layout.

### B. Per-View Specific Decisions
1. **Worktree removal** — group by phase (remove → prune → cleanup), show per-worktree status under remove phase
2. **Worktree creation** — group by phase (setup → deps → integrations), show steps under each
3. **Repo operations** — same phase grouping as creation
4. **Integration apply** — already correct, serves as reference
5. **Cleanup result** — could use the same grouped layout with phases (refs → config → branches → prune)

### C. Cleanup Confirmation
Options to consider:
- Remove from TUI menu flow (user already chose "Cleanup & exit")
- Keep it but make it more compact (no border box)
- Only show when launched via `sentei cleanup` with flags

### D. State Management Architecture
**DONE:** Dirty flag pattern implemented. Every mutation sets `stateStale=true`. Menu lazy-reloads on entry.

## Files to Modify

| File | Change |
|------|--------|
| `internal/tui/progress.go` | Use unified progress renderer |
| `internal/tui/create_progress.go` | Use unified progress renderer |
| `internal/tui/repo_progress.go` | Use unified progress renderer |
| `internal/tui/integration_progress.go` | Reference implementation (mostly keep) |
| `internal/tui/cleanup_result.go` | Improve visual layout |
| `internal/tui/cleanup_confirm.go` | Simplify for menu flow |
| `internal/progress/tracker.go` | Add `View()` render method |

## Testing

- Unit tests for the unified progress renderer
- teatest E2E for each progress view with the new layout
- Manual visual verification against .impeccable.md design principles
