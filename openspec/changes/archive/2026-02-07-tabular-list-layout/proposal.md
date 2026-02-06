## Why

The worktree list view uses hardcoded column widths (`30`/`16`/`40`) that don't adapt to terminal width. Long branch names overflow and misalign columns, while narrow terminals clip content without graceful degradation. The layout needs to be responsive and properly tabular.

## What Changes

- Replace manual `fmt.Sprintf("%-*s", ...)` column formatting with `lipgloss/table` (already bundled in lipgloss v1.1.0, zero new dependencies)
- Capture terminal width from `tea.WindowSizeMsg` and pass it to the table via `.Width()`
- Remove hardcoded `colWidthBranch`, `colWidthAge`, `colWidthSubject` constants
- All columns (cursor, checkbox, status, branch, age, subject) become table columns with per-cell styling via `StyleFunc(row, col)`
- Prefix columns (cursor, checkbox, status) use fixed widths; data columns (branch, age, subject) auto-size to fill available width
- Table renders borderless (no chrome, no separators) to preserve the current visual feel
- Smart shrinking in narrow terminals: `lipgloss/table` resizer uses median-based column shrinking, so the longest-variance column (branch) shrinks first

## Capabilities

### New Capabilities

_None — this is a rendering refactor, not new functionality._

### Modified Capabilities

- `tui-list-view`: Row layout changes from manual `fmt.Sprintf` + `lipgloss.JoinHorizontal` to `lipgloss/table` with responsive column widths. Existing row content (checkbox, status indicator, branch, age, subject) and behavior (selection, cursor, scrolling) are unchanged.

## Impact

- **Code**: `internal/tui/list.go` (view rendering), `internal/tui/styles.go` (column styles), `internal/tui/model.go` (store terminal width)
- **Dependencies**: None new — `lipgloss/table` is a subpackage of the existing `lipgloss v1.1.0` dependency
- **Tests**: `internal/tui/list_test.go` — test assertions that match on exact row strings will need updating to account for new column spacing
- **Risk**: Low — purely visual rendering change, no behavior or data model changes
