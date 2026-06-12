## ADDED Requirements

### Requirement: Elapsed readout earns its place
The dim elapsed readout beside the overall bar SHALL render only once elapsed time is at least 2 seconds; its fixed layout reserve SHALL be maintained while hidden so the bar width does not change when the readout appears.

#### Scenario: Short flow shows no elapsed
- **WHEN** a flow completes in under 2 seconds
- **THEN** no elapsed readout renders at any point and the bar occupies the same cells it would with the readout visible

#### Scenario: Long flow gains elapsed without reflow
- **WHEN** a flow passes the 2 second mark
- **THEN** the elapsed readout appears in the reserved cells and the bar's width is unchanged from the prior frame

### Requirement: Failed phases do not advertise percentages
A phase header whose phase contains failed steps SHALL render its counts without the percentage (`✗ <name>  <done>/<total>`), since the percent vocabulary elsewhere means completed work.

#### Scenario: Failed phase header format
- **WHEN** a phase resolves with 1 done and 1 failed of 2 steps
- **THEN** its header renders the failed indicator and `2/2` with no percentage

### Requirement: Skipped steps leave a visible trace
A step skipped by detection SHALL render as a dim step line stating it was skipped and why (`– skipped (already installed)`), during progress and on the flow summary, so detection decisions are auditable.

#### Scenario: Skip-install is visible
- **WHEN** integration setup detects a tool already on PATH and skips its install step
- **THEN** the progress view and the apply summary both show a dim skipped line naming the step and the reason
