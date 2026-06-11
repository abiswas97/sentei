# Proposal: progress-motion

## Why

The active indicator `◐` is a static glyph, so a running phase looks identical to a stuck one; the user explicitly asked for animated progress. The overall bar is hardcoded to 20 cells, which reads as an afterthought on wide terminals. Research into the current charm-land design language (Crush pills, gum, bubbles v2 defaults) shows animated working indicators and bars sized from the window, with the signature purple-to-pink blend.

## What Changes

- **Breathing dot**: the active indicator animates `·  ∙  ●  ∙` (the pending dot inflating toward the done dot) wherever `◐` marked live work: phase lines, step lines, the windowed-steps stat line, and the cleanup result's running line. Frozen frames still read as "between pending and done". `◐` leaves the vocabulary.
- **Responsive bar**: the overall bar fills the content width minus its right-hand meta (percentage, elapsed), floor 20 cells, driven by the existing WindowSizeMsg path.
- **Gradient fill**: the filled portion blends accent purple to selection pink, scaled to the fill (`WithColors` + `WithScaled`), with both endpoints as adaptive palette tokens (hex, so blending is smooth in both themes).

## Capabilities

### Modified

- `tui-design-system`: status indicator vocabulary (active glyph animated), progress bar sizing and fill rules.

## Impact

- `internal/tui/styles.go`: two gradient tokens per palette; `indicatorActive` replaced by breath spinner frames.
- `internal/tui/progress_layout.go`, `chrome.go`, `cleanup_result.go`: glyph injection, bar width computation.
- `internal/tui/model.go` / `update.go`: second spinner instance ticking while determinate progress views are visible (gate exists: `determinateProgressActive`).
- `.impeccable.md`: status indicator section + decision log.
- Tests asserting `◐` updated to the breath contract.
