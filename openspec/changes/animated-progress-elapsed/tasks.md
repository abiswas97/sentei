# Animated Progress: Tasks

## 1. Implementation (TDD)

- [x] 1.1 Tests first: ProgressLayout.overall() extraction; static fallback when Bar unset; FrameMsg/TickMsg gating (forwarded in progress views, swallowed elsewhere); syncProgressBar sets the spring target from the active flow
- [x] 1.2 ProgressLayout: overall() method, Bar/Elapsed fields with static fallback
- [x] 1.3 Per-flow layout methods extracted (removal, create, repo/migrate, integration); views render them
- [x] 1.4 Model: bar + watch components, global routing, syncProgressBar, stopwatch reset/start at the four flow starts, animated-bar render path with palette colors

## 2. Verification and ship

- [x] 2.1 Full gauntlet; playground: removal progress animates smoothly with elapsed readout; no animation residue after completion
- [x] 2.2 .impeccable.md Progress Views section + decision log (incl. fang declined)
- [ ] 2.3 PR feat/animated-progress, CI green, merge, cleanup, archive
