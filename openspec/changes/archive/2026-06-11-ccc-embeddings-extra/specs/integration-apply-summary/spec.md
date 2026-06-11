# integration-apply-summary Delta

## ADDED Requirements

### Requirement: Installs are functionally complete
An integration's install command SHALL produce a tool whose default setup command can run: for cocoindex-code that means installing the `embeddings-local` extra so the default local-embedding `ccc index` has its runtime dependencies.

#### Scenario: Fresh ccc install can index
- **WHEN** sentei installs cocoindex-code on a machine without it
- **THEN** the install SHALL include the embeddings-local extra so `ccc index` does not fail on missing modules
