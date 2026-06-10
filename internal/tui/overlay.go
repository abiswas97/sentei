package tui

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// compositeOverlay renders fg centered over bg and returns the combined
// canvas. Background lines are spliced ANSI-aware so colored content
// survives on both sides of the overlay. The canvas keeps the background's
// dimensions, growing only if the overlay is larger.
func compositeOverlay(fg, bg string) string {
	fgLines := strings.Split(fg, "\n")
	bgLines := strings.Split(bg, "\n")

	fgWidth := 0
	for _, l := range fgLines {
		fgWidth = max(fgWidth, ansi.StringWidth(l))
	}
	bgWidth := 0
	for _, l := range bgLines {
		bgWidth = max(bgWidth, ansi.StringWidth(l))
	}
	canvasWidth := max(bgWidth, fgWidth)

	for len(bgLines) < len(fgLines) {
		bgLines = append(bgLines, "")
	}

	top := max((len(bgLines)-len(fgLines))/2, 0)
	left := max((canvasWidth-fgWidth)/2, 0)

	out := make([]string, len(bgLines))
	copy(out, bgLines)
	for i, fgLine := range fgLines {
		row := top + i
		bgLine := bgLines[row]
		if pad := left - ansi.StringWidth(bgLine); pad > 0 {
			bgLine += strings.Repeat(" ", pad)
		}
		lhs := ansi.Truncate(bgLine, left, "")
		rhs := ansi.TruncateLeft(bgLine, left+ansi.StringWidth(fgLine), "")
		// Reset styling at the splice points so background SGR state never
		// bleeds into the overlay or vice versa.
		out[row] = lhs + "\x1b[0m" + fgLine + "\x1b[0m" + rhs
	}
	return strings.Join(out, "\n")
}
