# tui-chrome Delta

## ADDED Requirements

### Requirement: Progress flows end truthfully
When a progress flow completes, its layout SHALL reach a coherent terminal state before transitioning: the overall bar's final target SHALL be 100%, and any phase that never discovered work SHALL render as a dim `– <Name>  skipped` line and SHALL NOT count as outstanding work. Quitting the application from a live progress view SHALL print a one-line stderr notice naming the interrupted operation.

#### Scenario: No-work phases at completion
- **WHEN** a worktree creation completes with no dependencies or integrations enabled
- **THEN** those phases SHALL read `– skipped` (dim) and the overall bar SHALL target and settle at 100%

#### Scenario: Mid-run pending unchanged
- **WHEN** the same phases have not yet run while the flow is still in flight
- **THEN** they SHALL keep the existing `· pending` treatment and count as outstanding

#### Scenario: Quit leaves a trace
- **WHEN** the user quits during repository creation
- **THEN** stderr SHALL carry a warning naming the interrupted operation after exit
