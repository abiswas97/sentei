## Context

wt-sweep can discover and enrich worktrees but has no interactive interface. The current `main.go` prints a flat list to stdout. We need a Bubble Tea TUI that lets users browse, select, and delete worktrees. The data layer (`internal/git`, `internal/worktree`) is stable and provides `[]git.Worktree` with all needed metadata.

Bubble Tea uses the Elm architecture: Model holds state, Update handles messages, View renders. All UI is built by composing these three methods.

## Goals / Non-Goals

**Goals:**
- Interactive list with navigation, multi-select, and status indicators
- Confirmation dialog with safety warnings for dirty/untracked worktrees
- Parallel deletion with live progress feedback
- Post-deletion summary
- Clean separation between TUI components and business logic

**Non-Goals:**
- Sorting/filtering UI (post-MVP, PRD F7)
- Dry-run mode (post-MVP, PRD F8)
- Non-interactive/CI mode (post-MVP, PRD F9)
- Configuration file support (post-MVP, PRD F10)
- Branch protection rules (post-MVP, PRD F11)

## Decisions

### D1: Single Bubble Tea model with view states (not nested models)

Use one top-level `Model` with a `viewState` enum (`listView`, `confirmView`, `progressView`, `summaryView`). The `Update` and `View` methods switch on this state.

**Why over nested/composed models**: The app has a linear flow (list → confirm → progress → summary) with shared state (the worktree slice, selections). A single model avoids the complexity of inter-model messaging. The views are simple enough that splitting `View()` into helper functions per state keeps things readable.

### D2: Worktree list uses custom rendering (not bubbles/list)

Render the list manually with Lip Gloss rather than using the `bubbles/list` component.

**Why**: `bubbles/list` is designed for filterable item lists with a delegate pattern. Our list needs custom columns (branch, age, commit, status), multi-select checkboxes, and ASCII status indicators. A custom render with `lipgloss.JoinHorizontal` gives full control over column layout and alignment. The navigation logic (cursor, viewport scrolling) is trivial to implement.

### D3: Deletion logic lives in `internal/worktree/deleter.go`

The `Deleter` follows the same pattern as `EnrichWorktrees`: takes a `CommandRunner`, a slice of worktrees, max concurrency, and returns results. It sends progress updates via a channel that the TUI consumes as Bubble Tea messages.

**Why**: Keeps business logic testable without TUI. The channel-based progress pattern is idiomatic for Bubble Tea — the model spawns a command that returns `tea.Msg` values as deletions complete.

### D4: Progress via tea.Cmd channel pattern

Deletion starts a `tea.Cmd` that launches goroutines and sends `worktreeDeletedMsg` or `worktreeDeleteFailedMsg` back through the Bubble Tea runtime. A final `allDeletionsCompleteMsg` signals the transition to summary view.

**Why over polling/ticking**: Bubble Tea's message-passing is the canonical way to handle async work. Each deletion result arrives as a message, naturally updating the progress bar.

### D5: Filter bare entry from display

The bare repository entry (IsBare=true) is excluded from the displayed list. It's not a real worktree and users should never select it for deletion.

### D6: Relative time display for last activity

Display last commit date as relative time ("3 days ago", "2 months ago") using a simple helper function. No external dependency needed — `time.Since` and a few threshold checks.

### D7: File organization

```
internal/tui/
├── model.go      # Model struct, Init, Update, View, view state enum
├── list.go       # List rendering helpers (row rendering, column layout)
├── confirm.go    # Confirmation view rendering and update logic
├── progress.go   # Progress view rendering, deletion message types
├── summary.go    # Post-deletion summary rendering
├── styles.go     # Lip Gloss style definitions
└── keys.go       # Key binding definitions (keymap struct)

internal/worktree/
└── deleter.go    # DeleteWorktrees function, progress channel pattern
```

## Risks / Trade-offs

- **[Terminal compatibility]** → ASCII status indicators (`[ok]`, `[~]`, `[!]`, `[L]`) used throughout for maximum compatibility. No emoji dependency.
- **[Large worktree counts]** → With 100+ worktrees, the list could be slow to render. Mitigation: viewport-based rendering (only render visible rows). For MVP, render all — optimize if needed.
- **[Deletion race conditions]** → A worktree could be modified between selection and deletion. Mitigation: `git worktree remove --force` handles most cases; failures are reported in the summary.
