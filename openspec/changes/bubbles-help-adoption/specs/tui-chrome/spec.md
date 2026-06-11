# tui-chrome Delta

## MODIFIED Requirements

### Requirement: Key hints rendering
The system SHALL render view footers from `key.Binding` sets via `bubbles/help`, through a `viewFooter` helper that produces key-action pairs separated by ` · ` in the dim palette token with 2-space left padding. Render sites SHALL pass bindings declared in `keys.go`, never raw key or label strings.

#### Scenario: Multiple hints
- **WHEN** a footer is rendered for bindings (enter "confirm"), (esc "back"), (q "quit")
- **THEN** the output SHALL render `  enter confirm · esc back · q quit` in dim styling

#### Scenario: Single hint
- **WHEN** a footer is rendered for the single binding (q "quit")
- **THEN** the output SHALL render `  q quit` with no separator

#### Scenario: Narrow width truncates gracefully
- **WHEN** the hint row exceeds the available terminal width
- **THEN** trailing hints SHALL be dropped with an ellipsis rather than wrapping or hard-clipping
