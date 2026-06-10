package tui

import (
	"fmt"
	"strings"

	"github.com/abiswas97/sentei/internal/pipeline"
)

// progressChromeLines is the fixed line budget the layout spends outside
// step lists: title, blank, separators with surrounding blanks, bar, hints.
const progressChromeLines = 9

// ProgressLayout renders the standard progress view shared by every
// long-running flow: title, optional subtitle, separator, phases with
// indented steps (windowed to the terminal height), separator, overall
// progress bar, and key hints. Phase data is the phaseDisplay shape that
// buildPhaseDisplays produces from pipeline events; views with bespoke
// event types construct the same shape themselves.
type ProgressLayout struct {
	Title    string
	Subtitle string
	Phases   []phaseDisplay
	Width    int
	Height   int
	Hints    []KeyHint

	// OverallTotal overrides the bar's denominator when the flow knows its
	// full step count upfront and discovered phase totals would undercount.
	OverallDone  int
	OverallTotal int
}

func (l ProgressLayout) View() string {
	var b strings.Builder

	b.WriteString(viewTitle(l.Title))
	b.WriteString("\n")
	if l.Subtitle != "" {
		b.WriteString(styleAccent.Render("  " + truncateWithEllipsis(l.Subtitle, max(l.Width-2, 10))))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(viewSeparator(l.Width))
	b.WriteString("\n\n")

	stepBudget := l.Height - progressChromeLines - len(l.Phases)
	for _, p := range l.Phases {
		used := l.renderPhase(&b, p, stepBudget)
		stepBudget -= used
	}

	b.WriteString(viewSeparator(l.Width))
	b.WriteString("\n\n")

	done, total := l.OverallDone, l.OverallTotal
	if total == 0 {
		for _, p := range l.Phases {
			done += p.done
			total += p.total
			if p.total == 0 {
				// A phase with undiscovered work is still outstanding work;
				// without this the bar reads 100% beside pending phases.
				total++
			}
		}
	}
	b.WriteString(renderProgressBar(done, total))
	b.WriteString("\n")

	if len(l.Hints) > 0 {
		b.WriteString("\n")
		b.WriteString(viewKeyHints(l.Hints...))
		b.WriteString("\n")
	}

	return b.String()
}

// renderPhase writes one phase section and returns how many step lines it
// consumed from the windowing budget.
func (l ProgressLayout) renderPhase(b *strings.Builder, p phaseDisplay, stepBudget int) int {
	switch {
	case p.total == 0:
		// No work discovered yet: pending, never "100% done".
		fmt.Fprintf(b, "  %s %s  %s\n\n",
			styleIndicatorPending.Render(indicatorPending),
			stylePhasePending.Render(p.name),
			styleDim.Render("pending"))
		return 0

	case p.done == p.total && p.failed == 0:
		// Fully complete: collapse to a single line.
		fmt.Fprintf(b, "  %s %s  %s\n\n",
			styleIndicatorDone.Render(indicatorDone),
			stylePhaseDone.Render(p.name),
			styleDim.Render(fmt.Sprintf("%d/%d  100%%", p.total, p.total)))
		return 0
	}

	pct := (p.done * 100) / p.total
	if pct > 100 {
		pct = 100
	}
	ind := styleIndicatorActive.Render(indicatorActive)
	nameStyle := stylePhaseActive
	if p.failed > 0 {
		ind = styleIndicatorFailed.Render(indicatorFailed)
		if p.done == p.total {
			nameStyle = stylePhaseDone
		}
	}
	fmt.Fprintf(b, "  %s %s  %s\n",
		ind,
		nameStyle.Render(p.name),
		styleDim.Render(fmt.Sprintf("%d/%d  %d%%", min(p.done, p.total), p.total, pct)))

	window := WindowSteps(p.steps, stepBudget)
	for _, s := range window.Steps {
		var stepInd string
		switch s.status {
		case pipeline.StepDone, pipeline.StepSkipped:
			stepInd = styleIndicatorDone.Render(indicatorDone)
		case pipeline.StepRunning:
			stepInd = styleIndicatorActive.Render(indicatorActive)
		case pipeline.StepFailed:
			stepInd = styleIndicatorFailed.Render(indicatorFailed)
		default:
			stepInd = styleIndicatorPending.Render(indicatorPending)
		}
		label := truncateWithEllipsis(s.name, max(l.Width-6, 10))
		fmt.Fprintf(b, "    %s %s\n", stepInd, label)
	}
	used := len(window.Steps)
	if window.Windowed {
		b.WriteString(viewStatLine(window.Stats))
		b.WriteString("\n")
		used++
	}
	b.WriteString("\n")
	return used
}
