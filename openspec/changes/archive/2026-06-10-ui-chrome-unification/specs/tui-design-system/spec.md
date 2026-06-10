## ADDED Requirements

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
All key bindings SHALL be defined in `internal/tui/keys.go` as `key.Binding` variables, with contextual keys (meaning varies by view) and global keys (same meaning everywhere) clearly separated.

#### Scenario: Contextual details key
- **WHEN** `keyDetails` is referenced
- **THEN** it SHALL bind to `?` with help text "details"

#### Scenario: Global help key
- **WHEN** `keyGlobalHelp` is referenced
- **THEN** it SHALL bind to `F1` with help text "help"

#### Scenario: No duplicate key definitions
- **WHEN** a key binding is needed in any view
- **THEN** it SHALL reference the binding from `keys.go` rather than creating a local binding

### Requirement: Layout constants defined centrally
Layout constants SHALL be defined in `internal/tui/constants.go` and used by all progress views. Progress hold timing already exists on the Model (`minProgressDuration` via `WithMinProgressDuration` and `holdOrAdvance`) and SHALL remain the single timing mechanism; no parallel timing constant is introduced.

#### Scenario: Windowing constants
- **WHEN** windowing logic needs completed trail or pending lead counts
- **THEN** it SHALL use `WindowCompletedTrail` and `WindowPendingLead` from `constants.go`

#### Scenario: Progress bar width constant
- **WHEN** any view renders the overall progress bar
- **THEN** it SHALL use the shared bar width constant from `constants.go`
