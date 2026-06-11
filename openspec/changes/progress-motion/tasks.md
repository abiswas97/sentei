# Tasks: progress-motion

## 1. Implementation

- [x] 1.1 Breathing dot: breath spinner on the model, `ActiveGlyph` injection through `ProgressLayout`, all four `◐` sites converted, ticks gated to live progress, tests (TDD)
- [x] 1.2 Responsive bar: width = content minus meta with floor, WindowSizeMsg drives the animated model, static fallback takes width, tests
- [x] 1.3 Gradient fill: `barStart`/`barEnd` tokens in both palettes, `WithColors` + `WithScaled` on the overall bar, tests
- [x] 1.4 Docs: `.impeccable.md` status indicators + decision log entry
- [ ] 1.5 Gauntlet; VHS GIF verification of remove/create/cleanup flows at 80 and 120 columns; PR, CI, merge
