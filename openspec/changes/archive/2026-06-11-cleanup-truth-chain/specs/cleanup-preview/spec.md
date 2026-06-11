# cleanup-preview Delta

## ADDED Requirements

### Requirement: Aggressive preview states effective counts
The aggressive section's headline SHALL state how many branches aggressive mode would actually delete: the full count when all candidates are deletable, `N of M` when some are unmerged, and an explicit "none deletable without --force" form when zero. The `a` affordance (footer hint and confirm gate) SHALL exist only when the effective count is positive, and the clean preview SHALL list the safe-cleanup checks so coverage is visible before running.

#### Scenario: Mixed candidates
- **WHEN** three candidates exist and two are unmerged
- **THEN** the headline SHALL read `1 of 3 branches` would be deleted and the confirm prompt SHALL say `Delete 1 branch? (2 unmerged will be skipped)`

#### Scenario: Nothing deletable
- **WHEN** every candidate is unmerged
- **THEN** the headline SHALL state none are deletable without --force, the `a` hint SHALL be absent, and pressing `a` SHALL NOT open a confirm prompt

#### Scenario: Clean repository teaches coverage
- **WHEN** the scan finds nothing
- **THEN** the preview SHALL list each safe-cleanup check as a dim no-op line
