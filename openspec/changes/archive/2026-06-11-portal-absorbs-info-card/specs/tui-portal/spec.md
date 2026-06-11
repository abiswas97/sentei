# tui-portal Delta

## ADDED Requirements

### Requirement: Integration details page
The integration list and migrate-integrations views SHALL provide their integration details through the DetailPortal via the standard `?` path: one scrollable page listing every integration with its name, description, dependency install status (`●` installed, `·` will be installed), and URL. The portal SHALL be the only overlay in the TUI; no view SHALL render its own bordered overlay.

#### Scenario: ? opens integration details
- **WHEN** the user presses `?` on the integration list
- **THEN** the portal SHALL open titled "Integration Details" with one section per integration, in list order

#### Scenario: Dependency status rows preserved
- **WHEN** an integration has a dependency that is installed and another that is not
- **THEN** its section SHALL show `●` with "installed" for the first and `·` with "will be installed" for the second

#### Scenario: Carousel keys retired
- **WHEN** the portal is open over the integration list and the user presses `h` or `l`
- **THEN** the portal SHALL treat them as it treats any non-scroll key (swallowed); no paging occurs
