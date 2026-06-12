# Proposal: persistent-input-fields

## Why

The huh spike (branch `spike/huh-clone-input`, verdict: no-go) surfaced two genuine UX wins worth keeping without the library: input views collapsed their unfocused fields into bare text, and the clone destination preview hid with focus.

## What Changes

- All three text-input views (clone, create-branch, repo-name) render both fields persistently: focus moves the accent, never the geometry. `(empty)` is retired — blurred empty fields show their placeholders.
- The clone destination preview is always visible and tracks the URL live.
- A shared `inputFieldLabel` helper carries the focus-accent rule; all text inputs get an explicit shared width (`formInputWidth`), which also fixes a v2 textinput quirk where unset width renders only the placeholder's first rune.

## Capabilities

### Modified

- `tui-design-system`: input-field presentation.

## Impact

- clone_input.go, create_branch.go, repo_name.go, chrome.go helper, constants.go, model.go widths; one golden regenerated.
