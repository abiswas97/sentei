## ADDED Requirements

### Requirement: Completion settle is unconditional
Once a flow's final progress event has arrived, the progress view SHALL NOT advance to its summary until the displayed overall bar fill has reached completion (>= 99.5%) and a settled beat (a single named constant, 500-800ms) has elapsed since the fill first reached it, in every run mode (playground and real), for success and failure outcomes alike, with a hard timeout fallback so the view can never wedge if the fill cannot reach the threshold.

#### Scenario: Real-run removal ends settled
- **WHEN** a removal completes in a real (non-playground) run
- **THEN** the last rendered progress frame before the summary shows the bar at 100% with the success gradient

#### Scenario: Failure endings settle without success styling
- **WHEN** a flow's final event reports failures
- **THEN** the view holds until the displayed fill completes and the settled beat elapses, rendering the standard gradient, not the success gradient

#### Scenario: Settle is state-relative, not event-relative
- **WHEN** the spring takes longer than the settled beat to glide to completion after the final event
- **THEN** the beat starts only when the displayed fill reaches the threshold, so the time visibly settled is never less than the beat

### Requirement: Quit remains immediate during settle
`q` and `ctrl+c` SHALL quit immediately during the completion settle, leaving the existing stderr trace naming the in-flight operation.

#### Scenario: Quit mid-settle
- **WHEN** the user presses `q` while the settle beat is running
- **THEN** the program exits immediately with the operation trace on stderr

### Requirement: Playground keeps its entry hold
Playground mode SHALL retain its minimum progress duration (entry hold) in addition to the completion settle; real runs SHALL have no entry hold.

#### Scenario: Fast real flow gains only the settle
- **WHEN** a real-run flow completes faster than the playground entry hold
- **THEN** the only added latency before the summary is the glide-plus-beat of the completion settle
