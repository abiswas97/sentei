# Proposal: star-shimmer-vocabulary

## Why

Three motion rounds (breathing dot, braille, smooth breathing) all failed user review: glyph-frame spinners are quantized (dots have 4 sizes, braille hangs into the descender zone) and a lone `●` on summaries read as unexplained. The user chose a new direction from a live HTML lab: the Claude Code star twinkle plus Crush-style shimmer, with the full circle retired from the vocabulary entirely.

## What Changes

- **One lifecycle, one family**: pending `·` (the star's resting frame) → working star morph `· ✢ ✳ ✻ ✽` (120ms) → done `✦` (the twinkle crystallizes), with `✦` also serving as the success verdict on summary headlines. `✓` retires after one day; `✗` stays for failure; `●` is removed everywhere (invariant-tested).
- **Shimmer ensemble (V4)**: every working line carries a gradient band sweeping its text in the text's own color family — accent→light on phase headlines, grey→white on working step labels — with the star inside the band. Scans and the cleanup running line shimmer too; menu loading hints use a dim ramp.
- **Deterministic motion clock**: one tick source (60ms, gated to working surfaces exactly as today) drives star frames and shimmer band positions; every animation becomes a pure function of the tick count, replacing bubbles/spinner.
- **Cleanup preview would-act lines**: `●` (a third, future-action meaning) becomes accent `▸`.

## Capabilities

### Modified

- `tui-design-system`: status indicator vocabulary, working animation, verdict/state rule replaced by the star lifecycle.

## Impact

- `internal/tui/styles.go`: vocabulary constants, star frames, shimmer ramp tokens per palette.
- `internal/tui/motion.go` (new): motion clock, star frame/color, per-rune shimmer renderer.
- `internal/tui/model.go`: motion tick gating replaces spinner ticks; dispatch wrapper unchanged in shape.
- `internal/tui/progress_layout.go`, `chrome.go`, `cleanup_preview.go`, `cleanup_result.go`, `menu.go`, summary views: render-site updates.
- `.impeccable.md`: vocabulary table, timing, decision log.
- helpers_test.go pumpCmds drops the new tick message.
