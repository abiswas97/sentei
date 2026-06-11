# tui-list-view Delta

## ADDED Requirements

### Requirement: Mouse wheel scrolls the worktree list
The list view SHALL treat mouse wheel events as cursor navigation aliases: wheel down moves the cursor down one row, wheel up moves it up one row, through the same paths as the `j`/`k` keys.

#### Scenario: Wheel down moves cursor down
- **WHEN** the user scrolls the wheel down with the cursor on row 2 of 5
- **THEN** the cursor SHALL move to row 3, identical to pressing `j`

#### Scenario: Wheel up at the top boundary
- **WHEN** the user scrolls the wheel up with the cursor on the first row
- **THEN** the cursor SHALL remain on the first row (no wrap, no error)
