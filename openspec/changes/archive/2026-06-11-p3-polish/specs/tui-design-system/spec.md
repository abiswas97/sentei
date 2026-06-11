# tui-design-system Delta

## ADDED Requirements

### Requirement: P3 presentation rules
Sort arrows SHALL describe the displayed values' order (the Age column flips relative to its underlying date sort). Portal scroll hints SHALL appear only when content scrolls. Option footers SHALL include navigation hints. Tabbing into a prefilled input SHALL place the cursor at the end. Option-view cursors use `▸` and no view renders `●`.

#### Scenario: Age arrow matches the column
- **WHEN** the list sorts by age, date-ascending
- **THEN** the Age header SHALL show ▼ (the displayed ages descend)

#### Scenario: Fitting portal content offers no scroll keys
- **WHEN** portal content fits its viewport
- **THEN** the footer SHALL offer only close

#### Scenario: Tab lands at the end
- **WHEN** the user tabs into a field holding text
- **THEN** the cursor SHALL sit after the last character
