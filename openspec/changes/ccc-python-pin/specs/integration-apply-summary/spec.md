# integration-apply-summary Delta

## ADDED Requirements

### Requirement: Install resolution is environment-independent
Integration install commands SHALL NOT depend on ambient interpreter pins: the ccc install pins its Python so a `.python-version` in the working directory cannot make resolution unsatisfiable.

#### Scenario: Caller directory pins an old Python
- **WHEN** sentei applies the ccc integration from a directory whose `.python-version` is below 3.11
- **THEN** the install SHALL still resolve using the pinned interpreter
