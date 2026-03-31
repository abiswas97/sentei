## Context

sentei has 12+ TUI views built incrementally across 6 changes. Each view hand-rolls its own title, separator, and key hint rendering, leading to visual inconsistencies. The integration progress view is the closest to the target aesthetic but lacks features other views have (phase headers, pending phase display). The removal progress view is the biggest outlier — it uses a white-on-purple background badge (`styleHeader`) that nothing else uses and has no progress bar. Summary views use different success markers. The confirmation dialog wraps in a bordered box (`styleDialogBox`) unlike any other view.

Two downstream changes depend on this work: a detail portal component (Change 2) and a cleanup preview redesign (Change 3). Both need standardized chrome to build on.

## Goals / Non-Goals

**Goals:**
- Every view uses identical chrome (title format, separators, key hints)
- All progress views share a single rendering path with phase layout + progress bar
- Large item lists (30+ worktrees during parallel removal) render responsively based on terminal height
- Design vocabulary documented in `.impeccable.md` as an enforced contract
- Pure functions for all rendering logic — testable without Model or teatest

**Non-Goals:**
- Overlay/portal component (Change 2)
- Cleanup flow changes (Change 3)
- New features or behavioral changes — this is purely visual standardization
- Changing the Bubble Tea Cmd/Msg event flow — only the View() rendering changes

## Decisions

### D1: Pure functions over component Models

**Decision**: Chrome helpers (`viewTitle`, `viewSeparator`, `viewKeyHints`), windowing (`WindowSteps`), stat line (`viewStatLine`), and progress layout (`ProgressLayout.View()`) are all pure functions — data in, string out, no Model dependency.

**Alternatives considered**:
- *Bubble Tea sub-model per component*: Would require Init/Update/View lifecycle, message routing, and Model composition. Overkill for stateless rendering.
- *Method on Model*: Would couple rendering to Model, making isolated unit testing impossible.

**Why pure functions**: They're testable with table-driven tests across inputs (terminal sizes, item counts, states) without mock Models or teatest infrastructure. Each existing view's `viewXxx()` method becomes a thin adapter: build the data struct from domain state, call the pure function.

### D2: Flat files with naming convention, no sub-package

**Decision**: New files (`chrome.go`, `window.go`, `constants.go`) live flat in `internal/tui/` alongside existing files.

**Alternatives considered**:
- *`internal/tui/components/` sub-package*: Would require exporting types (`ProgressLayout`, `WindowResult`) that are only consumed within `tui`. Creates import ceremony without benefit since the Model struct is the center of the Elm architecture.

**Why flat**: Go sub-packages enforce export boundaries. Since all consumers are within `internal/tui/`, a sub-package would force us to export types purely for cross-package access, not because they're a real API boundary. The naming convention (`chrome.go`, `window.go`) provides sufficient organization.

### D3: Adaptive windowing via budget calculation

**Decision**: Windowing is computed per-render by subtracting fixed chrome lines from terminal height, then applying a priority-based selection: failed (always) > active (always) > recent completed > next pending.

**Alternatives considered**:
- *Fixed threshold (e.g., always window at >8 items)*: Ignores terminal size. A tall terminal can show 30 items fine.
- *Scrollable viewport with bubbles/viewport*: Adds statefulness (scroll position), key handling complexity, and is heavier than needed for a progress view where items are transient.

**Why budget calculation**: Truly responsive — adapts to any terminal height. Pure function — `WindowSteps(steps, availableLines)` is trivially testable. No scroll state to manage. The stat line `● N done · ◐ N active · · N pending  showing X of Y` makes the windowing transparent to the user.

### D4: Animation buffer via Cmd wrapping

**Decision**: A `bufferTransition(started time.Time, cmd tea.Cmd) tea.Cmd` helper ensures at least `MinProgressDisplay` (300ms) elapses before delivering completion messages. If the operation takes longer than 300ms, no delay is added.

**Alternatives considered**:
- *`time.Sleep` in the operation itself*: Blocks the goroutine, not the UI. But it conflates operation timing with display timing.
- *Tick-based approach*: Start a 300ms tick alongside the operation, only transition when both complete. More complex message handling.

**Why Cmd wrapping**: Stays within Bubble Tea's Cmd model. The wrapper Cmd runs the inner Cmd, checks elapsed time, sleeps the remainder if needed, then returns the message. Simple, composable, no new message types needed. Each phase records `time.Now()` at start; the completion handler wraps the transition Cmd with `bufferTransition`.

### D5: Drop `styleDialogBox` border from confirmations

**Decision**: Confirmation views use the standard view chrome (`viewTitle` + `viewSeparator` + `viewKeyHints`) instead of wrapping in `styleDialogBox`. Both `styleDialogBox` and `styleHeader` are deleted from `styles.go`.

**Alternatives considered**:
- *Keep border for confirmations as visual distinction*: The border creates a modal appearance without a background, looking like a floating box over nothing. Inconsistent with every other view.
- *Use `bubbletea-overlay` for true modal*: Overkill for confirmations which are a step in a flow, not an interruption. Reserved for the portal component (Change 2).

**Why standard chrome**: Visual consistency across all views. Confirmation views are not modals — they're full-screen steps in a flow. The title (`sentei ─ Confirm Cleanup`) and key hints (`enter confirm · esc back`) already communicate the interaction model clearly.

## Risks / Trade-offs

**[Risk] Existing test assertions break** → Expected and acceptable. All tests that assert on rendered output will need updating. This is part of the change scope, not a side effect. Tests are updated alongside the views they test.

**[Risk] `bufferTransition` makes tests slower** → Mitigated by making `MinProgressDisplay` a variable (not const) that tests can set to 0. Production code uses the default; tests override for speed.

**[Trade-off] Pure functions can't access Model state directly** → Accepted. Each view method builds a `ProgressLayout` or similar struct from Model state, then calls the pure function. This is a thin mapping layer — typically 10-15 lines per view. The testing benefit outweighs the minor adapter boilerplate.

**[Trade-off] Flat file organization may get crowded** → Acceptable at current scale (~20 files in `internal/tui/`). Adding 3 files (`chrome.go`, `window.go`, `constants.go`) is manageable. If `internal/tui/` grows beyond ~30 files, revisit sub-package extraction.
