## Why

The TUI displays status indicators (`[ok]`, `[~]`, `[!]`, `[L]`) next to each worktree but provides no explanation of what they mean. New users must guess or consult documentation outside the tool. Adding an inline legend removes this friction.

## What Changes

- Add a status indicator legend line to the list view footer, displayed below the existing status bar
- The legend shows all four indicators with their colored styling and a short label: `[ok] clean  [~] dirty  [!] untracked  [L] locked`
- Always visible (no toggle required) — costs one line of vertical space

## Capabilities

### New Capabilities

_(none — this is a small addition to an existing view)_

### Modified Capabilities

- `tui-list-view`: The status bar requirement expands to include a legend line showing all status indicator meanings

## Impact

- `internal/tui/list.go`: Add legend rendering in `viewStatusBar()` or as a new helper called from `viewList()`
- `internal/tui/styles.go`: May need a dedicated style for the legend line (or reuse `styleStatusBar`)
- Vertical space: the list viewport height calculation (`m.height = max(msg.Height-4, 5)`) may need adjustment to account for the extra line
