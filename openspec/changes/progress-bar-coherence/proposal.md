# Progress Bar Coherence

## Why

User-reported on the v1.9.0 demo: the animated bar shows a partial fill while its percentage text reads 100%. The shipped design had the text state actual progress while the cells eased toward it; with fast operations the spring (default frequency 18, settle ~1s) never visibly completes inside the 1.5s hold, so the only frame users dwell on is a partial bar contradicting its own label.

## What Changes

- The percentage label follows the displayed fill (the component's native animated percentage) so bar and label can never disagree; the phase lines above continue to state actual counts (`2/2 100%`).
- The spring is tuned snappier (frequency 30) so the fill visibly completes well inside the hold window.
- The manual percentage rendering in `renderProgressLayout` is removed in favor of the component's.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `tui-chrome`: the styled-progress-bar requirement's animation semantics change: the label reflects the displayed fill, not the target; the fill settles within the completion hold.

## Impact

- **Code**: `progress_layout.go` (`newOverallBar`, `renderProgressLayout`); one test assertion.
- **Verification**: re-recorded VHS demo, frame-checked for a full bar + 100% during the hold.
