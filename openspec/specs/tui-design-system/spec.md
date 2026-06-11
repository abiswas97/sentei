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
Layout constants SHALL be defined in `internal/tui/constants.go` and used by all progress views. Progress hold timing already exists on the Model (`minProgressDuration` via `WithMinProgressDuration` and `holdOrAdvance`) and SHALL remain the single timing mechanism; no parallel timing constant is introduced.

#### Scenario: Windowing constants
- **WHEN** windowing logic needs completed trail or pending lead counts
- **THEN** it SHALL use `WindowCompletedTrail` and `WindowPendingLead` from `constants.go`

#### Scenario: Progress bar width constant
- **WHEN** any view renders the overall progress bar
- **THEN** it SHALL use the shared bar width constant from `constants.go`

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

