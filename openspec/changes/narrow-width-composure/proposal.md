# Narrow-Width Composure

## Why

The ten-agent visual audit found sentei structurally broken below ~80 columns (P0): table cells wrap mid-word (`Wrap(true)`, untruncated branch cells), doubling row heights, destroying row identity, and pushing the status bar, legend, and key hints off-screen; the detail portal's border clips at the right edge; subject cells truncate with ASCII `...` while everything else uses `…`; the create view hard-clips its header path.

## What Changes

- One-line rows become law: `Wrap(false)`, every cell pre-truncated to its column width with `…` via `truncateWithEllipsis`.
- Columns carry priority: below 72 columns the Subject column is dropped, below 56 Age is dropped too; structure survives, detail degrades.
- The detail portal clamps so its full border always fits the terminal (regression test at 60 columns).
- The create view header routes through `truncateWithEllipsis` like the menu's.
- The `...` literal is eliminated.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `tui-list-view`: gains the one-line-row and column-priority requirements.

## Impact
internal/tui/list.go (table construction), portal sizing check, create_branch.go header; tests for cell truncation, column dropping, portal fit.
