package tui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	progressbar "charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"

	"github.com/abiswas97/sentei/internal/progress"
)

// progressChromeLines is the fixed line budget the layout spends outside
// step lists: title, blank, separators with surrounding blanks, bar, hints.
const progressChromeLines = 9

// ProgressLayout renders the standard progress view shared by every
// long-running flow: title, optional subtitle, separator, phases with
// indented steps (windowed to the terminal height), separator, overall
// progress bar, and key hints. Phase data is the progress.PhaseState shape that
// buildPhaseDisplays produces from pipeline events; views with bespoke
// event types construct the same shape themselves.
type ProgressLayout struct {
	Title    string
	Subtitle string
	Phases   []progress.PhaseState
	Width    int
	Height   int
	Hints    []key.Binding

	// OverallTotal overrides the bar's denominator when the flow knows its
	// full step count upfront and discovered phase totals would undercount.
	OverallDone  int
	OverallTotal int

	// Bar, Elapsed, and ActiveGlyph are injected by the model's render path
	// (animated spring bar, stopwatch readout, spinner frame). Left empty,
	// View falls back to the static bar, no elapsed line, and the static
	// active-indicator fallback: direct constructions stay pure.
	Bar         string
	Elapsed     string
	ActiveGlyph string

	// Motion carries the live shimmer closures and star frame; nil falls
	// back to static styles so direct constructions stay pure.
	Motion *Motion

	// Completed marks a flow whose result has arrived: phases that never
	// discovered work render as skipped and stop counting as outstanding.
	Completed bool
}

// activeGlyph returns the styled active indicator: the injected animation
// frame, or the static fallback for pure constructions.
func (l ProgressLayout) activeGlyph() string {
	if l.ActiveGlyph != "" {
		return l.ActiveGlyph
	}
	return styleIndicatorActive.Render(indicatorActiveFallback)
}

// overall returns the bar's done/total, honoring the explicit override and
// counting each undiscovered phase as outstanding work so the bar never
// reads 100% beside pending phases.
func (l ProgressLayout) overall() (int, int) {
	if l.OverallTotal != 0 {
		return l.OverallDone, l.OverallTotal
	}
	done, total := 0, 0
	for _, p := range l.Phases {
		done += p.Done
		total += p.Total
		if p.Total == 0 && !l.Completed {
			total++
		}
	}
	return done, total
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

	if l.Bar != "" {
		b.WriteString(l.Bar)
	} else {
		done, total := l.overall()
		b.WriteString(renderProgressBar(done, total, overallBarWidth(l.Width)-progressBarPercentReserve))
	}
	if l.Elapsed != "" {
		b.WriteString("  " + l.Elapsed)
	}
	b.WriteString("\n")

	if len(l.Hints) > 0 {
		b.WriteString("\n")
		b.WriteString(viewFooter(l.Width, l.Hints))
		b.WriteString("\n")
	}

	return b.String()
}

// renderPhase writes one phase section and returns how many step lines it
// consumed from the windowing budget.
func (l ProgressLayout) renderPhase(b *strings.Builder, p progress.PhaseState, stepBudget int) int {
	switch {
	case p.Total == 0:
		if l.Completed {
			// The flow finished and this phase never had work: skipped.
			fmt.Fprintf(b, "  %s %s  %s\n\n",
				styleDim.Render("–"),
				stylePhasePending.Render(p.Name),
				styleDim.Render("skipped"))
			return 0
		}
		// No work discovered yet: pending, never "100% done".
		fmt.Fprintf(b, "  %s %s  %s\n\n",
			styleIndicatorPending.Render(indicatorPending),
			stylePhasePending.Render(p.Name),
			styleDim.Render("pending"))
		return 0

	case p.Done == p.Total && p.Failed == 0:
		// Fully complete: collapse to a single line.
		fmt.Fprintf(b, "  %s %s  %s\n\n",
			styleIndicatorDone.Render(indicatorDone),
			stylePhaseDone.Render(p.Name),
			styleDim.Render(fmt.Sprintf("%d/%d  100%%", p.Total, p.Total)))
		return 0
	}

	pct := (p.Done * 100) / p.Total
	if pct > 100 {
		pct = 100
	}
	counts := styleDim.Render(fmt.Sprintf("%d/%d  %d%%", min(p.Done, p.Total), p.Total, pct))
	switch {
	case p.Failed > 0:
		nameStyle := stylePhaseActive
		if p.Done == p.Total {
			nameStyle = stylePhaseDone
		}
		fmt.Fprintf(b, "  %s %s  %s\n",
			styleIndicatorFailed.Render(indicatorFailed), nameStyle.Render(p.Name), counts)
	case l.Motion != nil:
		// The star rides inside the headline's accent shimmer band.
		fmt.Fprintf(b, "  %s  %s\n", l.Motion.Accent(l.Motion.Frame+" "+p.Name), counts)
	default:
		fmt.Fprintf(b, "  %s %s  %s\n",
			l.activeGlyph(), stylePhaseActive.Render(p.Name), counts)
	}

	window := WindowSteps(p.Steps, stepBudget)
	for _, s := range window.Steps {
		label := truncateWithEllipsis(s.Name, max(l.Width-6, 10))
		if s.Status == progress.StepRunning && l.Motion != nil {
			// Working steps shimmer in the body ramp, star in the band.
			fmt.Fprintf(b, "    %s\n", l.Motion.Body(l.Motion.Frame+" "+label))
			continue
		}
		var stepInd string
		switch s.Status {
		case progress.StepDone, progress.StepSkipped:
			stepInd = styleIndicatorDone.Render(indicatorDone)
		case progress.StepRunning:
			stepInd = l.activeGlyph()
		case progress.StepFailed:
			stepInd = styleIndicatorFailed.Render(indicatorFailed)
		default:
			stepInd = styleIndicatorPending.Render(indicatorPending)
		}
		fmt.Fprintf(b, "    %s %s\n", stepInd, label)
	}
	used := len(window.Steps)
	if window.Windowed {
		b.WriteString(viewStatLine(window.Stats, l.activeGlyph()))
		b.WriteString("\n")
		used++
	}
	b.WriteString("\n")
	return used
}

// newOverallBar constructs the spring-animated overall bar: block fill
// characters, gradient scaled to the filled portion, sized responsively
// from WindowSizeMsg (the construction width is only the pre-frame floor).
func newOverallBar() progressbar.Model {
	return progressbar.New(
		progressbar.WithWidth(minProgressBarWidth),
		progressbar.WithFillCharacters('█', '░'),
		progressbar.WithColors(colorBarStart, colorBarEnd),
		progressbar.WithScaled(true),
		// Luxurious spring: a full sweep takes ~1.2s (bar visually full at
		// ~1.0s), inside the hold plus settle floor. Critically damped:
		// the component clamps at 100%, so any overshoot would render as
		// the bar retreating.
		progressbar.WithSpringOptions(6, 1),
	)
}

// determinateProgressActive reports whether a ProgressLayout-rendered flow
// is on screen: the only place bar frames and stopwatch ticks may animate.
func (m Model) determinateProgressActive() bool {
	switch m.view {
	case progressView, createProgressView, repoProgressView, migrateProgressView, integrationProgressView:
		return true
	}
	return false
}

// activeProgressLayout returns the layout for the flow on screen, the single
// source for done/total shared by rendering and the spring target.
func (m Model) activeProgressLayout() (ProgressLayout, bool) {
	switch m.view {
	case progressView:
		return m.removalLayout(), true
	case createProgressView:
		return m.createLayout(), true
	case repoProgressView, migrateProgressView:
		return m.repoLayout(), true
	case integrationProgressView:
		return m.integrationLayout(), true
	}
	return ProgressLayout{}, false
}

// syncProgressBar springs the bar toward the active flow's completion and
// starts the elapsed stopwatch on the flow's first event.
func (m *Model) syncProgressBar() tea.Cmd {
	l, ok := m.activeProgressLayout()
	if !ok {
		return nil
	}
	done, total := l.overall()
	pct := 0.0
	switch {
	case total > 0:
		pct = min(float64(done)/float64(total), 1)
	case l.Completed:
		// Nothing was ever discovered and the flow is done: that is 100%.
		pct = 1
	}
	cmds := []tea.Cmd{m.bar.SetPercent(pct)}
	if !m.watch.Running() {
		cmds = append(cmds, m.watch.Start())
	}
	return tea.Batch(cmds...)
}

// renderProgressLayout injects the animated bar and elapsed readout into a
// flow's layout. The percentage text states actual progress; only the cells
// ease toward it. Bar colors come from the live palette tokens.
func (m Model) renderProgressLayout(l ProgressLayout) string {
	bar := m.bar
	// Colors re-read the live palette tokens each render: the adaptive
	// palette can arrive after the bar is constructed. A completed flow
	// settles the fill green for its hold — the bar joins the ✦ moment.
	if l.Completed {
		progressbar.WithColors(colorBarDoneStart, colorBarDoneEnd)(&bar)
	} else {
		progressbar.WithColors(colorBarStart, colorBarEnd)(&bar)
	}
	bar.EmptyColor = colorDim
	// The native percentage follows the displayed fill, so bar and label
	// never disagree; the phase headers state actual counts.
	l.Bar = "  " + bar.View()
	l.Elapsed = styleDim.Render(fmt.Sprintf("elapsed %ds", int(time.Since(m.progressStartedAt).Seconds())))
	l.ActiveGlyph = starGlyph(rampAccent, m.motionTick)
	l.Motion = m.motion()
	return l.View()
}
