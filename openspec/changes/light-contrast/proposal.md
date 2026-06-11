# Proposal: light-contrast

## Why

The light palette's dim (245) and warning orange (166) wash out on white backgrounds — flagged in the ten-agent visual audit (P2) and deferred to Wave 2d.

## What Changes

- Light palette only: `dim` 245 → 243, `warning` 166 → 130. Dark palette untouched.

## Capabilities

### Modified

- `tui-design-system`: adaptive palette contrast.

## Impact

- Two values in `internal/tui/styles.go`; light-background VHS verification.
