# Spinner for Indeterminate Waits

## Why

Sentei's two indeterminate waits render frozen text: the cleanup scan shows a static `◐ Scanning repository…` and the menu's worktree load shows a static `loading…` hint, making a slow scan indistinguishable from a hang. Decided 2026-06-11: indeterminate waits get `bubbles/spinner`; determinate progress keeps the bar.

## What Changes

- A shared `bubbles/spinner` (MiniDot, accent-colored) animates exactly two states: the cleanup-preview scanning line and the menu's "Remove worktrees" loading hint.
- Spinner ticks run only while an indeterminate state is visible; all other views are unaffected.
- Menu items gain a `loading` flag (data, not a string match) that drives the spinner placement.
- Determinate progress (phases, steps, bar) is untouched; the static indicator vocabulary (`◐` = active) remains for step states.
- `.impeccable.md` Timing section is updated: `holdOrAdvance` stops being "the only timing mechanism"; spinner ticks are the second, scoped to indeterminate waits.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `tui-design-system`: gains an indeterminate-wait indicator requirement (animated spinner, scope, tick lifecycle).

## Impact

- **Code**: `internal/tui/model.go` (spinner state, tick routing, menu item flag), `internal/tui/menu.go` (loading hint render), `internal/tui/cleanup_preview.go` (scanning line). `charm.land/bubbles/v2/spinner` already in the module.
- **Tests**: tick routing and rendering tests; existing teatest assertions on "Scanning repository" text remain valid (the text stays, only the glyph animates).
