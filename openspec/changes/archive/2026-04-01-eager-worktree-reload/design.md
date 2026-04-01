## Context

Bubble Tea uses the Elm architecture: `Update(msg) -> (Model, Cmd)`. Commands produce messages asynchronously; messages are delivered sequentially to `Update`. The app dispatches messages to view-specific handlers via a `switch m.view` in `Update()`.

Currently, after a mutation (removal, creation, integration apply), `stateStale` is set to `true`. When the user returns to the menu, the first `updateMenu` call hits a gate that consumes the entire update cycle to fire `loadWorktreeContext`, returning early before processing the actual message (typically a keypress). This swallows the user's first input.

Five mutation sites set `stateStale`: `progress.go` (removal), `create_progress.go`, `repo_progress.go`, `integration_progress.go`, `cleanup_result.go`. Only three of these flows return to the menu; the other two exit with `tea.Quit`.

## Goals / Non-Goals

**Goals:**
- Eliminate swallowed keypresses after mutations by decoupling reload from user input
- Ensure worktree data (count, list) is fresh by the time the user reaches the menu
- Handle out-of-order async responses safely (multiple `loadWorktreeContext` in flight)

**Non-Goals:**
- Splitting `loadWorktreeContext` into list-only vs full-enrichment phases (enrichment is fast enough and the head start from eager fire makes splitting unnecessary)
- Adding retry logic for failed reloads (pre-existing gap, not introduced by this change)
- Changing `loadWorktreeContext` itself (the function is fine; the problem is when it's called)

## Decisions

### 1. Fire reload at mutation completion, not at view transition

**Choice:** Call `loadWorktreeContext` via `tea.Batch` alongside `holdOrAdvance` at the three mutation completion sites that return to the menu.

**Alternative considered:** Fire reload at the `m.view = menuView` transition sites (summary -> menu). Simpler (no global handler needed), but the reload doesn't start until the user presses Enter on the summary screen. With 30+ worktrees and parallel enrichment, the head start matters.

**Alternative considered:** Reorder the `stateStale` check to run after keypress processing in `updateMenu`, using `tea.Batch` to combine both commands. Keeps the lazy approach but avoids swallowing input. Rejected because it still delays the reload and adds control flow complexity.

**Pattern at each mutation site:**
```go
m.worktreeGeneration++
updated, holdCmd := m.holdOrAdvance(targetView)
return updated, tea.Batch(holdCmd, loadWorktreeContext(m.runner, m.repoPath, m.worktreeGeneration))
```

### 2. Handle `worktreeContextMsg` globally in `Update()`

**Choice:** Process `worktreeContextMsg` before the `switch m.view` dispatch, same as the existing `progressHoldExpiredMsg` precedent.

**Why:** If the message is handled only in `updateMenu`, it gets dropped when arriving during summary/progress views (no matching case in those handlers). The global handler ensures the response is always applied.

**The handler applies the refresh unconditionally across views.** `reindex()` and `updateMenuHints()` are cheap, pure, in-memory operations. They write to `m.remove.worktrees`, `m.remove.visibleIndices`, `m.remove.cursor`, and `m.menuItems[2].hint`. No view other than menu/list reads these fields during rendering, so mutating them off-screen is harmless.

### 3. Generation token to guard against stale responses

**Choice:** Add a `worktreeGeneration uint64` field to the model. Increment it before each `loadWorktreeContext` call. Pass the current generation into the command; include it in `worktreeContextMsg`. The global handler only applies the response if `msg.generation == m.worktreeGeneration`.

**Why:** Multiple `loadWorktreeContext` calls can be in flight simultaneously (Init + post-mutation, or two rapid mutations). Without a token, an older response landing after a newer one would overwrite fresh data with stale data. This is the same pattern used by `progressHoldExpiredMsg` with `progressToken`.

### 4. Skip eager reload for `tea.Quit` flows

**Choice:** Remove `stateStale = true` from `repo_progress.go` and `cleanup_result.go` without replacing it. Do not fire `loadWorktreeContext` from these sites.

**Why:** These flows exit the process. Firing an async reload would spawn git subprocesses (up to 10 concurrent enrichment goroutines for 30+ worktrees) that get orphaned on process exit. Wasteful and pointless.

### 5. Delete `stateStale` entirely

**Choice:** Remove the `stateStale` field from the model, the gate in `updateMenu`, and the `worktreeContextMsg` case from `updateMenu` (now dead code after the global handler).

**Why:** The eager reload replaces the lazy mechanism completely. Keeping `stateStale` as a "fallback" adds a second code path that is never exercised, violating the no-dead-code principle. If a future mutation site forgets to fire the reload, the fix is to add the reload call there, not to maintain a shadow mechanism.

## Risks / Trade-offs

**[Enrichment cost during summary screen]** Eager reload spawns up to 10 concurrent git subprocesses per worktree while the user is reading the summary. On a machine under load, this could cause a brief CPU spike. -> Acceptable: the same work happens today, just slightly later. The concurrency cap (10) bounds the impact.

**[No retry on reload failure]** If `loadWorktreeContext` fails (e.g., git process locked), the error is silently dropped and the menu shows stale data. -> Pre-existing gap. The current `stateStale` approach also drops errors (the flag is cleared before the async call). Out of scope for this change.

**[Generation token overflow]** `uint64` overflow after 2^64 increments. -> Not a real concern for a TUI app.
