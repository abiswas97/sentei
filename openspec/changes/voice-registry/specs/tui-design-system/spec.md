# tui-design-system Delta

## ADDED Requirements

### Requirement: One voice, declared once
Every view and portal title SHALL be declared exactly once in the copy registry (`internal/tui/copy.go`) and referenced by const at render sites. Titles use sentence case. The same physical key SHALL carry the same verb across views unless the action genuinely differs (`?` is always "details").

#### Scenario: Title changes are one-line edits
- **WHEN** a view title needs rewording
- **THEN** exactly one declaration SHALL change

#### Scenario: Sentence case
- **WHEN** any view renders its title
- **THEN** the title SHALL be sentence case (first word capitalized only, proper nouns excepted)

### Requirement: Portal boxes carry no brand
The detail portal SHALL render its bare title inside the box; the `sentei ─` brand appears only on the view chrome behind it.

#### Scenario: Portal title is bare
- **WHEN** any portal opens
- **THEN** its title line SHALL NOT contain `sentei ─`
