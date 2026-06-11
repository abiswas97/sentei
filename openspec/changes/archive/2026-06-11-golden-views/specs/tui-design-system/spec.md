# tui-design-system Delta

## ADDED Requirements

### Requirement: Golden chrome pinning
The stable views (worktree list, confirm, removal summary, cleanup result, create input) SHALL be pinned by golden-file tests capturing their exact rendered output including styling. Golden updates SHALL be explicit (`-update`), never incidental.

#### Scenario: Chrome regression fails loudly
- **WHEN** any change alters a pinned view's exact output
- **THEN** the golden test SHALL fail until the golden is intentionally regenerated
