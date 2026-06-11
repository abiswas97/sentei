# tui-design-system Delta

## MODIFIED Requirements

### Requirement: Adaptive palette
The TUI SHALL select its palette from the terminal's reported background: dark terminals keep the dark palette (the documented baseline), light terminals get the light palette, and no report defaults to dark. Light-palette tokens SHALL hold readable contrast on white: dim text uses 243 and warning orange uses 130.

#### Scenario: Light terminal gets the light palette
- **WHEN** the terminal reports a light background
- **THEN** rendering SHALL use the light palette tokens

#### Scenario: Dark terminal keeps the dark palette
- **WHEN** the terminal reports a dark background
- **THEN** rendering SHALL use the dark palette tokens

#### Scenario: No background report defaults to dark
- **WHEN** the terminal does not answer the background query
- **THEN** rendering SHALL use the dark palette

#### Scenario: Light dim and warning are readable
- **WHEN** the light palette is active
- **THEN** dim text SHALL render in 243 and warnings in 130
