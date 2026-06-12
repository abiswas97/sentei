## MODIFIED Requirements

### Requirement: Apply ends in an outcome summary
The TUI SHALL transition from the integration progress view to an integration summary view when the apply finishes, showing per-worktree outcomes: succeeded steps with the done indicator and failed steps with the failed indicator and their error text, plus overall counts. When any step failed, the summary's title SHALL state that the apply finished with errors (a dedicated `copy.go` const, not the success title) and the headline SHALL lead with the failed count before the applied count.

#### Scenario: Fully successful apply
- **WHEN** an apply completes with all steps succeeding across 3 worktrees
- **THEN** the summary SHALL show each worktree with its completed steps and a success headline, and offer a key hint to return to the integration list

#### Scenario: Partially failed apply
- **WHEN** an apply completes with failed steps in 2 of 7 worktrees
- **THEN** the summary SHALL show the failed steps with their error text under their worktrees, the title SHALL state the apply finished with errors, and the headline SHALL lead with the failed count (`✗ 2 failed, 5 applied` ordering)

#### Scenario: State persistence failure
- **WHEN** the apply finishes but persisting the integration state to disk fails
- **THEN** the summary SHALL display the save error prominently and indicate that the shown integration set was not persisted
