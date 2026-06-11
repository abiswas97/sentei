# Spinner: Tasks

## 1. Implementation (TDD)

- [x] 1.1 Tests first: tick routing (tick while scanning advances frame + returns cmd; tick when idle returns no cmd); scanning line renders spinner frame + text; menu loading item renders spinner + loading…
- [x] 1.2 Model: spinner.Model field (MiniDot, accent via applyPalette), indeterminateWaitActive gate, TickMsg handling, tick start at scan entry and menu load kick
- [x] 1.3 menuItem.loading flag replaces the static hint string; viewMenu renders spinner for loading items
- [x] 1.4 cleanup_preview scanning line uses the spinner frame

## 2. Verification and ship

- [x] 2.1 Full gauntlet; playground: cleanup scan animates during hold, menu loading state, no stray animation elsewhere
- [x] 2.2 .impeccable.md Timing section + decision log updated
- [ ] 2.3 PR feat/spinner, CI green, merge, cleanup, archive
