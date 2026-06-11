# Proposal: integration-robustness

## Why

A real integration apply failed and shredded the UI (user-reported). Root cause chain: the default detect probe `<binary> --version` exits non-zero for CLIs that don't implement that flag (ccc exits 2), so an existing pipx install went undetected; sentei then ran `uv tool install`, which refused to overwrite pipx's executable and failed; the failure's error string — carrying the command's entire output — rendered raw into the summary with no clamp.

## What Changes

- **Detect by presence**: the default probe becomes `command -v <binary>` (POSIX, flag-agnostic), so existing installs are detected regardless of the CLI's flag conventions and the skip path fires as intended.
- **Bounded failure rendering (F2 from the lab)**: failed steps in apply/create summaries show a three-line peek — first line dim, the error's last non-empty line red (where CLI tools put `error:`), `… N more — ? for full output` — with the complete untrimmed output in the detail portal. The live progress row clamps to one truncated line (one-line rows law).

## Capabilities

### Modified

- `integrations` (or equivalent capability): detection semantics.
- `tui-design-system`: failure rendering bounds.

## Impact

- `internal/integration/manager.go` detectTool default; manager tests.
- `internal/tui/errpeek.go` (new helper) + `integration_summary.go`, `create_summary.go`, `integration_progress.go`.
- `.impeccable.md` decision log.
