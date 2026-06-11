# Proposal: ccc-python-pin

## Why

The ccc install resolves against whatever Python uv finds ambient — a `.python-version` pinning 3.10 in the caller's directory makes the >=3.11 requirement unsatisfiable (user-reported).

## What Changes

- The install command pins `--python 3.11`, matching the integration's declared python3.11+ dependency, so resolution is deterministic regardless of where sentei runs.

## Capabilities

### Modified

- `integration-apply-summary`: install determinism.

## Impact

- `internal/integration/ccc.go`; contract test extended.
