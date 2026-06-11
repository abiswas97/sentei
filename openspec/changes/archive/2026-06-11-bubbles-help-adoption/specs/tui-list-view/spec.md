# tui-list-view Delta

## MODIFIED Requirements

### Requirement: Status bar
The TUI SHALL display a status bar at the bottom showing the count of selected worktrees, filter state, available key bindings, and a status indicator legend. The sort indicator SHALL appear in the column headers (not the status bar). The legend SHALL appear on a separate line below the keybindings line. When filter mode is active, the legend line SHALL be replaced with contextual key hints (`enter: apply | esc: cancel`). The key hints SHALL derive from the list view's bindings declared in `keys.go` and SHALL include `?` (details) so the detail portal is discoverable. The hint subset is curated to fit the 80-column minimum beside the selection-count prefix; select-all and sort remain documented in the help sections.

#### Scenario: Status bar content
- **WHEN** 3 worktrees are selected and no filter is active
- **THEN** the status bar SHALL display the selection count and key hints (space: toggle, enter: delete, /: filter, ?: details, q: quit) within 80 columns

#### Scenario: Status bar with active filter
- **WHEN** a filter is applied with text "feat" matching 5 of 12 worktrees
- **THEN** the status bar SHALL include filter info such as `filter: "feat" (5/12)`

#### Scenario: Legend line content
- **WHEN** the list view is displayed
- **THEN** a legend line SHALL appear below the keybindings line showing all four status indicators with labels: `[ok] clean  [~] dirty  [!] untracked  [L] locked`

#### Scenario: Legend indicator colors
- **WHEN** the legend line is rendered
- **THEN** each indicator SHALL use the same color style as its corresponding in-table indicator (`[ok]` green, `[~]` orange, `[!]` red, `[L]` gray) and the labels SHALL use a dimmed style

#### Scenario: Legend replaced during filter mode
- **WHEN** the user is actively typing in the filter input
- **THEN** the legend line SHALL be replaced with `enter: apply | esc: cancel` in dimmed style

#### Scenario: Portal discoverable from the status bar
- **WHEN** the list view status bar is rendered with no active filter
- **THEN** it SHALL include a `?` hint derived from the details binding
