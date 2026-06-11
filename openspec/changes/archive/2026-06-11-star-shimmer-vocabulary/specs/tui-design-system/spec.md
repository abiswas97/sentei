# tui-design-system Delta

## MODIFIED Requirements

### Requirement: One working spinner
A single animated indicator SHALL mark everything sentei is actively doing: the star twinkle (`· ✢ ✳ ✻ ✽ ✻ ✳ ✢`, single-cell frames, 120ms), used by active phase lines, step lines, the windowed-steps stat line, the cleanup running line, the cleanup scan, and the menu worktree load. All motion derives from one deterministic clock: a gated tick counter from which star frames, star colors, and shimmer band positions are computed as pure functions. Ticks run only while a working surface is visible, and each entry path starts the tick chain in exactly one place (Init when the model starts in a working state, the dispatch wrapper on transitions into one). Pure layout constructions without a live frame SHALL fall back to the static `✻`.

#### Scenario: Active phase twinkles
- **WHEN** a determinate progress view renders a phase that is running
- **THEN** the phase indicator SHALL be the star's current frame

#### Scenario: One vocabulary
- **WHEN** any two working surfaces are visible in the same session (e.g. a scan, then a removal)
- **THEN** both SHALL render frames from the same clock

#### Scenario: Ticks gated to visible work
- **WHEN** no working surface is on screen
- **THEN** motion ticks SHALL stop

#### Scenario: No double tick chains
- **WHEN** a flow entry path would start the motion clock while a tick chain is already running
- **THEN** at most one tick chain SHALL drive the clock

#### Scenario: Pure layouts stay static
- **WHEN** a `ProgressLayout` is rendered without injected motion
- **THEN** active indicators SHALL render the static `✻` fallback

### Requirement: Verdict and state glyphs
The star family carries the item lifecycle and the success verdict: pending `·` (the star's resting frame), working star morph, done `✦` on item rows, and `✦` in bold success color on operation-level summary headlines. `✗` marks failure at both levels; `⚠` marks warnings; the cleanup preview's would-act lines use accent `▸`. The full circle `●` and the checkmark `✓` SHALL NOT appear anywhere in the TUI. The core rule: anything moving is being worked on; anything still is settled.

#### Scenario: Success headline gets the crystallized star
- **WHEN** a summary or result view renders its operation-level success line
- **THEN** the line SHALL lead with `✦` in the success color

#### Scenario: Done rows crystallize
- **WHEN** a list of steps, actions, or items renders a completed item
- **THEN** the item SHALL be marked `✦` alongside its `·`/working-frame peers

#### Scenario: No full circles
- **WHEN** any view renders
- **THEN** the glyph `●` SHALL NOT appear in the output

## ADDED Requirements

### Requirement: Working-text shimmer
Every working line SHALL carry a gradient shimmer band sweeping its text in the text's own color family: accent base to light peak on phase headlines, scans, and the cleanup running line; body base to white peak on working step labels; dim ramp on menu loading hints. The star sits inside its line's band. Counts, percentages, and meta text do not shimmer; done and pending lines are still. Ramp endpoints are palette tokens, adapted per theme.

#### Scenario: Phase headline shimmers accent
- **WHEN** a phase is running
- **THEN** its label SHALL render with the accent shimmer band, the star inside the band

#### Scenario: Working steps shimmer in body color
- **WHEN** a step is running
- **THEN** its label SHALL render with the body shimmer band

#### Scenario: Shimmer preserves content
- **WHEN** any text is shimmered
- **THEN** the stripped text SHALL equal the input and the rune count SHALL be unchanged

#### Scenario: Settled text is still
- **WHEN** a line is done or pending
- **THEN** it SHALL render with static styling only
