# tui-design-system Specification

## Purpose
TBD - created by archiving change ui-chrome-unification. Update Purpose after archive.
## Requirements
### Requirement: Component patterns documented in .impeccable.md
The `.impeccable.md` file SHALL include a "Component Patterns" section documenting the standard view chrome, progress layout, status indicators, stat line format, timing constants, and key mapping.

#### Scenario: View chrome documented
- **WHEN** a developer reads `.impeccable.md`
- **THEN** the Component Patterns section SHALL specify: title format (`sentei ─ <Title>` via `viewTitle()`), separator format (dotted `┄` line via `viewSeparator()`), and key hints format (`key action · key action` via `viewKeyHints()`)

#### Scenario: Status indicators documented
- **WHEN** a developer reads `.impeccable.md`
- **THEN** the Component Patterns section SHALL include a table mapping each indicator symbol (●, ◐, ·, ✗, ⚠) to its name, color, and semantic meaning

#### Scenario: Key mapping documented
- **WHEN** a developer reads `.impeccable.md`
- **THEN** the Component Patterns section SHALL specify `?` as contextual details, `F1` as global help, and all standard navigation keys (j/k, arrows, enter, esc, q, space)

### Requirement: Key bindings defined in single source file
All key bindings SHALL be defined in `internal/tui/keys.go` as `key.Binding` variables, with contextual keys (meaning varies by view) and global keys (same meaning everywhere) clearly separated. Per-view key presentation (footer hint subsets and named help sections) SHALL also be declared in `keys.go`, reusing the canonical key strings and overriding only descriptions; render sites SHALL reference these declarations and SHALL NOT contain raw key or label strings.

#### Scenario: Contextual details key
- **WHEN** `keyDetails` is referenced
- **THEN** it SHALL bind to `?` with help text "details"

#### Scenario: Global help key
- **WHEN** `keyGlobalHelp` is referenced
- **THEN** it SHALL bind to `F1` with help text "help"

#### Scenario: No duplicate key definitions
- **WHEN** a key binding is needed in any view
- **THEN** it SHALL reference the binding from `keys.go` rather than creating a local binding

#### Scenario: Contextual description declared once per view
- **WHEN** the enter key means "delete" in the confirm view and "continue" in an input view
- **THEN** each description SHALL be declared once in that view's `keys.go` presentation data and nowhere else

#### Scenario: Render sites carry no hint literals
- **WHEN** any view renders its footer or the help portal renders its content
- **THEN** the key strings and action labels SHALL come from `keys.go` declarations, with no string literals at the render site

### Requirement: Layout constants defined centrally
Layout constants SHALL be defined in `internal/tui/constants.go` and used by all progress views. Progress hold timing already exists on the Model (`minProgressDuration` via `WithMinProgressDuration` and `holdOrAdvance`) and SHALL remain the single timing mechanism; `progressSettleFloor` is part of that mechanism (the post-completion settle beat when holds are enabled), not a parallel one.

#### Scenario: Windowing constants
- **WHEN** windowing logic needs completed trail or pending lead counts
- **THEN** it SHALL use `WindowCompletedTrail` and `WindowPendingLead` from `constants.go`

#### Scenario: Progress bar width floor
- **WHEN** any view renders the overall progress bar
- **THEN** the bar width SHALL be the content width minus the rendered width of its right-hand meta (percentage and elapsed readout), never narrower than the shared floor constant in `constants.go`

### Requirement: Charm v2 rendering platform
The TUI SHALL build on the Charm v2 stack (`charm.land/bubbletea/v2`, `charm.land/bubbles/v2`, `charm.land/lipgloss/v2`). Terminal features (alternate screen, mouse mode, keyboard enhancements) SHALL be declared as fields on the root model's `tea.View`; no view other than the root model SHALL declare terminal features.

#### Scenario: Alt screen declared on the root view
- **WHEN** the root model's `View()` is rendered
- **THEN** it SHALL return a `tea.View` with the alternate screen enabled, and `tea.NewProgram` SHALL receive no terminal-feature options

#### Scenario: Rendered content unchanged by the platform swap
- **WHEN** any existing view renders after the migration
- **THEN** its visible content (text, layout, colors, indicators) SHALL be unchanged from the v1 implementation

### Requirement: Keyboard enhancements for quick create
The TUI SHALL request keyboard enhancements so terminals supporting the kitty keyboard protocol can distinguish `ctrl+enter` from `enter`. The quick-create binding SHALL start worktree creation with default options directly from the branch input on supporting terminals.

#### Scenario: Quick create on a supporting terminal
- **WHEN** the user presses `ctrl+enter` on the create-branch input with a valid branch name on a terminal with the kitty keyboard protocol
- **THEN** creation SHALL start immediately with default options, skipping the options view

#### Scenario: Graceful degradation without the protocol
- **WHEN** the terminal does not support keyboard enhancements
- **THEN** `enter` (continue to options) SHALL remain fully functional and no error SHALL surface

### Requirement: Adaptive palette
The TUI SHALL detect the terminal background at startup via `tea.RequestBackgroundColor` and select the palette accordingly: the dark palette on dark backgrounds and the light palette on light backgrounds. Both palettes SHALL be declared as data in `internal/tui/styles.go`, one value per token, and documented side by side in `.impeccable.md`. When the terminal does not report a background color, the dark palette SHALL remain active.

#### Scenario: Light terminal gets the light palette
- **WHEN** the terminal reports a light background (`BackgroundColorMsg.IsDark()` is false)
- **THEN** all subsequent renders SHALL use the light palette values for every token

#### Scenario: Dark terminal keeps the dark palette
- **WHEN** the terminal reports a dark background
- **THEN** rendering SHALL be unchanged from the pre-detection output

#### Scenario: No background report defaults to dark
- **WHEN** the terminal never responds to the background query
- **THEN** the dark palette SHALL remain active and no error SHALL surface

### Requirement: Indeterminate wait indicator
Indeterminate waits (operations with no measurable progress: the cleanup repository scan and the menu's worktree-context load) SHALL display an animated spinner in the accent color. Determinate progress SHALL keep the static indicator vocabulary and the progress bar. Spinner ticks SHALL run only while an indeterminate wait is visible.

#### Scenario: Cleanup scan animates
- **WHEN** the cleanup preview is entered and the scan has not completed
- **THEN** the scanning line SHALL render an animated spinner frame followed by "Scanning repository…"

#### Scenario: Menu load animates
- **WHEN** the bare-repo menu is shown before the worktree context has loaded
- **THEN** the "Remove worktrees" item hint SHALL render an animated spinner frame followed by "loading…"

#### Scenario: Ticks stop when the wait ends
- **WHEN** a spinner tick message arrives and no indeterminate wait is visible
- **THEN** the model SHALL NOT schedule another tick

#### Scenario: Determinate progress unchanged
- **WHEN** any progress view renders phases and steps
- **THEN** step states SHALL keep the static indicator vocabulary (`◐` active) and the overall bar

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

### Requirement: Selection and danger weight
Cursor rows SHALL use the `▸` marker with selected entry text carrying the accent in menu-style views; destructive confirmation footers SHALL render their destructive hint in the warning style; worktrees SHALL be named by one canonical label across all views (short HEAD hash for detached); the confirm screen SHALL use the list's badge vocabulary; action hints SHALL NOT advertise impossible actions (no delete hint at zero selected).

#### Scenario: Danger hint weighted
- **WHEN** the deletion confirm screen renders
- **THEN** `y delete` SHALL carry the warning style while `n go back` stays dim

#### Scenario: Canonical detached label
- **WHEN** a detached worktree appears in the list and on the confirm screen
- **THEN** both SHALL show the same short HEAD hash

#### Scenario: No dead-end hints
- **WHEN** the list has zero selections
- **THEN** the status bar SHALL NOT offer `enter delete`

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

