# Canvas Overlay: Design

## Decisions

### D1: Compositor, not bare layer Compose
`Layer.Draw` ignores position; positioning lives in `Compositor`, which flattens the layer tree to absolute coordinates and z-sorts before drawing. `compositeOverlay` builds `NewCanvas(w,h).Compose(NewCompositor(bgLayer, fgLayer.X().Y().Z(1)))`.

### D2: Contract preserved by the existing tests
The four overlay tests (centering, ANSI survival, short/shorter backgrounds) and the portal suite are the acceptance bar; the function signature is unchanged so no caller moves.
