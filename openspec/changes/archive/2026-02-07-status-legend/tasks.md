## 1. Legend Rendering

- [x] 1.1 Add `viewLegend()` method to Model in `internal/tui/list.go` that returns a styled legend string: `[ok] clean  [~] dirty  [!] untracked  [L] locked` using existing `styleStatus*` styles for indicators and `styleDim` for labels
- [x] 1.2 Call `viewLegend()` from `viewList()` after `viewStatusBar()`, appending the result with a newline

## 2. Viewport Height Fix

- [x] 2.1 Change height calculation in `updateList` WindowSizeMsg handler from `msg.Height-4` to `msg.Height-5` to account for the extra legend line

## 3. Verification

- [x] 3.1 Run `go build` and test visually with `--playground` flag to confirm legend renders correctly, colors match, and table rows don't overflow
