# Native Paste for Text Inputs

## Goal

Make native terminal paste work in every editable Sentei field. Bubble Tea v2 emits bracketed terminal paste as `tea.PasteMsg`; Sentei currently forwards only `tea.KeyPressMsg` to its `textinput` models, so terminal paste is silently ignored even though the widgets already support it.

Success means macOS `Cmd+V`, terminal paste commands such as `Ctrl+Shift+V`, and the existing text-input `Ctrl+V` clipboard command all edit the focused field without submitting the form or triggering shortcuts.

## Behavior

- Cover all current single-line inputs: worktree branch and base branch, worktree filter, repository name and location, optional repository description, clone URL, and clone name.
- Insert pasted text at the current cursor position. Do not replace the whole field, move focus, submit, navigate, or interpret pasted content as keybindings.
- Use the `textinput` sanitizer consistently for native and widget clipboard paste: newlines and tabs become spaces, other control characters are removed, and Unicode plus ordinary spaces are preserved.
- Preserve each screen's existing edit consequences: filter results reindex immediately; repository-name changes update the derived location; clone-URL changes update the derived name until the user manually edits that name; editing clone name marks it manually edited; validation errors clear as they do for typing.
- Deliver paste only to the focused, visible input. Ignore it on non-editable views. If a help or details portal is open, consume paste without changing the hidden form underneath.
- Do not add a paste footer hint: the shortcut is terminal/platform-owned, and the existing `Ctrl+V` widget binding remains available.

## Design

Keep input ownership in each view instead of adding a global field registry or form framework.

- Refactor each editable view's existing post-keypress input block into a small private edit helper that accepts `tea.Msg`. After consuming view-specific messages and keybindings, pass every unconsumed message to that helper. This is the standard Bubbles update pattern and supports both public `tea.PasteMsg` and the widget's private asynchronous `Ctrl+V` response while keeping typing and paste side effects identical.
- Widen the filter updater from `tea.KeyPressMsg` to `tea.Msg`; retain Escape/Enter handling only for key messages and pass every other message to the input followed by reindexing.
- In `Model.Update`, swallow paste while a portal is visible before per-view dispatch. Outside portals, let the existing view dispatcher determine whether the current view is editable.
- Do not synthesize a keypress from pasted content. Keeping `tea.PasteMsg` distinct is what prevents values such as `q`, `/`, or `enter` from becoming commands.

No public API, configuration, persisted state, dependency, or CLI behavior changes.

## Verification

- Unit-test paste into both focused fields on multi-field screens and confirm the unfocused field is unchanged.
- Verify branch/base input, filter reindexing, repository name/location derivation, description editing, clone URL derivation, and clone-name manual override.
- Cover cursor-middle insertion, multiline/tab normalization, Unicode preservation, control-character removal, and shortcut-like pasted text.
- Verify a focused field returns a non-nil command for `Ctrl+V`; do not read or overwrite the developer's real OS clipboard in tests. The upstream widget suite owns its private clipboard-response behavior, while Sentei's helpers guarantee unconsumed messages reach that widget.
- Verify paste is ignored on non-editable views and while help/details portals cover an editable view.
- Add a full-model test that dispatches `tea.PasteMsg` through `Model.Update`, not only directly to view helpers.
- Run focused TUI tests, `go test -race ./...`, `go vet ./...`, `golangci-lint run ./...`, `go build ./...`, and `git diff --check origin/main...HEAD`.

## Delivery

Implement on `codex/paste-text-inputs` in the dedicated `paste-text-inputs` worktree. Commit the feature and tests, push the branch, open a non-draft PR against `main`, and do not merge it.
