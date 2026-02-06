## Context

The list view in `internal/tui/list.go` renders worktree rows using hardcoded column widths (`colWidthBranch=30`, `colWidthAge=16`, `colWidthSubject=40`) with `fmt.Sprintf("%-*s", ...)` and `lipgloss.JoinHorizontal`. Terminal width is tracked for row count (viewport height) but not used for column sizing.

`lipgloss/table` (subpackage of the existing `lipgloss v1.1.0` dependency) provides responsive column layout with smart resizing — it uses median-based shrinking so high-variance columns (branch names) degrade gracefully in narrow terminals.

## Goals / Non-Goals

**Goals:**
- Columns align cleanly across all rows regardless of content length
- Layout adapts to terminal width — fills wide terminals, degrades gracefully in narrow ones
- Preserve existing visual appearance (no borders, same colors, same row content)

**Non-Goals:**
- Adding column headers or table chrome
- Changing row content, selection behavior, or keyboard navigation
- Supporting column reordering or user-configurable column visibility

## Decisions

### Use `lipgloss/table` for column layout

**Choice**: Replace `fmt.Sprintf` + `JoinHorizontal` with a borderless `lipgloss/table`.

**Why**: The `lipgloss/table` resizer handles the hard part — proportional column sizing with median-based shrinking. Writing equivalent logic manually would duplicate ~200 lines of sizing code already available.

**Why not manual dynamic widths**: We explored computing widths manually (`remaining * 40%` etc.), but the table's resizer is more sophisticated — it considers actual content width distribution per column, not just fixed ratios.

**Alternatives considered**:
- Manual dynamic widths — simpler but less intelligent shrinking behavior
- Bubbles list component — too opinionated for our custom row format

### All row elements become table columns

**Choice**: Map every visual element to a table column:

| Col | Content    | Width Strategy     |
|-----|------------|--------------------|
| 0   | Cursor     | Fixed via StyleFunc (3 chars incl. gap) |
| 1   | Checkbox   | Fixed via StyleFunc (5 chars incl. gap) |
| 2   | Status     | Fixed via StyleFunc (6 chars incl. gap) |
| 3   | Branch     | Proportional (~50% of remaining, wraps) |
| 4   | Age        | Fixed via StyleFunc (~16 chars) |
| 5   | Subject    | Proportional (~50% of remaining, wraps) |

**Why**: Keeping cursor/checkbox/status outside the table and prepending per-line was considered, but making them columns lets the table manage all horizontal spacing uniformly.

### Borderless table rendering

**Choice**: Disable all borders:
```go
table.New().
    BorderTop(false).BorderBottom(false).
    BorderLeft(false).BorderRight(false).
    BorderColumn(false).BorderHeader(false).BorderRow(false)
```

**Why**: The current UI has no table chrome. With `borderColumn=false`, the resizer allocates zero border budget — all width goes to content.

### Per-row styling via StyleFunc closure

**Choice**: The `StyleFunc(row, col)` closure captures `m.cursor`, `m.selected`, and `m.worktrees` to apply:
- Cursor row: bold white (`styleCursorRow`)
- Selected row: pink (`styleSelectedRow`)
- Normal row: light gray (`styleNormalRow`)
- Status column: color varies per worktree state (clean/dirty/untracked/locked)

This mirrors the Pokemon example's pattern of looking up `data[row]` to determine cell styles.

### Store terminal width in Model

**Choice**: Add `width int` to `Model`, update from `tea.WindowSizeMsg`, pass to table via `.Width(m.width)`.

### Table is rebuilt each render

**Choice**: Construct the `table.Table` in `viewList()` on every render call rather than caching it on the model.

**Why**: The table is cheap to construct (just string slices + a style function). Caching would require invalidation on selection change, cursor move, window resize, and data change — all of which happen frequently. Bubble Tea's `View()` is expected to be a pure function of model state.

## Risks / Trade-offs

**[Risk] Borderless table rendering untested in our codebase** → The ANSI example from lipgloss uses default borders. We should verify borderless rendering produces clean output before completing implementation. A quick manual test with the playground will confirm.

**[Risk] StyleFunc per-cell overhead** → The closure is called once per cell per render. With 6 columns × 50 rows = 300 calls, this is negligible.

**[Trade-off] Status indicator ANSI codes in table cells** → `statusIndicator()` returns pre-styled strings (via `lipgloss.Style.Render`). The table's resizer uses `lipgloss.Width()` which correctly accounts for ANSI escape sequences, so column width calculations remain accurate.

**[Decision] Wrapping enabled** → `.Wrap(true)` allows long branch names to wrap within their cell rather than truncate. Rows expand vertically when content wraps — age and subject stay top-aligned on the first line. Viewport logic remains row-index based; if wrapped rows exceed the terminal height, the terminal clips naturally.

**[Decision] Consistent column padding** → Data columns (branch, age, subject) get `Padding(0, 1)` via StyleFunc for a 1-char right gap. Prefix columns (cursor, checkbox, status) bake the inter-column gap into their fixed Width instead, avoiding wrapping issues on small fixed-width cells. Total padding budget: 3 chars (3 data columns × 1 char).

**[Decision] Proportional column widths** → Branch and subject columns split remaining width (after fixed columns and padding) 50/50. This adapts to terminal width — wider terminals give both columns more room, narrower terminals squeeze proportionally.
