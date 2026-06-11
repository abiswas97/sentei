# tui-chrome Delta

## MODIFIED Requirements

### Requirement: Styled overall progress bar
The system SHALL render the overall progress bar with the filled portion in the accent color and the unfilled track in dim; the percentage label SHALL use the default foreground and SHALL reflect the displayed fill, so the bar and its label never disagree. Actual completion counts SHALL remain visible in the phase headers. In live progress views the bar fill SHALL animate smoothly (spring easing) toward each new completion target and SHALL visibly settle at the target within the completion hold; a dim `elapsed Ns` readout SHALL render beside the bar. A bar SHALL never render uncolored.

#### Scenario: Bar colors
- **WHEN** a progress view renders a bar at 53%
- **THEN** the filled `█` cells SHALL carry the accent token and the `░` track cells SHALL carry the dim token

#### Scenario: Bounds clamped
- **WHEN** the done count exceeds the total for any reason
- **THEN** the bar SHALL clamp to 100% and never panic on a negative repeat count

#### Scenario: Label follows the fill
- **WHEN** the fill is easing through 40% toward a 100% target
- **THEN** the label SHALL read 40%, and the phase headers SHALL state the actual completion counts

#### Scenario: Completion settles within the hold
- **WHEN** a flow completes and the view holds before transitioning
- **THEN** the bar SHALL visibly reach a full fill with a 100% label during the hold

#### Scenario: Elapsed readout
- **WHEN** a progress flow has been running for 12 seconds
- **THEN** the bar line SHALL include a dim `elapsed 12s` readout
