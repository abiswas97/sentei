## Why

Four small honesty leaks measured by the June 2026 UX audit remain after the structural fixes: `elapsed 0s` renders beside a 46%-filled bar (floor-truncated seconds are noise on short flows); failed phase headers read `✗ feat-1 2/2 100%`, conflating attempted with succeeded; the failure summary is titled "Apply complete" with the green count leading; and skip-install detection is invisible, so a wrong detection (the ccc incident class) has no surface to be noticed on.

## What Changes

- The elapsed readout is suppressed until elapsed >= 2s; its layout reserve is kept so the bar does not reflow when it appears.
- Phase headers for phases containing failures drop the percentage; counts remain (`✗ feat-1 2/2`).
- The apply failure summary gets its own title const in `copy.go` ("Apply finished with errors") and the headline leads with the failed count.
- Skipped steps render a dim per-step trace line during progress and on the summary (`· Install ccc – skipped (already installed)`), reusing the existing dim skipped vocabulary.

## Capabilities

### New Capabilities

<!-- none -->

### Modified Capabilities
- `tui-progress`: added requirements for elapsed suppression, failed-phase header format, and skipped-step traces.
- `integration-apply-summary`: the outcome-summary requirement's failure presentation changes (title and headline ordering).

## Impact

- Affected code: `internal/tui` (progress_layout.go, integration_summary.go, copy.go), golden tests for affected stable views regenerated with frame review.
- Depends on `progress-declarations` for stable failed-phase semantics (sequenced after it); elapsed suppression and skip traces have no dependency.
- FAQ decisions locked: dim per-step skip line; drop-percent-on-failure (no new header format).
