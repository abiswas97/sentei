## ADDED Requirements

### Requirement: Dry-run scan with loading state
The system SHALL run a dry-run cleanup scan when the user selects cleanup from the TUI menu, displaying a loading indicator while scanning.

#### Scenario: Loading state during scan
- **WHEN** the user selects cleanup from the TUI menu
- **THEN** the view SHALL display `viewTitle("Cleanup Preview")`, a separator, and `◐ Scanning repository…` using the active indicator

#### Scenario: Scan completes and shows preview
- **WHEN** the dry-run scan completes
- **THEN** the loading state SHALL transition to the preview with at least `MinProgressDisplay` elapsed (animation buffer)

#### Scenario: Scan with nothing to clean
- **WHEN** the dry-run scan finds nothing for safe or aggressive mode
- **THEN** the view SHALL display `● Repository is clean` with `enter back · q quit` key hints

### Requirement: Cleanup preview displays safe results
The system SHALL display the results of a safe-mode dry-run grouped by category with counts.

#### Scenario: Safe results with actions
- **WHEN** the scan finds 3 stale refs, 1 config duplicate, and 1 stale worktree
- **THEN** the preview SHALL display under "Safe cleanup:" each category with `●` indicator and count

#### Scenario: Safe results with no-ops
- **WHEN** the scan finds no stale refs
- **THEN** the preview SHALL display `· No stale remote refs` using the pending indicator

### Requirement: Aggressive upgrade offer
The system SHALL display an aggressive cleanup section when the dry-run detects items that aggressive mode would clean beyond safe mode.

#### Scenario: Aggressive branches available
- **WHEN** the scan finds 5 merged branches not in any worktree
- **THEN** the preview SHALL display "Aggressive cleanup available:" with a count, the first 2-3 branch names inline, and "and N more — ? for details" if more exist

#### Scenario: Inline preview with few branches
- **WHEN** the scan finds 2 merged branches
- **THEN** the preview SHALL display both branch names inline with no "and N more" text

#### Scenario: No aggressive items
- **WHEN** the scan finds nothing additional for aggressive mode
- **THEN** the aggressive section SHALL NOT appear, and the `a` key hint SHALL NOT be shown

#### Scenario: Key hints with aggressive available
- **WHEN** aggressive cleanup items are available
- **THEN** the key hints SHALL show `enter run safe · a run aggressive · ? details · esc back`

#### Scenario: Key hints without aggressive
- **WHEN** no aggressive items are available
- **THEN** the key hints SHALL show `enter run safe · esc back`

### Requirement: Detail portal for aggressive cleanup
The user SHALL be able to press `?` from the cleanup preview to open a scrollable detail portal showing the full list of branches that aggressive mode would delete, with metadata.

#### Scenario: Open details
- **WHEN** the user presses `?` on the cleanup preview with aggressive items available
- **THEN** the detail portal SHALL open with title "Aggressive Cleanup Details" and a scrollable list of branch names with merge date and last commit subject

#### Scenario: No details when no aggressive items
- **WHEN** the user presses `?` on the cleanup preview with no aggressive items
- **THEN** nothing SHALL happen

### Requirement: Execute cleanup from preview
The user SHALL be able to execute safe or aggressive cleanup directly from the preview.

#### Scenario: Run safe cleanup
- **WHEN** the user presses Enter on the cleanup preview
- **THEN** the system SHALL execute cleanup in safe mode and transition to the cleanup result view

#### Scenario: Run aggressive cleanup
- **WHEN** the user presses `a` on the cleanup preview (aggressive items available)
- **THEN** the system SHALL execute cleanup in aggressive mode and transition to the cleanup result view

#### Scenario: Aggressive cleanup requires confirmation
- **WHEN** the user presses `a` on the cleanup preview
- **THEN** the system SHALL show a brief confirmation: "Delete N branches? y/n" before proceeding

### Requirement: Dry-run API
The cleanup package SHALL expose a `DryRun(runner, repoPath) DryRunResult` function returning structured results for both safe and aggressive modes from a single scan.

#### Scenario: DryRunResult structure
- **WHEN** `DryRun` is called
- **THEN** the result SHALL contain: `StaleRefs int`, `ConfigDuplicates int`, `GoneBranches []BranchInfo`, `OrphanedConfigs int`, `StaleWorktrees int`, `NonWtBranches []BranchInfo` (branches aggressive would delete), each `BranchInfo` containing Name, MergeDate, and LastCommitSubject

#### Scenario: Empty repository
- **WHEN** `DryRun` is called on a clean repository
- **THEN** all counts SHALL be zero and all slices SHALL be empty
