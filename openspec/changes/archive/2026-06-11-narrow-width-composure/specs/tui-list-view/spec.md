# tui-list-view Delta

## ADDED Requirements

### Requirement: One-line rows under any width
List rows SHALL always occupy exactly one terminal line: cells are truncated to their column widths with a trailing `…` and never wrap. Below 72 columns the Subject column SHALL be dropped; below 56 columns the Age column SHALL also be dropped; the remaining columns stay aligned and the status bar, legend, and key hints remain visible.

#### Scenario: Narrow terminal keeps structure
- **WHEN** the list renders at 60 columns with long branch names
- **THEN** every row SHALL be one line, branch names SHALL end with `…`, the Subject column SHALL be absent, and the status bar and legend SHALL be visible

#### Scenario: Single truncation idiom
- **WHEN** any list cell overflows its column
- **THEN** it SHALL truncate with the `…` glyph, never ASCII `...` and never a hard clip
