# tui-design-system Delta

## ADDED Requirements

### Requirement: Completion settles the bar green
While a flow is working, the overall bar fills with the accent gradient; once the flow's result has arrived, the fill SHALL render in the success gradient (`barDoneStart→barDoneEnd`, per theme) for the remainder of the hold.

#### Scenario: Working bar is accent
- **WHEN** a flow is still producing events
- **THEN** the bar fill SHALL use the accent gradient

#### Scenario: Done bar is green
- **WHEN** the flow's completion result has arrived
- **THEN** the bar fill SHALL use the success gradient until the view transitions
