# tui-design-system Delta

## ADDED Requirements

### Requirement: Milestone whisper
The repository state SHALL carry a lifetime count of worktrees removed through the TUI. When a removal run crosses a power of ten, the removal summary SHALL show one dim line acknowledging it; otherwise no line renders. State errors SHALL degrade silently — the whisper never becomes a warning.

#### Scenario: Crossing a power of ten
- **WHEN** a run takes the lifetime count from below a power of ten to at or above it
- **THEN** the summary SHALL whisper that milestone in dim text

#### Scenario: Ordinary runs stay quiet
- **WHEN** a run crosses no power of ten
- **THEN** the summary SHALL render no whisper line

#### Scenario: Garnish never alarms
- **WHEN** the state file cannot be read or written
- **THEN** the summary SHALL render normally with no whisper and no error
