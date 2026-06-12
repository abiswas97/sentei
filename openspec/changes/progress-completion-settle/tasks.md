## 1. Settle mechanism

- [x] 1.1 Replace `progressSettleFloor`'s event-relative semantics with the state-relative advance predicate in `holdOrAdvance`: final event AND displayed fill >= 99.5% AND settled beat elapsed since threshold; rename the constant to match the new meaning
- [x] 1.2 Apply unconditionally (real and playground); playground keeps the entry hold on top; add the hard timeout fallback (2x expected glide) with an invariant test that the view cannot wedge
- [x] 1.3 Verify quit-during-settle exits immediately with the stderr operation trace (existing test or add one)

## 2. Tests

- [x] 2.1 teatest: real-mode removal's last progress frame before summary renders 100% + success gradient (WaitFor-based, no sleeps)
- [x] 2.2 teatest: failure outcome holds to threshold without success gradient
- [x] 2.3 Table-driven unit tests for the advance predicate (event-not-final, fill-below-threshold, beat-not-elapsed, timeout fallback)

## 3. Verification

- [x] 3.1 Pick the beat duration (500 vs 800ms) by re-recording the real-repo motion tape at both and frame-reviewing the ending
- [x] 3.2 Full gauntlet; golden tests unaffected (settle changes timing, not stable-view rendering)
- [x] 3.3 Update `.impeccable.md` Timing section: guarantees now hold for all runs, settle is state-relative
