# Progress Endings: Design

## Decisions

### D1: Completion is layout data
Each flow's layout method sets `Completed` from its existing result state (`m.create.result != nil` etc.). `overall()` with `Completed` treats undiscovered phases as zero outstanding; `renderPhase` renders them `– <Name>  skipped` in dim. No new model state.

### D2: Final sync mirrors the removal fix
`syncProgressBar()` batched at each flow's completion message, after result assignment, before `holdOrAdvance`, so the spring's last target is 1.0 and settles during the hold.

### D3: Interruption leaves a trace
`Model.InterruptedFlow() string` names the live operation when the final model is still in a progress view; both `main.go` program sites log a warning after `Run` returns. Quitting stays instant; the trace costs one stderr line.

### D4 (deferred): fixed denominators for repo flows
Requires the pipeline to emit its plan upfront (a StepPending event type). Recorded as follow-up; out of scope here.
