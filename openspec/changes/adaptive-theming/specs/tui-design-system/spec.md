# tui-design-system Delta

## ADDED Requirements

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
