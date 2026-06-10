## Why

sentei's TUI views were built incrementally across multiple changes, leading to visual inconsistencies: the removal progress uses a white-on-purple background badge while every other view uses bold white text titles; create/repo progress views lack the progress bar that integration progress has; summary views use different success markers (`"v"` vs `●`); and the confirmation dialog wraps in a bordered box that nothing else uses. With 12+ views, this divergence undermines the polished, Charm-style experience sentei targets. Standardizing now — before adding more views (portal, cleanup preview) — prevents compounding the debt.

## What Changes

- Extract shared chrome helpers (`viewTitle`, `viewSeparator`, `viewKeyHints`) as pure functions used by every view
- Extract `ProgressLayout` — a shared rendering struct for all progress views, built on the existing `pipeline.Event`/`phaseDisplay` vocabulary, with phase-based layout, step indicators, and an overall progress bar
- Style the overall progress bar (accent fill, dim track) — today the integration view's bar renders with no color at all
- Add adaptive windowing for large item lists (30+ worktrees) that responds to terminal height, with a stat line showing indicator legend and counts
- Add an ellipsis-truncation helper used wherever paths and error text can overflow (today they hard-clip at the terminal edge)
- Enforce one indicator vocabulary: `●` done, `◐` active, `·` pending, `✗` failed — `●` is never reused for pending steps
- Consolidate key mappings into a single const file with contextual `?` (details) and `F1` (global help)
- Standardize confirmation views to use shared chrome (drop `styleDialogBox` border), with `·` hint separators everywhere (cleanup confirm currently uses `•`)
- Give the remove list the standard framing (repo subtitle + rules) the menu and integrations views already have
- Fix removal summary markers (`●`), the empty `Cleanup:` section header, and the cleanup view titling itself "Cleanup Complete" while still running
- Delete dead styles: `styleHeader`, `styleDialogBox`
- Expand `.impeccable.md` with Component Patterns section documenting the design system

Progress hold timing (`minProgressDuration`, `holdOrAdvance`) already landed on main and is out of scope here.

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
