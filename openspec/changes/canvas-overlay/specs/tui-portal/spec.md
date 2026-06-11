# tui-portal Delta

## MODIFIED Requirements

### Requirement: DetailPortal renders scrollable overlay
The system SHALL provide a `DetailPortal` component that renders pre-formatted content in a scrollable viewport composited centered over the current background view using lipgloss cell-accurate layer compositing, with no background glyphs bleeding through the overlay region.

#### Scenario: Portal displayed over background
- **WHEN** the portal is opened with content and the current view is the removal progress view
- **THEN** the portal SHALL render as a bordered, styled box centered over the background view, which stays visible around the box

#### Scenario: Scrollable content
- **WHEN** the portal content exceeds the viewport height (terminal height minus chrome)
- **THEN** the viewport SHALL be scrollable with j/k, up/down, and page up/page down keys

#### Scenario: Short content fits without scrolling
- **WHEN** the portal content fits within the viewport
- **THEN** no scroll indicators SHALL be shown

#### Scenario: No background bleed-through
- **WHEN** the portal is open over a list with an active cursor row
- **THEN** no background glyph (including the cursor marker) SHALL appear inside or beside the portal box region
