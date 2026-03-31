# UI Alignment Session Meta

> **Status:** All 3 changes fully specced (proposal + design + specs + tasks). Ready for implementation.
> **Date:** 2026-03-31
> **Branch:** feature/ui-alignment
> **DO NOT COMMIT** — working context for multi-session continuity

## Changes Overview

Three sequential changes, each depends on the previous:

### Change 1: `ui-chrome-unification` (openspec/changes/ui-chrome-unification/)
**Status:** Implementation complete — all tasks done, 49/50 checked (11.4 manual visual check deferred to user)

**Scope:**
- Shared chrome helpers: `viewTitle()`, `viewSeparator()`, `viewKeyHints()`
- Unified progress layout: title + separator + phases (4-space indent) + bar + hints
- Adaptive windowing for large item lists (terminal-height responsive)
- Stat line with indicator legend: `● 10 done · ◐ 3 active · · 17 pending  showing 6 of 30`
- Summary chrome helpers (shared title/separator/hints, bespoke content middle)
- Animation buffer constants (`minProgressDisplayMs`)
- Key mapping const file (contextual `?` = details, separate key for global help)
- Delete `styleHeader` (white-on-purple bg badge), standardize all titles to `styleTitle`
- `.impeccable.md` expanded with component patterns + design vocabulary
- **No dependencies**

### Change 2: `detail-portal-component` (openspec/changes/detail-portal-component/)
**Status:** Not started

**Scope:**
- `DetailPortal` shared component: overlay (`bubbletea-overlay`) + viewport (`bubbles/viewport`)
- Consistent chrome: title, scroll hints, dismiss key
- Help overlay as first consumer
- **Depends on:** Change 1

### Change 3: `cleanup-preview-redesign` (openspec/changes/cleanup-preview-redesign/)
**Status:** Not started

**Scope:**
- Cleanup preview with dry-run scan + `◐ Scanning repository…` loading state
- Aggressive upgrade offer with inline preview (first 2-3 branch names + "and N more")
- Detail portal for aggressive cleanup details (second consumer of portal)
- Dirty/unpushed confirmation gate for worktree removal
- `--yes` flag for CLI confirmation skip
- Confirmation is CLI-only for non-destructive ops; always-confirm for aggressive cleanup + dirty removal
- **Depends on:** Change 1, Change 2

## Design Decisions (from brainstorming)

### Progress Views
- **Cherry-pick approach:** Phase-based layout from create/repo + progress bar from integration + standardized chrome
- **4-space indent** for steps nested under phase headers
- **Adaptive windowing:** If items > available terminal lines, show sliding window (all active + all failed + last N completed + next M pending). If items fit, show all.
- **Stat line (Option 2):** `● 10 done · ◐ 3 active · · 17 pending  showing 6 of 30` — only shown when windowed
- **Parallel removal:** Multiple `◐` items simultaneously; failed items (`✗`) always pinned visible

### Summary Views
- **Hybrid approach (C):** Shared chrome helpers for title/separator/hints; bespoke content per view
- Remove summary uses `"v"` as success marker — standardize to `●`

### Confirmation
- **CLI-only** for non-destructive ops (safe cleanup, clean+pushed removal)
- **Always confirm** for: aggressive cleanup, removal of dirty/unpushed worktrees
- **`--yes` flag** to skip confirmation in CLI for scripting/CI
- TUI menu flow skips confirmation (user already chose interactively)
- CLI command echo moves to summary view

### Cleanup Flow (Change 3)
- **Preview-first:** Dry-run scan on entry, show results grouped by safe/aggressive
- **Upgrade offer:** If aggressive has work, show `enter safe · a aggressive · ? details · esc back`
- **Detail portal:** `?` opens scrollable overlay with full branch list + metadata
- **Inline preview:** First 2-3 branch names shown in summary, rest via portal

### Portal Component (Change 2)
- `bubbletea-overlay` for compositing + `bubbles/viewport` for scrollable content
- Reusable for: cleanup details, help overlay, future detail views
- `?` is contextual (details for current view), separate key for global help

### Animation & UX
- **Minimum display time** for progress states: `minProgressDisplayMs = 300`
- Prevents flicker on fast operations — UI shows state long enough to register
- Applied to: cleanup scan, create worktree phases, integration apply
- Does NOT add delay to operation — just ensures visibility

### Key Mapping
- Documented in Go const file, used uniformly
- `?` = contextual details/help for current view
- Separate key (TBD — `F1` or `h` outside text inputs) for global help

### Component Structure
- Flat files in `internal/tui/` with naming convention (e.g., `chrome.go`, `portal.go`)
- No sub-package — avoids export ceremony with Model
- Pure functions where possible for isolated testability

## Cross-Change Requirements

Every change MUST include:
- [ ] Full unit test coverage for new/modified functions
- [ ] E2E tests (teatest) for view flows
- [ ] Responsive behavior tests (different terminal sizes)
- [ ] Updates to `.impeccable.md` design system
- [ ] Updates to `sentei --help` if CLI flags change
- [ ] Updates to relevant openspec specs (delta specs)
- [ ] `go fmt`, `go vet`, `go test ./...` passing

## Files Likely Affected

### Change 1
| File | Action |
|------|--------|
| `internal/tui/chrome.go` | NEW — shared chrome helpers |
| `internal/tui/keys.go` | MODIFY — add key mapping consts, contextual `?` |
| `internal/tui/styles.go` | MODIFY — delete `styleHeader`, add animation consts |
| `internal/tui/progress.go` | MODIFY — use chrome, add windowing + bar |
| `internal/tui/create_progress.go` | MODIFY — use chrome, add bar |
| `internal/tui/repo_progress.go` | MODIFY — use chrome, add bar |
| `internal/tui/integration_progress.go` | MODIFY — use chrome helpers |
| `internal/tui/summary.go` | MODIFY — use chrome helpers, fix `"v"` → `●` |
| `internal/tui/create_summary.go` | MODIFY — use chrome helpers |
| `internal/tui/repo_summary.go` | MODIFY — use chrome helpers |
| `internal/tui/migrate_summary.go` | MODIFY — use chrome helpers |
| `internal/tui/cleanup_result.go` | MODIFY — use chrome helpers |
| `internal/tui/confirmation.go` | MODIFY — use chrome, drop `styleDialogBox` border |
| `.impeccable.md` | MODIFY — expand with component patterns |
| Tests for all above | NEW/MODIFY |

### Change 2
| File | Action |
|------|--------|
| `internal/tui/portal.go` | NEW — DetailPortal component |
| `internal/tui/help.go` | NEW — help overlay consumer |
| `go.mod` | MODIFY — add `bubbletea-overlay` dependency |

### Change 3
| File | Action |
|------|--------|
| `internal/tui/cleanup_preview.go` | NEW — replaces cleanup_confirm for TUI |
| `internal/tui/cleanup_confirm.go` | MODIFY — CLI-only path |
| `internal/tui/confirm.go` | MODIFY — dirty/unpushed gate logic |
| `internal/cleanup/` | MODIFY — dry-run scan API |
| `cmd/` | MODIFY — `--yes` flag |
