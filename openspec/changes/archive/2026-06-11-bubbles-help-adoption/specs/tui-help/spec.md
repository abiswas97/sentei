# tui-help Delta

## MODIFIED Requirements

### Requirement: Help content is contextual per view
Each view SHALL provide its own help content, so the help overlay shows relevant information for the active view rather than a generic help screen. The content SHALL derive from the same per-view key sections declared in `keys.go` that drive the footer hints, so footer and help can never disagree about bindings.

#### Scenario: Different views show different help
- **WHEN** the user opens help from the list view and then from the menu view
- **THEN** each SHALL display different key bindings and descriptions relevant to that specific view

#### Scenario: Help content format
- **WHEN** help content is rendered for any view
- **THEN** it SHALL display key bindings in a formatted table: key on the left, description on the right, grouped by category (navigation, actions, etc.)

#### Scenario: Footer and help agree
- **WHEN** a binding appears in a view's footer
- **THEN** the same key and description SHALL appear in that view's help sections
