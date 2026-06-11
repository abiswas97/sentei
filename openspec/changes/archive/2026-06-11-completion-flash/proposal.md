# Proposal: completion-flash

## Why

Wave 3a of the joy redesign: completion deserves a visible moment. The bar reaches 100% in the accent gradient and then simply cuts away; the settle beat exists (progressSettleFloor) but nothing marks success.

## What Changes

- When a flow's result has arrived (`Completed`), the bar's fill settles to a success-green gradient for the hold — the bar joins the ✦ moment. A settle-to-green rather than a flash-and-revert: a flash that reverts reads as a glitch; green that stays reads as done.
- New palette tokens `barDoneStart`/`barDoneEnd` per theme; the integration flow gains the `Completed` flag it was missing.

## Capabilities

### Modified

- `tui-design-system`: completed-bar color rule.

## Impact

- styles.go tokens, progress_layout.go render branch, integrationState.finalized; tests.
