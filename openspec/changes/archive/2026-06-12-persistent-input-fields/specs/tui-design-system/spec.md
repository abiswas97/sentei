# tui-design-system Delta

## ADDED Requirements

### Requirement: Persistent input fields
Text-input views SHALL render every field's input persistently: focus changes the label accent, never the layout. Blurred empty fields show their placeholders; the clone destination preview is always visible and tracks the URL live. Text inputs share one declared width.

#### Scenario: Focus moves only the accent
- **WHEN** the user tabs between fields
- **THEN** no line SHALL appear, disappear, or change indentation

#### Scenario: Placeholders survive blur
- **WHEN** an empty field loses focus
- **THEN** its placeholder SHALL remain visible
