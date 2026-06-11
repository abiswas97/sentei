# tui-design-system Delta

## ADDED Requirements

### Requirement: Charm v2 rendering platform
The TUI SHALL build on the Charm v2 stack (`charm.land/bubbletea/v2`, `charm.land/bubbles/v2`, `charm.land/lipgloss/v2`). Terminal features (alternate screen, mouse mode, keyboard enhancements) SHALL be declared as fields on the root model's `tea.View`; no view other than the root model SHALL declare terminal features.

#### Scenario: Alt screen declared on the root view
- **WHEN** the root model's `View()` is rendered
- **THEN** it SHALL return a `tea.View` with the alternate screen enabled, and `tea.NewProgram` SHALL receive no terminal-feature options

#### Scenario: Rendered content unchanged by the platform swap
- **WHEN** any existing view renders after the migration
- **THEN** its visible content (text, layout, colors, indicators) SHALL be unchanged from the v1 implementation

### Requirement: Keyboard enhancements for quick create
The TUI SHALL request keyboard enhancements so terminals supporting the kitty keyboard protocol can distinguish `ctrl+enter` from `enter`. The quick-create binding SHALL start worktree creation with default options directly from the branch input on supporting terminals.

#### Scenario: Quick create on a supporting terminal
- **WHEN** the user presses `ctrl+enter` on the create-branch input with a valid branch name on a terminal with the kitty keyboard protocol
- **THEN** creation SHALL start immediately with default options, skipping the options view

#### Scenario: Graceful degradation without the protocol
- **WHEN** the terminal does not support keyboard enhancements
- **THEN** `enter` (continue to options) SHALL remain fully functional and no error SHALL surface
