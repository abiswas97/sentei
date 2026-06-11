# Proposal: golden-views

## Why

Wave 3c: the chrome's exact output (spacing, vocabulary, styling) is guarded only by substring assertions; whole-frame regressions slip through between VHS reviews.

## What Changes

- Golden-file tests pin the five stable views (list, confirm, summary, cleanup result, create input), ANSI included, via `x/exp/golden`. Regeneration is intentional: `go test ./internal/tui/ -run TestGolden -update`.

## Capabilities

### Modified

- `tui-design-system`: chrome pinning rule.

## Impact

- internal/tui/golden_test.go + testdata/*.golden; no behavior change.
