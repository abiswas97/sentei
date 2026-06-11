# Charm v2 Migration: Design

## Context

Sentei is built on the Charm v1 generation: bubbletea 1.3.10, bubbles 1.0.0, lipgloss 1.1.0, teatest (v1), Go 1.24.2. Charm's stable v2 lives under the `charm.land` vanity module path and reshapes three load-bearing APIs: `View()` returns a declarative `tea.View` struct, key events split into `KeyPressMsg`/`KeyReleaseMsg`, and program options (alt screen, mouse) move into View fields. 61 files import bubbletea, 25 bubbles, 5 lipgloss, 3 teatest. The root `Model` (internal/tui/model.go) is the only `tea.Model`; `ProgressLayout.View()` and `ConfirmationViewModel.View()` are plain render helpers returning strings.

## Goals / Non-Goals

**Goals**
- Behavior-identical port of all existing functionality to the v2 stack.
- Enable v2 keyboard enhancements: `ctrl+enter` quick-create becomes functional on kitty-protocol terminals (it is currently unreachable on most terminals). Gains its first E2E test.
- Enable mouse wheel scrolling in the worktree list and the portal viewport.
- Go 1.25 toolchain in go.mod and CI (required by bubbletea v2).

**Non-Goals**
- No adaptive theming (next change; v2 unblocks it via `tea.RequestBackgroundColor`).
- No bubbles/help, spinner, or visual redesign of any view.
- No mouse click/motion handling; wheel only.
- No real-cursor adoption for text inputs (see Decision 4).

## Decisions

### D1: All four modules move to charm.land paths in one change
`charm.land/bubbletea/v2` (pin latest v2.0.x), `charm.land/bubbles/v2` (v2.1.x), `charm.land/lipgloss/v2` + `lipgloss/v2/table`, `charm.land/x/exp/teatest/v2`. Mixing generations (e.g. lipgloss v1 with bubbletea v2) is unsupported upstream because v2 made lipgloss pure (Bubble Tea owns I/O, colorprofile downsamples). Alternative considered: incremental per-module migration; rejected, the modules are coupled at the renderer boundary. If the `charm.land` teatest path does not resolve, fall back to `github.com/charmbracelet/x/exp/teatest/v2` (same code, older path).

### D2: Only the root Model adopts tea.View; render helpers keep returning string
`Model.View() tea.View` wraps the composed frame via `tea.NewView(content)` and sets `AltScreen = true`, `MouseMode = tea.MouseModeCellMotion`, and the keyboard-enhancements request. `ProgressLayout.View()` and `ConfirmationViewModel.View()` stay `string`-returning helpers (they are not `tea.Model`s). The two `tea.WithAltScreen()` program options in main.go are deleted. Rationale: View fields are program-global state; declaring them once at the root matches the v2 "declare what you want" model and keeps sub-views pure string renderers (existing testing strategy intact).

### D3: Key handling becomes KeyPressMsg; the space binding moves to "space"
Every `case tea.KeyMsg:` becomes `case tea.KeyPressMsg:` (no release handling anywhere). `keys.go` Toggle changes `key.WithKeys(" ")` → `key.WithKeys("space")` because v2 `String()` reports `"space"`. Test key-event literals are rewritten to v2 constructors. No other binding strings change.

### D4: Text inputs keep the virtual cursor
bubbles v2 textinput defaults to the real terminal cursor (`Cursor()` returns `*tea.Cursor`, positioned via `tea.View.Cursor`), and the v1 `input.Cursor.BlinkCmd()` call in create_branch.go disappears. For a behavior-identical migration we opt into `SetVirtualCursor(true)` and drop the blink commands, preserving v1 rendering exactly. Alternative considered: adopt the real cursor (more canonical, true terminal cursor blink); rejected for this change because it requires plumbing cursor positions through the view composition and re-verifying every input view; revisit alongside the huh-forms evaluation. textinput styles use `DefaultDarkStyles()` (v2 requires explicit dark/light selection) until the adaptive-theming change makes the choice dynamic.

### D5: Wheel events map to existing navigation, not new behavior
`tea.MouseWheelMsg` in the list view maps wheel-up/down to the existing cursor-up/down paths (same code the `j`/`k` keys hit). The portal viewport gets wheel support natively from bubbles v2 viewport. The integration carousel maps wheel to its existing left/right navigation. No other view reacts to the mouse. Rationale: wheel is an input alias, not a feature; aliasing keeps scenarios and tests shared with keyboard paths.

### D6: viewport/textinput constructor migration is localized
bubbles v2 moved viewport to functional options (`viewport.New(viewport.WithWidth(w), ...)`) and field access to getters/setters. Only portal.go constructs a viewport; only model.go constructs textinputs. The changes stay inside those files; no signature changes leak to callers.

## Risks / Trade-offs

- [Renderer rewrite changes edge-case output (wide glyphs, ANSI framing)] → existing teatest E2E + unit render tests assert content; manual playground pass at 80x18 and normal size before merge.
- [`compositeOverlay` (overlay.go) parses ANSI from lipgloss output via x/ansi; v2 output could shift] → portal teatest coverage exercises the composite path; x/ansi is bumped in lockstep; explicit manual check of portal-over-list rendering.
- [Keyboard enhancements degrade on non-kitty terminals] → degradation is to current behavior (ctrl+enter indistinguishable); enter-based flow remains primary. Help/hints unchanged.
- [teatest v2 output framing differs (sync sequences), breaking WaitFor matchers] → matchers assert on plain substrings; fix per-failure, never with sleeps (condition-based waits only, per testing standards).
- [charm.land module proxy availability in CI] → modules are proxied through proxy.golang.org like any path; no special CI config expected.

## Migration Plan

1. New worktree/branch `refactor/charm-v2`.
2. Swap module paths (go.mod + all imports), `go get` v2 pins, `go mod tidy`, Go 1.25 in go.mod and CI workflows.
3. Compile-error-driven mechanical pass: View struct, KeyPressMsg, space binding, viewport/textinput constructors, teatest v2.
4. Add wheel handling (D5) and the keyboard-enhancements request + quick-create E2E.
5. Full gauntlet (fmt, vet, golangci-lint, tests both OSes via CI) + playground verification.
6. Single PR; merge commit to main. Ships as `refactor` type, riding into the adaptive-theming release.

**Rollback**: revert the merge commit; no data, config, or persisted-state formats are touched.
