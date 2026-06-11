# Tab Presence: Design

## Decisions

### D1: One operation table, two voices
`flowIdentity` maps progress views to a sentence form ("worktree removal", for the quit trace) and a tab verb ("removing"). `InterruptedFlow` and `windowTitle` both read it; the knowledge exists once.

### D2: The tab mirrors the bar's truth
`terminalProgress` reuses `activeProgressLayout().overall()` (the same source as the spring target): indeterminate while `indeterminateWaitActive`, error state if any phase reports failures, value otherwise, nil when no flow is live. No new progress math.
