# Bubbles/help Adoption: Tasks

## 1. Key presentation data (TDD)

- [x] 1.1 Tests first: per-view presentation declarations exist for every view; every footer binding appears in its view's sections (drift guard); contextual descriptions resolve per view
- [x] 1.2 keys.go: `keySection` type, `withDesc` constructor, per-view declarations (footer subset + named sections) for all views incl. conditional summary footers

## 2. Footer rendering (TDD)

- [x] 2.1 Tests first: `viewFooter` renders `  key action · key action` dim from bindings; single binding no separator; narrow width truncates with ellipsis
- [x] 2.2 chrome.go: `viewFooter` wrapping bubbles/help ShortHelpView; help styles wired into applyPalette; delete `KeyHint` + `viewKeyHints`

## 3. Render-site migration

- [x] 3.1 Swap all ~20 viewKeyHints call sites to viewFooter(view declarations)
- [x] 3.2 Route the six stray inline hint sites through viewFooter (list.go status bar, repo_options, create_options, repo_summary x2, migrate_summary); list status bar gains `?` hint
- [x] 3.3 help.go: portal sections derive from keys.go declarations; delete hand-written tables; existing help tests updated

## 4. Verification and ship

- [x] 4.1 Full gauntlet (suite + race, vet, gofmt, golangci-lint)
- [x] 4.2 Playground tmux: every view footer at 80x18, F1 portal spot checks, narrow-width (60 col) truncation check
- [x] 4.3 .impeccable.md View Chrome + Key Mapping sections updated, decision log entry
- [ ] 4.4 PR feat/bubbles-help, CI green, merge, cleanup, archive change
