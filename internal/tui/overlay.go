package tui

import (
	"charm.land/lipgloss/v2"
)

// compositeOverlay renders fg centered over bg and returns the combined
// canvas, composited by lipgloss layers so wide runes and ANSI styling are
// handled cell-accurately. The canvas keeps the background's dimensions,
// growing only if the overlay is larger.
func compositeOverlay(fg, bg string) string {
	fgW, fgH := lipgloss.Width(fg), lipgloss.Height(fg)
	bgW, bgH := lipgloss.Width(bg), lipgloss.Height(bg)
	w, h := max(bgW, fgW), max(bgH, fgH)

	// The overlay claims its rows in full (centered, space-padded flanks)
	// so no orphan background glyphs survive beside the box; the background
	// stays visible above and below.
	canvas := lipgloss.NewCanvas(w, h)
	canvas.Compose(lipgloss.NewCompositor(
		lipgloss.NewLayer(bg),
		lipgloss.NewLayer(lipgloss.PlaceHorizontal(w, lipgloss.Center, fg)).
			Y(max((h-fgH)/2, 0)).
			Z(1),
	))
	return canvas.Render()
}
