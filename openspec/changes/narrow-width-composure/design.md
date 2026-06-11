# Narrow-Width Composure: Design

## Decisions

### D1: Pre-truncate, never wrap
The table renders with `Wrap(false)`; branch and subject cells are truncated to their computed widths with the shared `truncateWithEllipsis` before insertion, so the table never needs to cut. One truncation idiom app-wide.

### D2: Column priority as data
`showSubject = width >= 72`, `showAge = width >= 56` (0 width counts as wide for tests). Headers and rows are built dynamically; the style function receives the live width set. Branch absorbs freed space.

### D3: Portal fit is a tested invariant
A regression test renders the portal at 60x18 over a 60-wide background and asserts no output line exceeds the terminal width; the clamp is fixed wherever the test points.

## Risks
- [lipgloss table internals still pad dropped widths] → dynamic column construction avoids zero-width columns entirely.
