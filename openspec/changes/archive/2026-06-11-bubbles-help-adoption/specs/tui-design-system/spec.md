# tui-design-system Delta

## MODIFIED Requirements

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
