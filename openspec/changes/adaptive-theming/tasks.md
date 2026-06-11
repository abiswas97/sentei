# Adaptive Theming: Tasks

## 1. Palette restructure (TDD)

- [x] 1.1 Tests first: palette switching (`applyPalette(lightPalette)` changes token values and rebuilt styles; switching back restores dark exactly), with a defer restoring dark
- [x] 1.2 styles.go: `palette` struct, `darkPalette`/`lightPalette` data values (approved mapping), `applyPalette` constructor rebuilding all styleX vars, init applies dark

## 2. Detection wiring (TDD)

- [x] 2.1 Tests first: `Update` with a light `tea.BackgroundColorMsg` applies the light palette; a dark one is a no-op; `Init` includes the background request cmd
- [x] 2.2 model.go: `Init` batches `tea.RequestBackgroundColor`; top-of-Update handler applies the palette by `IsDark()`

## 3. Verification and docs

- [x] 3.1 Throwaway in-package test t.Logs light-palette renders of menu, list, and portal for eyeball inspection, then is deleted
- [x] 3.2 Playground tmux pass on dark terminal at 80x18: no regression vs current rendering
- [x] 3.3 `.impeccable.md`: palette table gains the Light column, theme line updated from aspiration to fact, decision log entry
- [ ] 3.4 Full gauntlet, PR `feat/adaptive-theming`, CI green, merge, cleanup, archive change
