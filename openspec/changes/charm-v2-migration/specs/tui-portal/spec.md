# tui-portal Delta

## ADDED Requirements

### Requirement: Mouse wheel scrolls the portal viewport
While the portal is visible, mouse wheel events SHALL scroll the portal's viewport content and SHALL NOT reach the background view.

#### Scenario: Wheel scrolls portal content
- **WHEN** the portal is open with scrollable content and the user scrolls the wheel down
- **THEN** the portal content SHALL scroll down, identical to pressing `j`

#### Scenario: Background view unaffected by wheel while portal open
- **WHEN** the portal is open over the worktree list and the user scrolls the wheel
- **THEN** the list cursor SHALL NOT move
