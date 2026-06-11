# tui-design-system Delta

## MODIFIED Requirements

### Requirement: Layout constants defined centrally
Layout constants SHALL be defined in `internal/tui/constants.go` and used by all progress views. Progress hold timing already exists on the Model (`minProgressDuration` via `WithMinProgressDuration` and `holdOrAdvance`) and SHALL remain the single timing mechanism; `progressSettleFloor` is part of that mechanism (the post-completion settle beat when holds are enabled), not a parallel one.

#### Scenario: Windowing constants
- **WHEN** windowing logic needs completed trail or pending lead counts
- **THEN** it SHALL use `WindowCompletedTrail` and `WindowPendingLead` from `constants.go`

#### Scenario: Progress bar width floor
- **WHEN** any view renders the overall progress bar
- **THEN** the bar width SHALL be the content width minus the rendered width of its right-hand meta (percentage and elapsed readout), never narrower than the shared floor constant in `constants.go`

## ADDED Requirements

### Requirement: One working spinner
A single animated indicator SHALL mark everything sentei is actively doing: the breathing dot (`· ∙ • ● • ∙`, single-cell frames, 10fps — the pending dot inflating toward the done dot, the only frame family optically centered and weight-matched beside bold text), used by active phase lines, step lines, the windowed-steps stat line, the cleanup running line, the cleanup scan, and the menu worktree load. One spinner instance drives all of them; its ticks run only while any such work is visible, and each entry path starts the tick chain in exactly one place (Init when the model starts in a working state, the dispatch wrapper on transitions into one). Pure layout constructions without a live frame SHALL fall back to the static midpoint `∙`.

#### Scenario: Active phase spins
- **WHEN** a determinate progress view renders a phase that is running
- **THEN** the phase indicator SHALL be the spinner's current frame, styled as the active indicator

#### Scenario: One vocabulary
- **WHEN** any two working surfaces are visible in the same session (e.g. a scan, then a removal)
- **THEN** both SHALL render frames from the same spinner

#### Scenario: Ticks gated to visible work
- **WHEN** no working surface is on screen
- **THEN** spinner ticks SHALL stop

#### Scenario: No double tick chains
- **WHEN** a flow entry path would start the spinner while a tick chain is already running
- **THEN** at most one tick chain SHALL drive the spinner (frames never advance at multiple speeds)

#### Scenario: Pure layouts stay static
- **WHEN** a `ProgressLayout` is rendered without an injected live frame
- **THEN** active indicators SHALL render the static `∙` fallback

### Requirement: Gradient bar fill
The overall progress bar's filled portion SHALL blend from the `barStart` to the `barEnd` palette token, scaled to span exactly the filled cells, in both themes. The fill characters remain `█`/`░`. The spring SHALL be tuned for a visible glide: fills ease over most of a second rather than snapping, while still settling within the completion hold. The static fallback bar (pure constructions) keeps the solid accent fill.

#### Scenario: Gradient spans the fill
- **WHEN** the animated overall bar renders at any percentage above zero
- **THEN** the filled cells SHALL blend `barStart` to `barEnd`, with the blend endpoints at the first and last filled cell

#### Scenario: Adaptive endpoints
- **WHEN** the light palette is active
- **THEN** the blend SHALL use the light palette's `barStart`/`barEnd` tokens

#### Scenario: Settles within the hold
- **WHEN** a flow completes and the completion hold begins
- **THEN** the bar SHALL visibly reach full before the hold expires

#### Scenario: Flows that outlive the hold still settle
- **WHEN** a flow's final event arrives after the hold duration has already elapsed
- **THEN** the view SHALL stay up for at least the settle floor so the bar finishes at 100% instead of cutting away mid-glide

### Requirement: Element motion, instant cuts
View-to-view navigation SHALL cut instantly; no slide, fade, or transition animation may delay a keypress. Motion belongs to state-driven elements only: the bar's spring, the working spinner, and hold pacing.

#### Scenario: Navigation is immediate
- **WHEN** the user presses a key that changes views
- **THEN** the next view SHALL render immediately with no transition animation

### Requirement: Verdict and state glyphs
`✓` and `✗` are verdicts about a whole operation and SHALL mark one-line summary headlines (`✓ 3 worktrees removed successfully`); `●`, the working animation, and `·` are states of items within an operation and SHALL always render among peers. `●` SHALL NOT appear alone on a screen.

#### Scenario: Success headline gets the checkmark
- **WHEN** a summary or result view renders its operation-level success line
- **THEN** the line SHALL lead with `✓` in the success color

#### Scenario: Item rows keep state glyphs
- **WHEN** a list of steps, actions, or items renders inside a summary or progress view
- **THEN** completed items SHALL keep `●` alongside their `·`/working-frame peers
