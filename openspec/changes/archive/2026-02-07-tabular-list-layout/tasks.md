## 1. Model Changes

- [x] 1.1 Add `width int` field to `Model` in `internal/tui/model.go`
- [x] 1.2 Update `tea.WindowSizeMsg` handler in `updateList` to capture `msg.Width` into `m.width`

## 2. Table Rendering

- [x] 2.1 Add `import "github.com/charmbracelet/lipgloss/table"` to `internal/tui/list.go`
- [x] 2.2 Remove `colWidthBranch`, `colWidthAge`, `colWidthSubject` constants
- [x] 2.3 Rewrite `viewList()` to build a borderless `lipgloss/table` with 6 columns (cursor, checkbox, status, branch, age, subject)
- [x] 2.4 Implement `StyleFunc` closure that applies fixed widths to columns 0-2 and 4, and per-row styles (cursor/selected/normal) based on model state
- [x] 2.5 Set `.Width(m.width)` and `.Wrap(true)` on the table
- [x] 2.6 Render only the visible row slice (`m.offset` to `m.offset+m.height`) via table — build rows for the visible window only

## 3. Column Sizing and Padding

- [x] 3.1 Add column width constants in `internal/tui/styles.go` for fixed-width columns (cursor=3, checkbox=5, status=6, age=16)
- [x] 3.2 Compute proportional branch and subject widths from `m.width` (~50/50 of remaining space after fixed columns and padding)
- [x] 3.3 Add `Padding(0, 1)` to data column styles (branch, age, subject) in StyleFunc; prefix columns bake gap into fixed Width

## 4. Testing

- [x] 4.1 Update `internal/tui/list_test.go` assertions to match new column-aligned output format
- [x] 4.2 Manual test with playground (`--playground`) in wide and narrow terminals to verify column alignment and wrapping
- [x] 4.3 Manual test against real repo with long branch names to verify wrapping behavior
- [x] 4.4 Run `go vet ./...` and `go test ./...` to confirm no regressions

## 5. OpenSpec Updates

- [x] 5.1 Update design.md to reflect wrapping, padding, and proportional width decisions
- [x] 5.2 Update specs/tui-list-view/spec.md to reflect wrapping behavior
- [x] 5.3 Update tasks.md with revised task list
