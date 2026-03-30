## MODIFIED Requirements

### Requirement: Display confirmation dialog with selected worktrees
The TUI SHALL show a confirmation dialog listing all selected worktrees before deletion, with clear identification of each worktree by branch name and its clean/dirty status. The confirmation dialog SHALL use the shared confirmation view component, displaying worktree-specific details (branch names, status indicators) as key-value content and the equivalent CLI command at the bottom.

#### Scenario: Confirmation with clean worktrees only
- **WHEN** the user enters the confirmation view with 3 selected worktrees, all clean
- **THEN** the dialog SHALL list all 3 worktrees with their branch names and "(clean)" label
- **AND** the equivalent CLI command SHALL be displayed (e.g., `sentei remove --branches feature-a,feature-b,feature-c`)

#### Scenario: Confirmation with dirty worktrees
- **WHEN** at least one selected worktree has HasUncommittedChanges=true
- **THEN** the dialog SHALL display a warning icon next to that worktree with "HAS UNCOMMITTED CHANGES" and a summary warning at the bottom stating the count of worktrees with uncommitted changes

#### Scenario: Confirmation with untracked files
- **WHEN** at least one selected worktree has HasUntrackedFiles=true
- **THEN** the dialog SHALL display a warning indicating untracked files present

## ADDED Requirements

### Requirement: Confirmation view reachable via filter flags
The worktree deletion confirmation view SHALL be reachable when filter flags (`--stale`, `--merged`, `--all`) resolve a non-empty selection and the user is in TUI mode.

#### Scenario: Filter flags resolve selection then show list
- **WHEN** the user runs `sentei remove --merged` and 3 worktrees have merged branches
- **THEN** the TUI SHALL show the list view with those 3 worktrees pre-selected, and the user can proceed to confirmation via Enter
