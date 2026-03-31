# Change 1 Init Prompt: ui-chrome-unification

> Copy everything below the line into your first message to Claude.

---

## Context

We're implementing the first of three sequential UI standardization changes for sentei (a TUI git worktree manager). This session was planned in a prior brainstorming session — all specs, design, and tasks are already written. Your job is to implement, not design.

### Background Files (read these first)

1. **Session meta** (full design decisions, cross-change requirements): `docs/handoff/session-meta.md`
2. **Original handoff** (screenshots, issues found): `docs/handoff/ui-alignment.md`
3. **Full brainstorming session** (118K transcript): `docs/handoff/session-c45565c0.txt`
4. **Design system**: `.impeccable.md`

### This Change

**Name:** `ui-chrome-unification`
**OpenSpec location:** `openspec/changes/ui-chrome-unification/`
**Artifacts:** proposal.md, design.md, specs/*/spec.md, tasks.md — all complete

**What it does:**
- Extracts shared chrome helpers (`viewTitle`, `viewSeparator`, `viewKeyHints`) as pure functions
- Creates `ProgressLayout` — shared rendering for all progress views with phases, steps, progress bar
- Adds adaptive windowing for large item lists (30+ worktrees) responsive to terminal height
- Adds stat line with indicator legend: `● N done · ◐ N active · · N pending  showing X of Y`
- Adds animation buffer constants (`MinProgressDisplay = 300ms`) to prevent flicker
- Consolidates key mappings (contextual `?`, global `F1`)
- Standardizes all views to use shared chrome (drops `styleHeader` bg badge, `styleDialogBox` border)
- Fixes removal summary `"v"` → `●`
- Expands `.impeccable.md` with Component Patterns section

### Key Design Decisions (from brainstorming)

- **Pure functions** for all rendering logic — data in, string out, no Model dependency. Enables table-driven unit testing.
- **Flat files** in `internal/tui/` with naming convention (`chrome.go`, `window.go`, `constants.go`) — no sub-package.
- **Adaptive windowing** via budget calculation: `availableLines = termHeight - fixedChrome`. Priority: failed (always) > active (always) > recent completed > next pending.
- **Animation buffer** via Cmd wrapping (`bufferTransition`), not `time.Sleep`. `MinProgressDisplay` is a `var` (not const) so tests can set to 0.
- **Confirmation views** drop border box, use standard chrome. `styleDialogBox` and `styleHeader` are deleted.

### How to Execute

```
/opsx:apply ui-chrome-unification
```

Follow TDD strictly — the tasks.md is ordered by dependency and each task specifies test-first. Use `gotestsum` for running tests.

### Cross-Change Requirements (applies to ALL tasks)

- Full unit test coverage for new/modified functions
- E2E tests (teatest) for view flows
- Responsive behavior tests (different terminal sizes)
- Updates to `.impeccable.md` design system
- `go fmt`, `go vet`, `go test ./...` passing
- Never use `time.Sleep` in tests — use `teatest.WaitFor` or condition-based polling
- Use impeccable skill for any TUI design questions
- If you encounter pre-existing errors, fix them

### What Comes After

This is Change 1 of 3. Changes 2 (`detail-portal-component`) and 3 (`cleanup-preview-redesign`) depend on this. Don't implement anything from those changes — stay scoped to `ui-chrome-unification`.
