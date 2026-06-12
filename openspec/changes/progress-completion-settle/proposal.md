## Why

Every documented timing guarantee (minimum hold, settle at 100%, the green completion bar) is currently playground-only: real runs advance the instant the final event lands, cutting the progress view mid-spring at 10-86% fill (measured frame-exact in the June 2026 UX audit). Even with holds enabled, the settle floor is event-relative while the spring needs ~1.2s to traverse, yielding a measured 0.16s at 100%, below a comfortable eye fixation. The ending of every flow contradicts the "layout tells the truth" invariant for every real user.

## What Changes

- The completion settle becomes unconditional and state-relative: after a flow's final event, the progress view may not advance until the displayed bar fill has reached ~100% AND a settled beat (500-800ms) has elapsed since it got there.
- On the final event the spring target syncs to 100% (existing mechanism) and the success gradient applies on success (existing Wave 3a mechanism, now actually visible outside playground).
- Failure endings get the same truth-hold without success styling.
- Playground keeps its additional 1.5s entry hold on top; real runs gain no entry hold.
- `q`/`ctrl+c` still quit immediately mid-settle (existing stderr trace covers the abandoned operation).

## Capabilities

### New Capabilities

<!-- none -->

### Modified Capabilities
- `tui-progress`: gains completion-settle requirements (added; the existing spec has no timing requirements to modify).

## Impact

- Affected code: `internal/tui` (`holdOrAdvance`, `progressSettleFloor` semantics in constants.go, model.go), `main.go` (playground wiring unchanged in meaning).
- Independent of `progress-package`/`progress-declarations`; can merge first.
- Adds at most ~0.5-0.8s to fast flows in exchange for every flow ending at a visible, truthful 100%.
- Tests: condition-based (teatest WaitFor on the rendered 100% state), no sleeps.
