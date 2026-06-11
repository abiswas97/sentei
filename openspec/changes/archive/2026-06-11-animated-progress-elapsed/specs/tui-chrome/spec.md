# tui-chrome Delta

## MODIFIED Requirements

### Requirement: Styled overall progress bar
The system SHALL render the overall progress bar with the filled portion in the accent color and the unfilled track in dim; the percentage label SHALL use the default foreground and SHALL reflect actual completed progress. In live progress views the bar fill SHALL animate smoothly (spring easing) toward each new completion target rather than jumping; a dim `elapsed Ns` readout SHALL render beside the bar. A bar SHALL never render uncolored.

#### Scenario: Bar colors
- **WHEN** a progress view renders a bar at 53%
- **THEN** the filled `█` cells SHALL carry the accent token and the `░` track cells SHALL carry the dim token

#### Scenario: Bounds clamped
- **WHEN** the done count exceeds the total for any reason
- **THEN** the bar SHALL clamp to 100% and never panic on a negative repeat count

#### Scenario: Animation toward target
- **WHEN** a step completes and the overall target moves from 40% to 50%
- **THEN** subsequent animation frames SHALL move the fill toward 50% while the percentage text reads the actual 50%

#### Scenario: Elapsed readout
- **WHEN** a progress flow has been running for 12 seconds
- **THEN** the bar line SHALL include a dim `elapsed 12s` readout
