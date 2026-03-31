# Playground Progress Min-Duration Design

**Date:** 2026-03-31  
**Status:** Approved

## Problem

Two independent bugs affect `--playground` mode:

1. **Slow loading** — `DelayRunner` (800 ms/command) is passed as the single model runner. `loadWorktreeContext` calls `EnrichWorktrees`, which issues ~3 git commands per worktree. With 7 playground worktrees that is ~17 s of "loading…" before the menu becomes usable.

2. **Flashing progress screens** — Deletion uses `os.RemoveAll` (not the runner), so `DelayRunner` never throttled it. Create, migrate, and integration operations execute fast git commands and complete before the user can read the progress screen.

`DelayRunner` delivered no benefit while causing serious harm. It was the wrong tool for the job.

## Design

### Fix 1 — Remove `DelayRunner`

In `main.go`, stop wrapping `tuiRunner` with `DelayRunner`. Pass the real `runner` directly to `NewMenuModel`. Reads become instant.

Delete `DelayRunner` (`git.DelayRunner`) entirely — it has no remaining callers.

### Fix 2 — Minimum Progress View Duration

Add three fields to `Model`:

```go
minProgressDuration time.Duration // 0 in production; ~1.5s in playground
progressStartedAt   time.Time     // set when entering any progress view
progressToken       int           // incremented on each entry; guards stale timers
```

Add one message type:

```go
type progressHoldExpiredMsg struct{ token int }
```

Add one helper method:

```go
// holdOrAdvance transitions to targetView immediately if the minimum duration
// has already elapsed, or schedules a tea.Tick for the remaining time.
// The caller must have already stored all result state into the model before calling.
func (m Model) holdOrAdvance(targetView viewState) (Model, tea.Cmd)
```

`holdOrAdvance` implementation:
- If `m.minProgressDuration == 0` or `time.Since(m.progressStartedAt) >= m.minProgressDuration`, set `m.view = targetView` and return `nil`.
- Otherwise compute `remaining = m.minProgressDuration - time.Since(m.progressStartedAt)` and return `tea.Tick(remaining, func(time.Time) tea.Msg { return progressHoldExpiredMsg{m.progressToken} })`.

`progressHoldExpiredMsg` handler (in `Update`):
- If `msg.token != m.progressToken`, discard (stale timer from a previous flow).
- Otherwise set `m.view` to the stored target and return `nil`.

The target view must be stored in the model before the hold so the expired handler knows where to go. Add `progressTargetView viewState` to `Model`.

**Setting the timer on progress entry:** Each place that transitions into a progress view sets:
```go
m.progressStartedAt = time.Now()
m.progressToken++
```

This covers: `updateConfirm` (→ `progressView`), `updateCreateConfirm` / `updateCreateOptions` (→ `createProgressView`), `updateCloneConfirm` (→ `repoProgressView`), `updateMigrateConfirm` (→ `migrateProgressView`), `updateIntegrationList` (→ `integrationProgressView`), `updateCleanupConfirm` (→ `cleanupResultView`).

### Affected Progress Flows

| Flow | Terminal message | Target view |
|------|-----------------|-------------|
| Deletion | `cleanupCompleteMsg` (final step: prune → cleanup → summary) | `summaryView` |
| Create worktree | `createCompleteMsg` | `createSummaryView` |
| Repo ops (clone/migrate/create) | `repoDoneMsg` | `repoSummaryView` / `migrateNextView` |
| Integration | final integration msg | `integrationResultView` |
| Standalone cleanup | `cleanupResultView` transition when result populated | `cleanupResultView` (stays, shows populated) |

Each terminal handler: store result into model state, call `holdOrAdvance(targetView)`.

### Constructor Change

`NewMenuModel` gains an options pattern or a `WithMinProgressDuration(d)` option:

```go
func NewMenuModel(runner git.CommandRunner, shell git.ShellRunner, repoPath string,
    cfg *config.Config, context repo.RepoContext, opts ...ModelOption) Model
```

In `main.go`, playground mode passes `WithMinProgressDuration(1500 * time.Millisecond)`.

## Architecture

```
main.go
  playground → NewMenuModel(..., WithMinProgressDuration(1500ms))
  production → NewMenuModel(...)          // minProgressDuration = 0

Model
  runner                  // real runner, no delay
  minProgressDuration     // 0 or 1.5s
  progressStartedAt       // set on progress view entry
  progressToken           // stale-message guard
  progressTargetView      // where to go when hold expires

updateConfirm / updateXxxConfirm
  → set progressStartedAt, progressToken++
  → set view = progressView (etc.)

updateProgress (terminal message)
  → store result state into model
  → call holdOrAdvance(summaryView)
      if elapsed >= min  → m.view = summaryView, return nil
      if elapsed <  min  → return tea.Tick(remaining, progressHoldExpiredMsg{token})

Update (top-level)
  progressHoldExpiredMsg{token}
      if token != m.progressToken → discard
      else → m.view = m.progressTargetView
```

## Removed Code

- `git.DelayRunner` struct and its `Run` / `RunShell` methods — deleted
- `playgroundDelay` constant in `main.go` — deleted
- `tuiRunner` variable in `main.go` — deleted (use `runner` directly)

## Testing

**Existing tests:** `minProgressDuration` defaults to 0 → all completion handlers call `holdOrAdvance` → elapsed always ≥ 0 → immediate transition → no behaviour change. All existing tests pass unchanged.

**New unit tests for `holdOrAdvance`:**
- `minProgressDuration = 0` → transitions immediately, returns nil cmd
- `minProgressDuration > 0, elapsed < min` → returns a cmd (tea.Tick), does not transition yet
- `minProgressDuration > 0, elapsed >= min` → transitions immediately
- Stale token: `progressHoldExpiredMsg` with wrong token → discarded, view unchanged

**New playground E2E test:**
- Construct model with `minProgressDuration = 50ms`
- Trigger deletion flow through to `allDeletionsCompleteMsg`
- Assert still in `progressView` immediately after
- Use `teatest.WaitFor` to poll for `summaryView` — no `time.Sleep`

**`DelayRunner` tests:** Deleted alongside the struct.
