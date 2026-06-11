package tui

import (
	"fmt"
	"strings"

	"github.com/abiswas97/sentei/internal/git"
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
	if m.view == cleanupPreviewView {
		return m.cleanupDetailContent()
	}
	if m.view == integrationSummaryView {
		return m.integrationSummaryDetailContent()
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
	return "Worktree Details", b.String()
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
