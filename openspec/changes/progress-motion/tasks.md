# Tasks: progress-motion

## 1. Implementation

- [x] 1.1 Breathing dot: breath spinner on the model, `ActiveGlyph` injection through `ProgressLayout`, all four `◐` sites converted, ticks gated to live progress, tests (TDD)
- [x] 1.2 Responsive bar: width = content minus meta with floor, WindowSizeMsg drives the animated model, static fallback takes width, tests
- [x] 1.3 Gradient fill: `barStart`/`barEnd` tokens in both palettes, `WithColors` + `WithScaled` on the overall bar, tests
- [x] 1.4 Docs: `.impeccable.md` status indicators + decision log entry
- [x] 1.5 Gauntlet; VHS GIF verification of remove/create/cleanup flows at 80 and 120 columns; PR, CI, merge (#73)

## 2. Smoothing revision (user GIF review: staccato breath, snappy bar, two vocabularies)

- [x] 2.1 One working spinner: heavy braille dot (single-cell frames, 10fps) replaces both the breath spinner and MiniDot; one instance, one visibility gate, single tick chain per entry path (Init + dispatch wrapper, explicit starts removed), tests (TDD)
- [x] 2.2 Silky bar: spring frequency lowered so fills glide rather than snap, still settling within the hold; tests unchanged (behavioral contract)
- [x] 2.3 Docs: status indicator table, timing section, decision log entry recording the revision and the instant-cuts transition principle
- [x] 2.4 Gauntlet; VHS re-record incl. slow-deletion repo; PR, CI, merge
