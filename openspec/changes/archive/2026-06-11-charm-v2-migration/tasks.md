# Charm v2 Migration: Tasks

## 1. Toolchain and dependencies

- [x] 1.1 Create worktree `../refactor-charm-v2` on branch `refactor/charm-v2`
- [x] 1.2 Bump go.mod to Go 1.25; bump Go version in CI workflows (test matrix, lint, release)
- [x] 1.3 Swap go.mod requires to charm.land v2 modules (bubbletea, bubbles, lipgloss, x/exp/teatest); `go get` pins, `go mod tidy`
- [x] 1.4 Rewrite all import paths (61 bubbletea, 25 bubbles, 5 lipgloss incl. lipgloss/table, 3 teatest files)

## 2. Mechanical API migration (compile-error-driven)

- [x] 2.1 `Model.View()` returns `tea.View` with AltScreen + MouseMode + keyboard-enhancements declared; remove `tea.WithAltScreen()` from both `tea.NewProgram` sites in main.go
- [x] 2.2 Replace every `case tea.KeyMsg:` with `tea.KeyPressMsg` across views and tests; rewrite test key-event literals to v2 constructors
- [x] 2.3 Change Toggle binding `" "` → `"space"` in keys.go; sweep for any other `" "` key matches
- [x] 2.4 Migrate portal.go viewport to v2 constructor options and getters/setters
- [x] 2.5 Migrate textinput construction in model.go: v2 options, `SetVirtualCursor(true)`, `DefaultDarkStyles()`, drop `Cursor.BlinkCmd()` calls in create_branch.go/clone_input.go/repo_name.go
- [x] 2.6 Migrate teatest E2E files to teatest v2 API; all WaitFor matchers stay condition-based
- [x] 2.7 Full suite green: `go test ./...`, `go vet ./...`, gofmt, golangci-lint

## 3. New behavior (spec deltas)

- [x] 3.1 List view: `tea.MouseWheelMsg` maps wheel up/down to existing cursor up/down paths, with table-driven tests including top/bottom boundary scenarios
- [x] 3.2 Portal: verify v2 viewport wheel scrolling works while open and does not reach the background view (test per tui-portal delta)
- [x] 3.3 Integration carousel: wheel maps to existing left/right navigation, with test
- [x] 3.4 Quick-create E2E: ctrl+enter on create-branch input starts creation with default options (kitty-protocol key event via teatest)

## 4. Verification and ship

- [x] 4.1 Playground pass (tmux subagent, keys sent individually) at 80x18 and normal size: every view eyeballed against .impeccable.md; explicit portal-over-list composite check
- [x] 4.2 Update `.impeccable.md` decision log (v2 platform, virtual-cursor choice, wheel-as-alias); verify no spec/code drift introduced
- [x] 4.3 PR with refactor-type commits; CI green (build, lint, both OS test jobs, commitlint, codecov/patch); merge and clean up worktree
