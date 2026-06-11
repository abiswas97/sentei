# Design: voice-registry

Titles are data: `copy.go` holds one const per title (28), referenced at
every render site — changing the app's voice is editing one declaration.
Sentence case ("Remove worktrees", "Confirm deletion") replaces Title Case;
"Integrations" is unchanged (proper-noun-like single word). Portal titles
get their own consts and render bare inside the box (`styleTitle`, no
`sentei ─` prefix). The Info binding's default description becomes
"details" and per-view overrides die; "select entry" becomes "select".
