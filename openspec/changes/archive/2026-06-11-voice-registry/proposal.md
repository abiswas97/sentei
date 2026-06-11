# Proposal: voice-registry

## Why

The app's copy drifted: 28 title strings scattered across views in inconsistent Title Case, footer verbs varying per view ("info" vs "details", "select entry" vs "select"), and portal boxes repeating the `sentei ─` brand that is already on screen behind them. Approved as Wave 2c of the joy redesign.

## What Changes

- **Voice registry**: every view and portal title declared once in `copy.go`, sentence case throughout (calm, charm-land).
- **Verb consistency**: `?` means "details" everywhere; menu select is "select".
- **Portal de-brand**: portal boxes render their bare title; the brand stays on the chrome behind them.

## Capabilities

### Modified

- `tui-design-system`: copy voice and portal title rules.

## Impact

- `internal/tui/copy.go` (new), 26 view files (mechanical const sweep), keys.go, portal.go; 25 test files re-asserted.
