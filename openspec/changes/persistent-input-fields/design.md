# Design: persistent-input-fields

Backport of the huh spike's layout without the dependency. The huh verdict
(no-go) and its findings live in the decision log; the spike branch is
preserved for reference. The v2 textinput placeholder bug: with Width unset,
`placeholderView` sizes its rune buffer to Width+1 = 1 and renders a single
character — masked before because blurred inputs were never rendered.
