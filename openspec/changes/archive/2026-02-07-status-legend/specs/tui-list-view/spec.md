## MODIFIED Requirements

### Requirement: Status bar
The TUI SHALL display a status bar at the bottom showing the count of selected worktrees, available key bindings, and a status indicator legend. The legend SHALL appear on a separate line below the keybindings line.

#### Scenario: Status bar content
- **WHEN** 3 worktrees are selected out of 10 total
- **THEN** the status bar SHALL display the selection count and key hints (space: toggle, a: all, enter: delete, q: quit)

#### Scenario: Legend line content
- **WHEN** the list view is displayed
- **THEN** a legend line SHALL appear below the keybindings line showing all four status indicators with labels: `[ok] clean  [~] dirty  [!] untracked  [L] locked`

#### Scenario: Legend indicator colors
- **WHEN** the legend line is rendered
- **THEN** each indicator SHALL use the same color style as its corresponding in-table indicator (`[ok]` green, `[~]` orange, `[!]` red, `[L]` gray) and the labels SHALL use a dimmed style

#### Scenario: Viewport height adjustment
- **WHEN** the terminal sends a WindowSizeMsg
- **THEN** the visible list height SHALL account for the legend line so the table does not overflow the terminal
