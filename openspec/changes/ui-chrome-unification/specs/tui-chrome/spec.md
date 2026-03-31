## ADDED Requirements

### Requirement: View title rendering
The system SHALL provide a `viewTitle(title string) string` function that renders `  sentei ─ <title>` using bold white (color 15) styling.

#### Scenario: Standard title
- **WHEN** `viewTitle("Removing Worktrees")` is called
- **THEN** the output SHALL contain "sentei ─ Removing Worktrees" with bold white styling and 2-space left padding

#### Scenario: Empty title
- **WHEN** `viewTitle("")` is called
- **THEN** the output SHALL contain "sentei ─" with no trailing text

### Requirement: Separator rendering
The system SHALL provide a `viewSeparator(width int) string` function that renders a dotted line (`┄`) in gray (color 241) spanning `width - 4` characters with 2-space left padding.

#### Scenario: Standard width
- **WHEN** `viewSeparator(80)` is called
- **THEN** the output SHALL contain 76 `┄` characters with 2-space left padding in gray styling

#### Scenario: Narrow width
- **WHEN** `viewSeparator(4)` or `viewSeparator(0)` is called
- **THEN** the output SHALL fall back to a default width of 40

### Requirement: Key hints rendering
The system SHALL provide a `viewKeyHints(hints ...KeyHint) string` function that renders key-action pairs separated by ` · ` in gray (color 241) with 2-space left padding.

#### Scenario: Multiple hints
- **WHEN** `viewKeyHints(KeyHint{"enter", "confirm"}, KeyHint{"esc", "back"}, KeyHint{"q", "quit"})` is called
- **THEN** the output SHALL render `  enter confirm · esc back · q quit` in gray styling

#### Scenario: Single hint
- **WHEN** `viewKeyHints(KeyHint{"q", "quit"})` is called
- **THEN** the output SHALL render `  q quit` with no separator

### Requirement: Progress layout rendering
The system SHALL provide a `ProgressLayout` struct with a `View() string` method that renders the standard progress view layout: title, optional subtitle, separator, phases with steps, separator, overall progress bar, and key hints.

#### Scenario: Complete layout with subtitle
- **WHEN** a `ProgressLayout` is rendered with Title="Creating Worktree", Subtitle="feature/foo → from main", and 3 phases
- **THEN** the output SHALL contain the title line, subtitle in accent color (62), top separator, phase sections, bottom separator, and overall progress bar

#### Scenario: Layout without subtitle
- **WHEN** a `ProgressLayout` is rendered with no Subtitle
- **THEN** the output SHALL omit the subtitle line and render title directly followed by the separator

#### Scenario: Completed phase collapses steps
- **WHEN** a phase has Done == Total and no failed steps
- **THEN** the phase SHALL render as a single line with `100% ●` and no individual step lines

#### Scenario: Completed phase with failures keeps steps visible
- **WHEN** a phase has Done == Total but Failed > 0
- **THEN** the phase SHALL render with expanded steps so failed items are visible

#### Scenario: Active phase shows steps
- **WHEN** a phase has Done < Total
- **THEN** the phase SHALL render with expanded step lines using status indicators (● done, ◐ active, · pending, ✗ failed)

#### Scenario: Pending phase
- **WHEN** a phase has Total == 0 (no events received yet)
- **THEN** the phase SHALL render as a single line with "pending" in gray text

#### Scenario: Step indentation
- **WHEN** steps are rendered under a phase header
- **THEN** each step SHALL use 4-space indentation (phase headers use 2-space)

#### Scenario: Overall progress bar
- **WHEN** the layout is rendered with phases totaling 10 done out of 20 total steps
- **THEN** the bottom progress bar SHALL render as `████████████░░░░░░░░ 50%` with 20-character bar width

#### Scenario: Phase status line format
- **WHEN** a phase named "Removing worktrees" has 12 done out of 30 total
- **THEN** the phase header SHALL render as `  Removing worktrees     12/30  40%` with the count and percentage right-aligned

### Requirement: Adaptive windowing for step lists
The system SHALL provide a `WindowSteps(steps []ProgressStep, availableLines int) WindowResult` function that selects which steps to display when the list exceeds available terminal space.

#### Scenario: All items fit
- **WHEN** 5 steps are provided with availableLines=10
- **THEN** all 5 steps SHALL be returned and `Windowed` SHALL be false

#### Scenario: Items exceed budget
- **WHEN** 30 steps are provided with availableLines=8
- **THEN** `Windowed` SHALL be true, `Showing` SHALL equal the number of visible items, and the stat line budget (1 line) SHALL be reserved

#### Scenario: Failed items always visible
- **WHEN** 30 steps include 3 failed items and availableLines=5
- **THEN** all 3 failed items SHALL appear in the visible set regardless of budget pressure

#### Scenario: Active items always visible
- **WHEN** 30 steps include 5 active items and availableLines=6
- **THEN** all 5 active items SHALL appear in the visible set

#### Scenario: Remaining budget fills with recent completed and next pending
- **WHEN** windowing is active with budget remaining after failed + active items
- **THEN** the remaining lines SHALL show the most recently completed items first, then the next pending items

#### Scenario: Budget of zero
- **WHEN** availableLines=0
- **THEN** only failed and active items SHALL be returned (minimum viable display)

### Requirement: Stat line rendering
The system SHALL provide a `viewStatLine(stats WindowStats) string` function that renders an indicator legend with counts when windowing is active.

#### Scenario: Standard stat line
- **WHEN** stats show 10 done, 3 active, 17 pending, 0 failed, showing 6 of 30
- **THEN** the output SHALL render `    ● 10 done · ◐ 3 active · · 17 pending  showing 6 of 30` with appropriate indicator colors and gray text

#### Scenario: Stat line with failures
- **WHEN** stats include 2 failed items
- **THEN** the output SHALL include `✗ 2 failed` in the legend in red (color 196)

#### Scenario: Zero count omission
- **WHEN** a status category has zero items (e.g., 0 failed)
- **THEN** that category SHALL be omitted from the stat line

### Requirement: Animation buffer for fast operations
The system SHALL provide a `bufferTransition(started time.Time, cmd tea.Cmd) tea.Cmd` function that ensures at least `MinProgressDisplay` duration elapses before delivering the inner Cmd's message.

#### Scenario: Fast operation buffered
- **WHEN** an operation completes in 50ms and MinProgressDisplay is 300ms
- **THEN** the wrapper Cmd SHALL wait an additional 250ms before returning the message

#### Scenario: Slow operation not delayed
- **WHEN** an operation completes in 500ms and MinProgressDisplay is 300ms
- **THEN** the wrapper Cmd SHALL return the message immediately with no additional delay

#### Scenario: Test override
- **WHEN** MinProgressDisplay is set to 0 in test code
- **THEN** the wrapper Cmd SHALL return the message immediately regardless of elapsed time
