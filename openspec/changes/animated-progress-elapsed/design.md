# Animated Progress: Design

## Context

Four flows render through `ProgressLayout`, whose `View()` computes overall done/total internally; `renderProgressBar` is a pure function. `bubbles/progress` animates via self-scheduled `FrameMsg` toward targets set with `SetPercent`; `bubbles/stopwatch` ticks once per second with id-scoped `TickMsg`.

## Goals / Non-Goals

**Goals**: spring-smoothed bar in live progress views with spec-identical styling; elapsed readout; one source of truth per flow for done/total; existing pure-render tests intact.

**Non-Goals**: animating the cleanup scan (indeterminate, owned by the spinner); per-step animation.

## Decisions

### D1: Layout construction extracted per flow
Each flow's `ProgressLayout` literal moves from its view function into a `<flow>Layout()` method; the view renders it and `syncProgressBar` reads its `overall()` (extracted from `View`) to compute the spring target. Done/total knowledge exists once per flow.

### D2: Model owns the components; ProgressLayout stays pure
`Model.bar` (progress) and `Model.watch` (stopwatch) live beside the spinner. `ProgressLayout` gains optional `Bar`/`Elapsed` string fields: set by the model's render path (animated bar, elapsed text), empty in direct construction (tests, fallback) where `View()` falls back to the static `renderProgressBar`. Palette adaptivity: the bar copy gets `FullColor`/`EmptyColor` from the live tokens at render time.

### D3: Frame and tick routing is gated globally
`progress.FrameMsg` and `stopwatch.TickMsg` are handled at the top of `Update` like spinner ticks: forwarded to the components only while a determinate progress view is active, swallowed otherwise. `syncProgressBar()` (one method, switching on the active flow's layout) is batched after each flow event; stopwatch Reset+Start cmds are batched where `progressStartedAt` is set for the four flows.

### D4: Percentage text shows actual progress
The animated cells ease; the number states the truth. The text comes from `overall()`, not the spring position, per the modified spec.

## Risks / Trade-offs

- [FrameMsgs continue after completion] → the gate swallows frames outside live progress views; completion transitions the view, ending the cycle.
- [Spring lag reads as stale at flow end] → on the final target (100%) the hold window (1.5s) exceeds the spring settle time.

## Migration Plan

Single PR `feat/animated-progress`. Playground: removal and integration progress observed mid-flight for smooth fill + elapsed. Rollback: revert merge commit.
