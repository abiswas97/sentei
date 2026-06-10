package tui

import (
	"fmt"
	"strings"
)

// KeyHint is one key-action pair rendered in a view's hint footer.
type KeyHint struct {
	Key    string
	Action string
}

// viewTitle renders the standard view title: `  sentei ─ <title>` in bold white.
func viewTitle(title string) string {
	text := "  sentei ─"
	if title != "" {
		text += " " + title
	}
	return styleTitle.Render(text)
}

// viewSeparator renders a dotted separator line in dim gray spanning
// width-4 characters. Widths of 4 or less fall back to a default of 40.
func viewSeparator(width int) string {
	if width <= 4 {
		width = 40
	}
	return styleSeparator.Render("  " + strings.Repeat("┄", width-4))
}

// viewKeyHints renders key-action pairs joined by ` · ` in dim gray.
func viewKeyHints(hints ...KeyHint) string {
	if len(hints) == 0 {
		return ""
	}
	parts := make([]string, len(hints))
	for i, h := range hints {
		parts[i] = h.Key + " " + h.Action
	}
	return styleDim.Render("  " + strings.Join(parts, " · "))
}

// truncateWithEllipsis cuts s to fit width, replacing the overflow with a
// trailing `…`. The only sanctioned way to fit overflowing paths, branch
// names, and error text — chrome content must never hard-clip at the
// terminal edge.
func truncateWithEllipsis(s string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= width {
		return s
	}
	if width == 1 {
		return "…"
	}
	return string(runes[:width-1]) + "…"
}

// viewStatLine renders the windowing legend: per-status counts (zero counts
// omitted) followed by how many of the total items are visible.
func viewStatLine(stats WindowStats) string {
	var parts []string
	if stats.Done > 0 {
		parts = append(parts, fmt.Sprintf("%s %d done", styleIndicatorDone.Render(indicatorDone), stats.Done))
	}
	if stats.Active > 0 {
		parts = append(parts, fmt.Sprintf("%s %d active", styleIndicatorActive.Render(indicatorActive), stats.Active))
	}
	if stats.Pending > 0 {
		parts = append(parts, fmt.Sprintf("%s %d pending", styleIndicatorPending.Render(indicatorPending), stats.Pending))
	}
	if stats.Failed > 0 {
		parts = append(parts, fmt.Sprintf("%s %d failed", styleIndicatorFailed.Render(indicatorFailed), stats.Failed))
	}
	legend := strings.Join(parts, "  ")
	showing := styleDim.Render(fmt.Sprintf("  showing %d of %d", stats.Showing, stats.Total))
	return "    " + legend + showing
}

// renderProgressBar renders the overall bar: accent-colored fill, dim track,
// percentage in the default foreground. Inputs are clamped so done > total
// can never produce a negative repeat count or a percentage above 100.
func renderProgressBar(done, total int) string {
	pct := 0
	if total > 0 {
		pct = (done * 100) / total
	}
	if pct > 100 {
		pct = 100
	}
	filled := 0
	if total > 0 {
		filled = (done * progressBarWidth) / total
	}
	filled = min(max(filled, 0), progressBarWidth)
	fill := styleAccent.Render(strings.Repeat("█", filled))
	track := styleDim.Render(strings.Repeat("░", progressBarWidth-filled))
	return fmt.Sprintf("  %s%s %d%%", fill, track, pct)
}
