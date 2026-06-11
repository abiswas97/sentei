# tui-design-system Delta

## MODIFIED Requirements

### Requirement: Layout constants defined centrally
Layout constants SHALL be defined in `internal/tui/constants.go` and used by all progress views. Progress hold timing already exists on the Model (`minProgressDuration` via `WithMinProgressDuration` and `holdOrAdvance`) and SHALL remain the single timing mechanism; no parallel timing constant is introduced.

#### Scenario: Windowing constants
- **WHEN** windowing logic needs completed trail or pending lead counts
- **THEN** it SHALL use `WindowCompletedTrail` and `WindowPendingLead` from `constants.go`

#### Scenario: Progress bar width floor
- **WHEN** any view renders the overall progress bar
- **THEN** the bar width SHALL be the content width minus the rendered width of its right-hand meta (percentage and elapsed readout), never narrower than the shared floor constant in `constants.go`

## ADDED Requirements

### Requirement: Animated active indicator
The active status indicator SHALL be an animated "breathing dot" cycling `·` `∙` `●` `∙`, replacing the static `◐` everywhere live work is marked: phase lines, step lines, the windowed-steps stat line, and the cleanup result running line. Every frame SHALL be a single cell so status columns stay aligned. Pure layout constructions without a live frame SHALL fall back to the static midpoint `∙`.

#### Scenario: Active phase breathes
- **WHEN** a determinate progress view renders a phase that is running
- **THEN** the phase indicator SHALL be the current breath frame, styled as the active indicator

#### Scenario: Ticks gated to live progress
- **WHEN** no determinate progress view (or cleanup running line) is on screen
- **THEN** breath spinner ticks SHALL stop

#### Scenario: Pure layouts stay static
- **WHEN** a `ProgressLayout` is rendered without an injected live frame
- **THEN** active indicators SHALL render the static `∙` fallback

### Requirement: Gradient bar fill
The overall progress bar's filled portion SHALL blend from the `barStart` to the `barEnd` palette token, scaled to span exactly the filled cells, in both themes. The fill characters remain `█`/`░`. The static fallback bar (pure constructions) keeps the solid accent fill.

#### Scenario: Gradient spans the fill
- **WHEN** the animated overall bar renders at any percentage above zero
- **THEN** the filled cells SHALL blend `barStart` to `barEnd`, with the blend endpoints at the first and last filled cell

#### Scenario: Adaptive endpoints
- **WHEN** the light palette is active
- **THEN** the blend SHALL use the light palette's `barStart`/`barEnd` tokens
