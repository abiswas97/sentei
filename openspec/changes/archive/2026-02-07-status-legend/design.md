## Context

The list view footer currently renders a single status bar line with selection count and keybindings. Status indicators (`[ok]`, `[~]`, `[!]`, `[L]`) appear in each row but are unexplained. The rendering lives in `viewStatusBar()` in `internal/tui/list.go`.

## Goals / Non-Goals

**Goals:**
- Show a colored legend line explaining all four status indicators
- Keep it always visible (no toggle)
- Minimal vertical space cost (one line)

**Non-Goals:**
- Interactive legend (toggle with `?` key)
- Contextual legend (only showing indicators present in the current list)
- Legend in other views (confirm, progress, summary)

## Decisions

### 1. Legend as a separate line below the status bar

Append a new line after the existing keybindings line rather than merging them into one line. This keeps both lines readable and avoids cramming too much into a single line.

Alternative considered: single combined line — rejected because at narrow terminal widths the combined content would be too long.

### 2. Render legend using the same indicator styles

Reuse `styleStatusClean`, `styleStatusDirty`, `styleStatusUntracked`, `styleStatusLocked` for the legend indicators so the colors match the table exactly. Labels use `styleDim` (or `styleStatusBar` foreground) to stay visually subordinate.

### 3. Adjust viewport height for the extra line

The current height calculation is `m.height = max(msg.Height-4, 5)`. The `-4` accounts for: header (1) + blank line (1) + status bar padding-top (1) + status bar text (1). Adding a legend line means changing this to `-5` so the table doesn't overflow.

### 4. Build legend string in a `viewLegend()` helper

Separate function keeps `viewStatusBar()` unchanged and makes testing straightforward.

## Risks / Trade-offs

- [One less visible row] → Acceptable for the discoverability benefit; the legend is compact at ~45 chars so it doesn't dominate the footer.
- [Hardcoded indicator list] → If new indicators are added later, the legend must be updated manually. Acceptable given the low churn rate of status types.
