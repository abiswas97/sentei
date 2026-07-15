## Context

`holdOrAdvance` (model.go) transitions instantly when `minProgressDuration == 0`, which is every non-playground run; `progressSettleFloor` only applies when holds are enabled and is measured from the final event, not from the bar reaching full. The spring (frequency 6) needs ~1.2s to traverse, so even the playground's settle delivers 4 frames at 100% (measured). User decision: short settle only — no entry hold for real runs; endings always tell the truth.

## Goals / Non-Goals

**Goals**: every flow's last progress frame shows a settled 100% bar (green on success); fast real flows stay fast (sub-second added latency); playground demos keep their legibility hold.

**Non-Goals**: entry holds for real runs; changing spring physics; changing what 100% means (that is `progress-declarations`).

## Decisions

**D1: Settle is a state predicate, not a timer re-tune.**
The advance condition becomes: final event received AND displayed fill >= threshold (~99.5%) AND `settledBeat` (500-800ms, one constant) elapsed since the threshold was first met. This is immune to spring-frequency changes; re-tuning the existing event-relative floor to ~1.7s was rejected because it re-breaks the moment anyone touches the spring.

**D2: One mechanism for both modes.**
Playground = settle + entry hold; real = settle only. `progressSettleFloor`'s event-relative semantics are replaced, not special-cased; the constant is renamed to match its new meaning so no reader trusts the old doc comment.

**D3: Failure endings settle too.**
The hold-until-truth applies regardless of outcome; only the gradient differs. A failure that cuts away mid-bar is the same lie as a success that does.

**D4: Spring sync on final event stays the existing mechanism.**
No jump-cut to 100%: the spring glides (~1.2s) and the settle beat starts when the displayed fill arrives. Worst case added wait ~1.2s glide + 0.5s beat on a flow that completed instantly; acceptable against the strobe it replaces, and the glide shortens as `progress-declarations` makes mid-flow fill more truthful (less distance to travel at the end).

## Risks / Trade-offs

- [Spring never reaches threshold due to float asymptote] → threshold at 99.5% with a hard timeout fallback (2x expected glide) so the view can never wedge; invariant-tested.
- [Added latency annoys the 10x/day user] → measured cost <=1.7s worst case, typically <1s; the audit's F6 quantified the alternative (sub-perceptible strobing) as worse. Revisit with real usage if it grates.
- [teatest flakiness around animation frames] → assertions poll rendered output via WaitFor (no sleeps, per project rule); the threshold state is deterministic given pumped frames.

## Migration Plan

Single PR. Re-record the real-repo motion tape afterward and frame-verify: last progress frame of each flow shows 100% (green on success). Rollback: revert; behavior returns to instant-advance.

## Open Questions

- Exact beat duration (500 vs 800ms): pick in implementation by re-recording the real-repo tape at both and frame-reviewing; not worth a lab.
