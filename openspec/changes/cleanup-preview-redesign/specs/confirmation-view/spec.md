## MODIFIED Requirements

### Requirement: Confirmation view shows equivalent CLI command
The confirmation view SHALL display the equivalent CLI command in summary/result views instead of in the confirmation dialog, so users discover the CLI equivalent after seeing what happened.

#### Scenario: CLI command in removal summary
- **WHEN** the removal summary view is rendered after deleting worktrees
- **THEN** the summary SHALL include the equivalent CLI command (e.g., `sentei remove --branches feature-a,feature-b`)

#### Scenario: CLI command in cleanup result
- **WHEN** the cleanup result view is rendered after cleanup completes
- **THEN** the result SHALL include the equivalent CLI command (e.g., `sentei cleanup --mode safe`)

#### Scenario: CLI command not in confirmation dialog
- **WHEN** any confirmation dialog is rendered (CLI path)
- **THEN** the CLI command echo SHALL still be shown (CLI users benefit from seeing what they asked for)
