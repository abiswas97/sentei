# tui-design-system Delta

## ADDED Requirements

### Requirement: Confirm rows are columnar
The confirm-deletion screen SHALL list selected worktrees with the status badge in a fixed-width left gutter, names aligned in a column after it (pre-truncated to a stable width), and risk notes trailing only on at-risk rows. Clean rows carry no note text.

#### Scenario: Badges align in a gutter
- **WHEN** the confirm screen renders any mix of clean and at-risk selections
- **THEN** every badge SHALL start at the same column and every name SHALL start at the same column

#### Scenario: Long names do not shift the columns
- **WHEN** a selected worktree's label exceeds the name-column cap
- **THEN** the label SHALL truncate with an ellipsis and the columns SHALL hold
