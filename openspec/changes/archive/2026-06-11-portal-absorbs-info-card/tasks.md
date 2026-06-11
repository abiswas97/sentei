# Portal Absorbs Info Card: Tasks

## 1. Implementation (TDD)

- [x] 1.1 Tests first: detailContent for both integration views returns the all-integrations page (sections in order, dep status rows); carousel state/keys gone; wheel maps only to the list cursor
- [x] 1.2 renderIntegrationsDetail in integration_list.go; detailContent routing in help.go/model.go
- [x] 1.3 Delete: showInfo/infoCursor, updateIntegrationInfo, renderIntegrationInfo, styleInfoCard, both lipgloss.Place blocks, wheel carousel branch, keys.Left/Right, in-view ? handlers
- [x] 1.4 viewChromeRows constant replaces the two msg.Height-6 literals

## 2. Verification and ship

- [x] 2.1 Full gauntlet; playground: ? on integrations opens portal with all sections, esc restores, h/l swallowed
- [x] 2.2 .impeccable.md Dialogs section + decision log updated
- [x] 2.3 PR refactor/portal-absorb, CI green, merge, cleanup, archive
