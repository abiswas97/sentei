# Progress Bar Coherence: Design

## Context

`bubbles/progress` renders its own percentage from `percentShown` (the spring's displayed position) when `ShowPercentage` is on; v1.9.0 disabled it and rendered the target percent manually, creating the label/fill contradiction during animation.

## Decisions

### D1: One source for what the bar says
The component's native percentage (animated, ` %3.0f%%` default format, plain default-foreground style) replaces the manual label. Truthful completion counts live in the phase headers, which is where the eye reads progress anyway. The previous "text states the target" decision is reversed: visual coherence beats label precision on a 20-cell bar.

### D2: Spring tuned to the hold
`WithSpringOptions(30, 1)` settles in roughly a third of the 1.5s hold, so every flow ends with a visibly full bar at 100% before transitioning.

## Risks / Trade-offs

- [Label briefly lags actual completion] → by design; phases state the truth, and the lag is sub-second after tuning.

## Migration Plan

Single fix PR; re-record the VHS demo and frame-verify the full-bar-at-100% hold frame. Rollback: revert.
