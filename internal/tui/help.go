package tui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/progress"
	"github.com/abiswas97/sentei/internal/repo"
)

// renderHelpSections formats key bindings as an aligned two-column table
// grouped by category. Sections are the keys.go presentation declarations,
// the same data that drives the footer hints.
func renderHelpSections(sections []keySection) string {
	keyWidth := 0
	for _, s := range sections {
		for _, bd := range s.bindings {
			keyWidth = max(keyWidth, len(bd.Help().Key))
		}
	}

	var b strings.Builder
	for i, s := range sections {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(styleDim.Render(s.name))
		b.WriteString("\n")
		for _, bd := range s.bindings {
			h := bd.Help()
			fmt.Fprintf(&b, "  %s  %s\n", styleAccent.Render(fmt.Sprintf("%-*s", keyWidth, h.Key)), h.Desc)
		}
	}
	return b.String()
}

// helpContent returns the contextual help title and body for the active view.
// Global entries already covered by a view section are omitted so keys never
// appear twice.
func (m Model) helpContent() (string, string) {
	name, sections := m.helpForView()
	covered := make(map[string]bool)
	for _, s := range sections {
		for _, bd := range s.bindings {
			covered[bd.Help().Key] = true
		}
	}
	global := keySection{name: helpGlobalSection.name}
	for _, bd := range helpGlobalSection.bindings {
		if !covered[bd.Help().Key] {
			global.bindings = append(global.bindings, bd)
		}
	}
	if len(global.bindings) > 0 {
		sections = append(sections, global)
	}
	return "Help — " + name, renderHelpSections(sections)
}

func (m Model) helpForView() (string, []keySection) {
	switch m.view {
	case menuView:
		return "Menu", menuSections
	case listView:
		return "Worktree List", listSections
	case confirmView:
		return "Confirm Deletion", confirmSections

	case progressView:
		return "Removing Worktrees", progressSections
	case createProgressView:
		return "Creating Worktree", progressSections
	case repoProgressView, migrateProgressView:
		return "Repository Operation", progressSections
	case integrationProgressView:
		return "Applying Integrations", progressSections

	case summaryView, createSummaryView, repoSummaryView, migrateSummaryView, integrationSummaryView:
		return "Summary", summarySections
	case createBranchView, repoNameView, cloneInputView:
		return "Input", inputSections
	case createOptionsView, repoOptionsView:
		return "Options", optionsSections
	case integrationListView, migrateIntegrationsView:
		return "Integrations", integrationSections
	case cleanupPreviewView:
		return "Cleanup Preview", cleanupPreviewSections
	case cleanupConfirmView, createConfirmView, cloneConfirmView, migrateConfirmView:
		return "Confirmation", confirmationSections

	default:
		return "sentei", nil
	}
}

// detailContent returns the `?` portal content for the active view, or empty
// when the view has no contextual details (the key then falls through to the
// view's own handling, e.g. the integration info card).
func (m Model) detailContent() (string, string) {
	if m.determinateProgressActive() {
		layout, ok := m.activeProgressLayout()
		if !ok || !progressNeedsDetails(layout, m.progressTopLevelError()) {
			return "", ""
		}
		return portalProgressDetails, renderProgressDetails(layout, m.progressTopLevelError())
	}
	if m.view == cleanupPreviewView {
		return m.cleanupDetailContent()
	}
	if m.view == integrationSummaryView {
		return m.integrationSummaryDetailContent()
	}
	if m.view == integrationListView || m.view == migrateIntegrationsView {
		if len(m.integ.integrations) == 0 {
			return "", ""
		}
		return portalIntegrationDetails, m.renderIntegrationsDetail()
	}
	if m.view != listView || m.remove.filterActive {
		return "", ""
	}
	if len(m.remove.visibleIndices) == 0 || m.remove.cursor >= len(m.remove.visibleIndices) {
		return "", ""
	}
	wt := m.remove.worktrees[m.remove.visibleIndices[m.remove.cursor]]

	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", styleTitle.Render(worktreeLabel(wt)))
	rows := []struct{ label, value string }{
		{"Path", wt.Path},
		{"Branch", wt.Branch},
		{"HEAD", wt.HEAD},
		{"Last commit", wt.LastCommitSubject},
		{"Committed", formatCommitDate(wt)},
		{"Status", worktreeStatusText(wt)},
	}
	valueWidth := max(m.portal.contentWidth()-14, 20)
	for _, r := range rows {
		if r.value == "" {
			continue
		}
		fmt.Fprintf(&b, "%s  %s\n", styleDim.Render(fmt.Sprintf("%-12s", r.label)), truncateWithEllipsis(r.value, valueWidth))
	}
	return portalWorktreeDetails, b.String()
}

func (m Model) progressTopLevelError() error {
	switch m.view {
	case progressView:
		return m.remove.run.result.Err
	case createProgressView:
		if m.create.result != nil {
			return m.create.result.Err
		}
	case repoProgressView, migrateProgressView:
		switch result := m.repo.result.(type) {
		case repo.CreateResult:
			return result.Err
		case repo.CloneResult:
			return result.Err
		case repo.MigrateResult:
			return result.Err
		}
	case integrationProgressView:
		return errors.Join(m.integ.prepareErr, m.integ.executionErr, m.integ.saveErr)
	}
	return nil
}

func progressNeedsDetails(layout ProgressLayout, topErr error) bool {
	if topErr != nil {
		return true
	}
	viewport := BuildProgressViewport(layout.Phases, layout.Height, layout.Completed)
	if viewport.HistoryOmitted > 0 || viewport.Queued > 0 {
		return true
	}
	if viewport.Focus != nil && WindowSteps(viewport.Focus.Steps, max(viewport.DetailRows-1, 0)).Windowed {
		return true
	}
	for _, phase := range layout.Phases {
		if phase.Failed > 0 {
			return true
		}
		for _, step := range phase.Steps {
			if step.Error != nil || step.Status == progress.StepFailed {
				return true
			}
		}
	}
	return false
}

func renderProgressDetails(layout ProgressLayout, topErr error) string {
	var b strings.Builder
	if topErr != nil {
		fmt.Fprintf(&b, "%s\n  %s\n\n", styleError.Render("Operation error"), topErr)
	}
	for phaseIndex, phase := range layout.Phases {
		if phaseIndex > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "%s\n", styleTitle.Render(phase.Name))
		fmt.Fprintf(&b, "  %s  %s\n", styleDim.Render("Phase ID"), phase.ID)
		fmt.Fprintf(&b, "  %s  %s\n", styleDim.Render("Status"), progressPhaseStatus(phase))
		for _, step := range phase.Steps {
			fmt.Fprintf(&b, "\n  %s\n", step.Name)
			fmt.Fprintf(&b, "    %s  %s\n", styleDim.Render("Step ID"), step.ID)
			fmt.Fprintf(&b, "    %s  %s\n", styleDim.Render("Status"), progressStepStatus(step.Status))
			if step.Message != "" {
				label := "Message"
				if step.Status == progress.StepSkipped {
					label = "Skip reason"
				}
				fmt.Fprintf(&b, "    %s  %s\n", styleDim.Render(label), step.Message)
			}
			if step.Error != nil {
				fmt.Fprintf(&b, "    %s  %s\n", styleError.Render("Error"), step.Error)
			}
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func progressPhaseStatus(phase progress.PhaseState) string {
	switch {
	case phase.Failed > 0:
		return "failed"
	case phase.Settled():
		return "done"
	case phaseHasStatus(phase, progress.StepRunning):
		return "running"
	case phase.Total == 0:
		return "empty"
	default:
		return "pending"
	}
}

func progressStepStatus(status progress.StepStatus) string {
	switch status {
	case progress.StepRunning:
		return "running"
	case progress.StepDone:
		return "done"
	case progress.StepFailed:
		return "failed"
	case progress.StepSkipped:
		return "skipped"
	default:
		return "pending"
	}
}

func formatCommitDate(wt git.Worktree) string {
	if wt.LastCommitDate.IsZero() {
		return ""
	}
	return wt.LastCommitDate.Format("2006-01-02 15:04")
}

func worktreeStatusText(wt git.Worktree) string {
	var parts []string
	if wt.HasUncommittedChanges {
		parts = append(parts, styleWarning.Render("uncommitted changes"))
	}
	if wt.HasUntrackedFiles {
		parts = append(parts, styleStatusUntracked.Render("untracked files"))
	}
	if wt.IsLocked {
		reason := "locked"
		if wt.LockReason != "" {
			reason += ": " + wt.LockReason
		}
		parts = append(parts, styleStatusLocked.Render(reason))
	}
	if wt.IsDetached {
		parts = append(parts, styleDim.Render("detached HEAD"))
	}
	if len(parts) == 0 {
		return styleSuccess.Render("clean")
	}
	return strings.Join(parts, ", ")
}
