# Animated Progress Bar and Elapsed Time

## Why

The overall progress bar jumps in event-sized steps and long operations (30+ worktrees) give no sense of running time. Approved 2026-06-11: adopt `bubbles/progress` (spring-smoothed motion between updates) and `bubbles/stopwatch` (elapsed readout) on the determinate progress views.

## What Changes

- The overall bar in all four progress flows (removal, create, repo/migrate, integrations) animates smoothly toward each new completion target via `bubbles/progress`, styled to the existing spec: 20 cells, accent `█` fill, dim `░` track, percentage text in the default foreground showing actual (not animated) progress.
- An `elapsed Ns` readout (dim) renders beside the bar, driven by `bubbles/stopwatch`, reset at each flow start.
- Each flow's `ProgressLayout` construction is extracted into a layout method shared by the view and the spring-target computation (single source for done/total).
- `ProgressLayout` rendered directly (tests, fallback) keeps the static bar; the animated bar is injected by the model.
- fang+cobra is recorded as declined in the decision log (CLI layer is small and tested; revisit if the subcommand surface grows).

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `tui-chrome`: the styled-progress-bar requirement gains animation semantics (spring toward target, percentage shows actual progress).

## Impact

- **Code**: `progress_layout.go` (overall extraction, Bar/Elapsed fields), `model.go` (bar + stopwatch state, FrameMsg/TickMsg routing, syncProgressBar), the four flow files (layout extraction, sync cmd, stopwatch start), `.impeccable.md`.
- **Tests**: layout/overall extraction tests, FrameMsg gating, spring-target sync; existing renderProgressBar tests unchanged (static fallback preserved).
