## Context

The TUI progress view (`internal/tui/progress.go`) is wired to `worktree.DeleteWorktrees()` which sends `DeletionEvent`s over a channel as each worktree starts/completes/fails deletion. The current `startDeletion()` function returns a single `tea.Cmd` closure that reads all events from the channel internally, discards them, waits for completion, and returns one `allDeletionsCompleteMsg`. The Bubble Tea Update loop never sees intermediate events, so the progress bar and per-worktree indicators stay static until everything finishes.

In real-world monorepos (3GB+ worktrees, 25+ branches), individual `git worktree remove` operations take seconds to over a minute. The progress UI is essential — not cosmetic.

The test playground creates tiny worktrees that delete instantly, making progress UI unobservable during development.

## Goals / Non-Goals

**Goals:**
- Deletion channel events reach the Bubble Tea Update loop individually, triggering re-renders after each event
- Progress bar percentage and per-worktree status lines update incrementally in real-time
- Playground mode provides visibly slow deletions for testing the progress UI

**Non-Goals:**
- Changing the `DeleteWorktrees()` API or the `DeletionEvent` channel contract
- Adding spinner animations or tick-based updates (status text transitions are sufficient)
- Making the delay configurable via CLI flags (hardcoded for playground only)

## Decisions

### D1: Bubble Tea Cmd chaining for streaming channel events

**Decision**: Use a `waitForDeletionEvent` Cmd pattern — a function that reads one event from the channel and returns it as a `tea.Msg`. On receipt, `updateProgress` handles the event and returns another `waitForDeletionEvent` Cmd to read the next. When the channel closes (deletion goroutine finished), return a sentinel `allDeletionsCompleteMsg`.

**Alternatives considered**:
- *`tea.Program.Send()` from goroutine*: Requires passing `*tea.Program` into the model or deleter. Breaks Elm architecture encapsulation and makes testing harder.
- *Tick-based polling*: Use `tea.Tick` to periodically check a shared state. Adds unnecessary complexity, latency, and race conditions around shared mutable state.

**Why Cmd chaining**: It's the idiomatic Bubble Tea pattern for streaming external events. Each Cmd is a blocking read on the channel, returns exactly one Msg, and the framework handles the scheduling. No shared mutable state, no references to the program, fully testable.

### D2: Store progress channel on Model

**Decision**: Add a `progressCh <-chan worktree.DeletionEvent` field to `Model`. `startDeletion()` creates the channel, spawns the deletion goroutine, stores the channel on the model, and returns the first `waitForDeletionEvent` Cmd.

**Rationale**: The channel must survive across multiple Update calls. Storing it on the model is the standard Bubble Tea approach for long-lived async resources (similar to how file watchers or websocket connections are handled).

### D3: DelayRunner wrapper for playground

**Decision**: Add a `DelayRunner` struct implementing `CommandRunner` that wraps another runner and sleeps for a configurable duration before delegating each `Run()` call. In `main.go`, when `--playground` is set, wrap the runner with `DelayRunner` only for the TUI model (enrichment uses the unwrapped runner).

**Alternatives considered**:
- *Add delay parameter to `DeleteWorktrees()`*: Mixes testing concerns into production code.
- *Sleep inside playground's git operations*: Playground doesn't control deletion — the deleter does.
- *Create more/larger worktrees in playground*: Unreliable timing, wastes disk space, doesn't guarantee visible delay.

**Why DelayRunner**: Clean decorator pattern. The deleter, TUI, and playground code remain unaware of the delay. It's injected at the composition root (`main.go`) and only in playground mode. Since enrichment completes before TUI launch, the delay only affects deletion commands — exactly what we want.

### D4: Delay duration of 800ms per operation

**Decision**: Use 800ms delay in playground mode. With 6 worktrees and concurrency of 5, this produces ~2 rounds of visible progress over ~1.5 seconds total — enough to see the bar move and statuses transition without feeling sluggish.

## Risks / Trade-offs

**[Channel read after model copy]** Bubble Tea copies the model on each Update. The channel field is a reference type (`<-chan`), so copies share the same underlying channel. No risk of lost events. → No mitigation needed.

**[Playground delay feels artificial]** 800ms is a fixed value that may feel too fast or slow on different machines. → Acceptable for a dev/demo tool. Real-world repos provide natural delay. Can be made configurable later if needed.

**[Message ordering]** With concurrent deletions, `DeletionStarted` for worktree B may arrive before `DeletionCompleted` for worktree A. The UI handles this correctly since each message targets a specific path and updates only that entry's status. → No risk.
