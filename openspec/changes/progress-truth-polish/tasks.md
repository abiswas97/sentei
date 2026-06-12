## 1. Elapsed suppression

- [x] 1.1 Render the elapsed readout only at elapsed >= 2s, keeping the fixed reserve so the bar never reflows; table-driven layout tests for hidden/visible at the boundary

## 2. Failed phase headers

- [x] 2.1 Phase headers with failures render counts without the percentage; layout tests cover failed, mixed, and all-success phases (after `progress-declarations` merges, so semantics are stable)

## 3. Failure summary voice

- [x] 3.1 Add the errors title const to `copy.go` ("Apply finished with errors", sentence case) and use it when any step failed
- [x] 3.2 Headline leads with the failed count (`✗ 2 failed, 2 applied`); update integration summary tests

## 4. Skip traces

- [x] 4.1 Skipped steps render the dim `– skipped (<reason>)` line during progress (reusing the existing dim skipped vocabulary) and on the apply summary
- [x] 4.2 E2E: an apply with detection-skipped installs shows the skip lines in both views

## 5. Verification

- [x] 5.1 Full gauntlet; regenerate affected golden views with frame review; commitlint
- [x] 5.2 Update `.impeccable.md` decision log (elapsed threshold, failed-header format, errors title, skip traces)
