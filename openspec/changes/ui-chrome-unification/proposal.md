## Why

sentei's TUI views were built incrementally across multiple changes, leading to visual inconsistencies: the removal progress uses a white-on-purple background badge while every other view uses bold white text titles; create/repo progress views lack the progress bar that integration progress has; summary views use different success markers (`"v"` vs `●`); and the confirmation dialog wraps in a bordered box that nothing else uses. With 12+ views, this divergence undermines the polished, Charm-style experience sentei targets. Standardizing now — before adding more views (portal, cleanup preview) — prevents compounding the debt.

## What Changes

- Extract shared chrome helpers (`viewTitle`, `viewSeparator`, `viewKeyHints`) as pure functions used by every view
- Extract `ProgressLayout` — a shared rendering struct for all progress views with phase-based layout, step indicators, and an overall progress bar
- Add adaptive windowing for large item lists (30+ worktrees) that responds to terminal height, with a stat line showing indicator legend and counts
- Add animation buffer constants (`MinProgressDisplay = 300ms`) to prevent flicker on fast operations
- Consolidate key mappings into a single const file with contextual `?` (details) and `F1` (global help)
- Standardize confirmation views to use shared chrome (drop `styleDialogBox` border)
- Fix removal summary `"v"` success marker → `●`
- Delete dead styles: `styleHeader`, `styleDialogBox`
- Expand `.impeccable.md` with Component Patterns section documenting the design system

## Capabilities

### New Capabilities
- `tui-chrome`: Shared chrome helpers (title, separator, key hints) and progress layout component with adaptive windowing — the reusable rendering foundation for all TUI views
- `tui-design-system`: Documented design language in `.impeccable.md` covering component patterns, indicator vocabulary, timing constants, and key mapping

### Modified Capabilities
- `tui-progress`: Progress views use unified `ProgressLayout` with progress bar, adaptive windowing, and stat line instead of bespoke rendering per view
- `confirmation-view`: Confirmation dialog drops bordered box style, uses standard view chrome

## Impact

- `internal/tui/` — new files: `chrome.go`, `window.go`, `constants.go`; modified: all progress views, all summary views, `confirmation.go`, `styles.go`, `keys.go`
- `.impeccable.md` — expanded with Component Patterns section
- No API changes, no new dependencies, no CLI flag changes
- All existing E2E and unit tests will need updating to match new rendering output
