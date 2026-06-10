## 1. Overlay Compositing (no new dependency)

- [x] 1.1 Write failing tests for `compositeOverlay` in `internal/tui/overlay_test.go` — centered placement, background visible at edges, ANSI-colored background lines survive splicing, background shorter than overlay
- [x] 1.2 Create `internal/tui/overlay.go` with an ANSI-aware centered compositing helper on `x/ansi` (`Truncate`, `TruncateLeft`, `StringWidth`); `go mod tidy` flips `x/ansi` to direct

## 2. DetailPortal Component

- [x] 2.1 Write failing tests for `DetailPortal` in `internal/tui/portal_test.go` — open/close lifecycle, scroll position resets on open, key interception (scroll keys work, navigation keys blocked, quit passes through)
- [x] 2.2 Create `internal/tui/portal.go` with `DetailPortal` struct (title, content, visible, viewport), `Open(title, content)`, `Close()`, `Update(msg)`, `View(background)` methods to pass tests
- [x] 2.3 Write tests for portal sizing — standard terminal (80x24), resize while open, margin calculation
- [x] 2.4 Implement portal sizing logic with terminal-relative margins
- [x] 2.5 Write tests for portal chrome — title bar, scroll indicator visible/hidden, dismiss hint
- [x] 2.6 Implement portal chrome rendering using `viewTitle` and `viewKeyHints` from Change 1

## 3. Model Integration

- [x] 3.1 Add `portal DetailPortal` field to `Model` in `model.go`
- [x] 3.2 Wire portal into `Model.Update` — when `portal.Visible`, delegate keys to portal first (except quit)
- [x] 3.3 Wire portal into `Model.View` — when `portal.Visible`, composite portal over current view using `bubbletea-overlay`
- [x] 3.4 Write tests for portal key delegation — portal intercepts when visible, passes through when hidden

## 4. Help Overlay

- [x] 4.1 Define `helpContent() string` method interface and implement for list view, menu view, progress views, summary views, confirmation views
- [x] 4.2 Write tests for help content generation — each view returns non-empty, formatted content with key bindings
- [x] 4.3 Wire `F1` key in `Model.Update` — opens portal with `helpContent()` from current view, toggles closed if already open
- [x] 4.4 Write E2E test: press F1 on list view → help portal visible → press Esc → portal closes

## 5. Contextual Details Key

- [x] 5.1 Wire `?` key in `Model.Update` — calls view-specific detail provider, opens portal if content available, toggles closed if already open, no-op if no details
- [x] 5.2 Write tests for `?` key — opens when details available, no-op when not, toggles closed
- [x] 5.3 Ensure `?` and `F1` don't conflict — if help is open and `?` is pressed, close help and open details (or vice versa)

## 6. Design System Documentation

- [x] 6.1 Add portal/overlay section to `.impeccable.md` Component Patterns — sizing, chrome, key behavior, when to use portal vs inline content

## 7. Verification

- [x] 7.1 Run `go fmt ./...` and `go vet ./...`
- [x] 7.2 Run `go test ./...` — all tests pass
- [x] 7.3 Run `go build` — binary builds
- [x] 7.4 Manual test: open help from multiple views, verify contextual content, scroll, dismiss
- [x] 7.5 Session meta doc removed earlier (stale handoff docs); completion status lives in this tasks file

## 8. Post-review polish (from playground visual review)

- [x] 8.1 Portal box hugs short content instead of filling the terminal
- [x] 8.2 One-space clear margin around the box (no background glyphs touching the border)
- [x] 8.3 Scroll indicator reads `↓ more` instead of a position percentage
- [x] 8.4 Worktree detail values truncate with `…` to the portal width
- [x] 8.5 Global help entries dedupe against view sections (no doubled quit row)
