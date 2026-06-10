## ADDED Requirements

### Requirement: Create flow starts pristine on every menu entry
The TUI SHALL reset all create-flow input state when the user enters the create flow from the menu: the branch name input SHALL be empty (showing its placeholder), the base branch input SHALL hold the repo default branch, and staged options SHALL be at their defaults.

#### Scenario: Branch input cleared after a completed create
- **WHEN** the user creates a worktree named `feature/one`, returns to the menu, and enters the create flow again
- **THEN** the branch name input SHALL be empty with the placeholder visible, and typing `feature/two` SHALL result in exactly `feature/two`

#### Scenario: Branch input cleared after an abandoned create
- **WHEN** the user enters the create flow, types a partial branch name, presses esc back to the menu, and enters the create flow again
- **THEN** the branch name input SHALL be empty with the placeholder visible

### Requirement: Removal run state is created fresh per run
The TUI SHALL hold all per-run deletion state (outcomes, per-worktree statuses, total, progress channel, teardown results, prune error, cleanup result) in a dedicated run structure that is created anew when a deletion run starts, so no value from a previous run can influence the current one.

#### Scenario: Second removal run in one session shows correct progress
- **WHEN** the user deletes 2 worktrees, returns to the list, selects 1 more worktree, and confirms deletion
- **THEN** the progress view SHALL start at 0%, count only the current run's outcomes, and reach exactly 100% when the single deletion finishes

#### Scenario: Second removal run completes and reaches summary
- **WHEN** a second deletion run finishes all its selected worktrees
- **THEN** the TUI SHALL transition to the summary view exactly as it does for a first run

#### Scenario: Pending phases show pending on a second run
- **WHEN** a second deletion run is still removing worktrees
- **THEN** the prune/cleanup phase SHALL render as pending, not as completed from the previous run

### Requirement: Removal selection is cleared after a completed run
The TUI SHALL clear the selection map after a deletion run completes, so the list view's selection count reflects only selectable, currently listed worktrees.

#### Scenario: Selection count after returning to the list
- **WHEN** the user deletes 2 selected worktrees and returns to the removal list
- **THEN** the footer SHALL show `0 selected` and no row SHALL render a checked checkbox
