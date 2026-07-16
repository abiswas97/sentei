# Native Text Input Paste Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make native terminal paste and the existing widget clipboard command edit every focused Sentei text input safely and consistently.

**Architecture:** Keep message ownership in each editable view. View-specific keybindings are consumed first; every remaining message is forwarded to the focused Bubbles `textinput`, preserving the widget's public bracketed-paste and private asynchronous clipboard messages. View helpers retain filter and derived-field side effects, while the root model prevents a visible portal from leaking native paste into its hidden background.

**Tech Stack:** Go 1.25, Bubble Tea v2, Bubbles v2 `textinput`, existing TUI unit and model tests.

---

### Task 1: Route paste through worktree, repository, and clone inputs

**Files:**
- Modify: `internal/tui/create_branch.go`
- Modify: `internal/tui/repo_name.go`
- Modify: `internal/tui/clone_input.go`
- Test: `internal/tui/create_branch_test.go`
- Test: `internal/tui/repo_name_test.go`
- Test: `internal/tui/clone_input_test.go`

- [ ] **Step 1: Add failing focused-field and derived-value tests**

Add `tea.PasteMsg` cases for both focused fields on each screen. Require branch/base isolation, repository-name-to-location derivation, direct location editing, clone-URL-to-name derivation, and clone-name manual override. Seed `validationErr` before paste and require it to clear. Include `tea.PasteMsg{Content: "a\n\tb\x1b界"}` and expect the Bubbles sanitizer result `"a  b界"`.

- [ ] **Step 2: Run focused tests and verify RED**

```bash
go test ./internal/tui -run 'TestUpdate(CreateBranch|RepoName|CloneInput)_Paste' -count=1
```

Expected: FAIL because each update function ignores `tea.PasteMsg`.

- [ ] **Step 3: Extract per-view edit helpers and forward unconsumed messages**

In each current update function, replace the outer type switch with `if keyMsg, ok := msg.(tea.KeyPressMsg); ok`, keep its Back, Tab, Confirm, and QuickCreate branches unchanged, and use `keyMsg` for all `key.Matches` calls. After that guarded block, call a private edit helper accepting the original `tea.Msg`. The create helper clears `validationErr`, updates `branchInput` when `focusedField == 0` and `baseInput` otherwise, then returns the updated model and command.

Apply the same pattern in repository and clone inputs, moving their existing derivation logic into `updateRepoNameInput` and `updateCloneTextInput`. Forward every unconsumed message; do not type-switch on the widget's private clipboard response.

- [ ] **Step 4: Run focused and package tests**

```bash
go test ./internal/tui -run 'TestUpdate(CreateBranch|RepoName|CloneInput)_' -count=1
go test ./internal/tui -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit the input-flow slice**

```bash
git add internal/tui/create_branch.go internal/tui/create_branch_test.go \
  internal/tui/repo_name.go internal/tui/repo_name_test.go \
  internal/tui/clone_input.go internal/tui/clone_input_test.go
git commit -m "feat(tui): support paste in form inputs"
```

### Task 2: Cover filter, optional description, and portal isolation

**Files:**
- Modify: `internal/tui/list.go`
- Modify: `internal/tui/repo_options.go`
- Modify: `internal/tui/model.go`
- Test: `internal/tui/remove_filter_test.go`
- Test: `internal/tui/repo_options_test.go`
- Test: `internal/tui/portal_test.go`

- [ ] **Step 1: Add failing tests for remaining editable surfaces**

Dispatch `tea.PasteMsg` through `updateList` with an active filter and require immediate reindexing. Paste into a visible, focused repository description and require the value to change. Require paste to be ignored when that description is hidden or unfocused. Open F1 help over a create-branch model, send paste through `Model.Update`, and assert the portal remains visible while the branch input stays unchanged.

- [ ] **Step 2: Run focused tests and verify RED**

```bash
go test ./internal/tui -run 'Test(UpdateFilterInput|UpdateRepoOptions|Portal).*Paste' -count=1
```

Expected: FAIL because filter and description accept only keypress messages and portal paste reaches the background dispatcher.

- [ ] **Step 3: Forward non-key messages only while an input is active**

Change `updateFilterInput` to accept `tea.Msg`; inspect Escape and Enter only for `tea.KeyPressMsg`, then update the input, copy its value to `filterText`, and reindex. In `updateList`, route non-key messages to the filter only while `filterActive`.

In repository options, preserve `ghAuthStatusMsg` and navigation. When publish is enabled and `optionsCursor == repoOptDescription`, route every non-navigation/action message to `descInput.Update(msg)`; ignore paste when hidden or unfocused.

In the visible-portal branch of `Model.Update`, consume native paste before background dispatch:

```go
if m.portal.Visible() {
    if _, ok := msg.(tea.PasteMsg); ok {
        return m, nil
    }
    // Existing key, wheel, and background-event handling remains.
}
```

- [ ] **Step 4: Run focused, package, and race tests**

```bash
go test ./internal/tui -run 'Test(UpdateFilterInput|UpdateRepoOptions|Portal).*Paste' -count=1
go test ./internal/tui -count=1
go test -race ./internal/tui -count=1
```

Expected: PASS with no races.

- [ ] **Step 5: Commit the remaining-surface slice**

```bash
git add internal/tui/list.go internal/tui/remove_filter_test.go \
  internal/tui/repo_options.go internal/tui/repo_options_test.go \
  internal/tui/model.go internal/tui/portal_test.go
git commit -m "fix(tui): isolate paste to visible text inputs"
```

### Task 3: Full-model regression, verification, and PR delivery

**Files:**
- Create: `internal/tui/paste_test.go`
- Modify only when a failing regression proves it necessary: files from Tasks 1-2

- [ ] **Step 1: Add the full-model regression matrix**

Send `tea.PasteMsg` through `Model.Update` for create branch, repository name, clone URL, active filter, and description. Assert the focused value and domain side effect. Include cursor-middle insertion and `"q/enter"` to prove paste edits text without changing views.

Add a structural `Ctrl+V` test without executing the OS clipboard command:

```go
updated, cmd := m.Update(tea.KeyPressMsg{Code: 'v', Mod: tea.ModCtrl})
if cmd == nil {
    t.Fatal("focused text input must return its clipboard command")
}
if updated.(Model).view != createBranchView {
    t.Fatal("ctrl+v must not navigate")
}
```

- [ ] **Step 2: Run the full verification gauntlet**

```bash
go test ./internal/tui -run 'TestModel_Paste|TestModel_CtrlVPasteCommand' -count=1
GOCACHE=/tmp/sentei-paste-go-cache go test -race ./...
GOCACHE=/tmp/sentei-paste-go-cache go vet ./...
GOCACHE=/tmp/sentei-paste-go-cache GOLANGCI_LINT_CACHE=/tmp/sentei-paste-lint-cache golangci-lint run ./...
GOCACHE=/tmp/sentei-paste-go-cache go build ./...
git diff --check origin/main...HEAD
```

Expected: every command exits 0 and lint prints `0 issues.`

- [ ] **Step 3: Commit the full-model regression**

```bash
git add internal/tui/paste_test.go
git commit -m "test(tui): cover native paste routing"
```

- [ ] **Step 4: Review the final branch diff**

Review `origin/main...HEAD` for only approved paste behavior. Reject unrelated refactors, new dependencies, public API changes, paste hints, or clipboard mutation in tests. Re-run focused tests after any review fix.

- [ ] **Step 5: Push and open the PR**

```bash
git push -u origin codex/paste-text-inputs
gh pr create --base main --head codex/paste-text-inputs \
  --title "feat(tui): support native paste in text inputs" \
  --body-file /tmp/sentei-paste-pr-body.md
```

The PR body must summarize all editable surfaces, portal isolation, normalization behavior, verification evidence, and state that no dependency or public API changed. Open it ready for review and do not merge it.
