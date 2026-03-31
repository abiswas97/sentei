## 1. Dependency Setup

- [ ] 1.1 Add `github.com/rmhubbert/bubbletea-overlay` to `go.mod` via `go get`
- [ ] 1.2 Verify dependency resolves and `go build` succeeds

## 2. DetailPortal Component

- [ ] 2.1 Write failing tests for `DetailPortal` in `internal/tui/portal_test.go` — open/close lifecycle, scroll position resets on open, key interception (scroll keys work, navigation keys blocked, quit passes through)
- [ ] 2.2 Create `internal/tui/portal.go` with `DetailPortal` struct (title, content, visible, viewport), `Open(title, content)`, `Close()`, `Update(msg)`, `View(background)` methods to pass tests
- [ ] 2.3 Write tests for portal sizing — standard terminal (80x24), resize while open, margin calculation
- [ ] 2.4 Implement portal sizing logic with terminal-relative margins
- [ ] 2.5 Write tests for portal chrome — title bar, scroll indicator visible/hidden, dismiss hint
- [ ] 2.6 Implement portal chrome rendering using `viewTitle` and `viewKeyHints` from Change 1

## 3. Model Integration

- [ ] 3.1 Add `portal DetailPortal` field to `Model` in `model.go`
- [ ] 3.2 Wire portal into `Model.Update` — when `portal.Visible`, delegate keys to portal first (except quit)
- [ ] 3.3 Wire portal into `Model.View` — when `portal.Visible`, composite portal over current view using `bubbletea-overlay`
- [ ] 3.4 Write tests for portal key delegation — portal intercepts when visible, passes through when hidden

## 4. Help Overlay

- [ ] 4.1 Define `helpContent() string` method interface and implement for list view, menu view, progress views, summary views, confirmation views
- [ ] 4.2 Write tests for help content generation — each view returns non-empty, formatted content with key bindings
- [ ] 4.3 Wire `F1` key in `Model.Update` — opens portal with `helpContent()` from current view, toggles closed if already open
- [ ] 4.4 Write E2E test: press F1 on list view → help portal visible → press Esc → portal closes

## 5. Contextual Details Key

- [ ] 5.1 Wire `?` key in `Model.Update` — calls view-specific detail provider, opens portal if content available, toggles closed if already open, no-op if no details
- [ ] 5.2 Write tests for `?` key — opens when details available, no-op when not, toggles closed
- [ ] 5.3 Ensure `?` and `F1` don't conflict — if help is open and `?` is pressed, close help and open details (or vice versa)

## 6. Design System Documentation

- [ ] 6.1 Add portal/overlay section to `.impeccable.md` Component Patterns — sizing, chrome, key behavior, when to use portal vs inline content

## 7. Verification

- [ ] 7.1 Run `go fmt ./...` and `go vet ./...`
- [ ] 7.2 Run `go test ./...` — all tests pass
- [ ] 7.3 Run `go build` — binary builds
- [ ] 7.4 Manual test: open help from multiple views, verify contextual content, scroll, dismiss
- [ ] 7.5 Update session meta doc with completion status
