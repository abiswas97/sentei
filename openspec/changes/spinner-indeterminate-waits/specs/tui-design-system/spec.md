# tui-design-system Delta

## ADDED Requirements

### Requirement: Indeterminate wait indicator
Indeterminate waits (operations with no measurable progress: the cleanup repository scan and the menu's worktree-context load) SHALL display an animated spinner in the accent color. Determinate progress SHALL keep the static indicator vocabulary and the progress bar. Spinner ticks SHALL run only while an indeterminate wait is visible.

#### Scenario: Cleanup scan animates
- **WHEN** the cleanup preview is entered and the scan has not completed
- **THEN** the scanning line SHALL render an animated spinner frame followed by "Scanning repository…"

#### Scenario: Menu load animates
- **WHEN** the bare-repo menu is shown before the worktree context has loaded
- **THEN** the "Remove worktrees" item hint SHALL render an animated spinner frame followed by "loading…"

#### Scenario: Ticks stop when the wait ends
- **WHEN** a spinner tick message arrives and no indeterminate wait is visible
- **THEN** the model SHALL NOT schedule another tick

#### Scenario: Determinate progress unchanged
- **WHEN** any progress view renders phases and steps
- **THEN** step states SHALL keep the static indicator vocabulary (`◐` active) and the overall bar
