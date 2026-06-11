# tui-design-system Delta

## ADDED Requirements

### Requirement: Terminal tab presence
The root view SHALL set the terminal window title to `sentei · <repo>` at rest and `sentei · <repo> · <verb> <done>/<total>` during live flows, and SHALL mirror flow progress into the terminal's native progress indicator (OSC 9;4): indeterminate during spinner waits, the completion value during progress flows, the error state when a phase has failures, and none otherwise.

#### Scenario: Title at rest
- **WHEN** the menu is shown for repository "repo"
- **THEN** the window title SHALL be `sentei · repo`

#### Scenario: Title and native progress in flight
- **WHEN** a removal is running at 2 of 3
- **THEN** the title SHALL be `sentei · repo · removing 2/3` and the native progress SHALL carry the matching value

#### Scenario: Failures surface natively
- **WHEN** any phase reports failed steps
- **THEN** the native progress SHALL use the error state
