package tui

import (
	"fmt"
	"strings"

	"github.com/abiswas97/sentei/internal/git"
)

type helpEntry struct {
	key  string
	desc string
}

type helpSection struct {
	name    string
	entries []helpEntry
}

var helpGlobalSection = helpSection{name: "Global", entries: []helpEntry{
	{"F1", "toggle this help"},
	{"q / ctrl+c", "quit"},
}}

// renderHelpSections formats key bindings as an aligned two-column table
// grouped by category.
func renderHelpSections(sections []helpSection) string {
	keyWidth := 0
	for _, s := range sections {
		for _, e := range s.entries {
			keyWidth = max(keyWidth, len(e.key))
		}
	}

	var b strings.Builder
	for i, s := range sections {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(styleDim.Render(s.name))
		b.WriteString("\n")
		for _, e := range s.entries {
			fmt.Fprintf(&b, "  %s  %s\n", styleAccent.Render(fmt.Sprintf("%-*s", keyWidth, e.key)), e.desc)
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
		for _, e := range s.entries {
			covered[e.key] = true
		}
	}
	global := helpSection{name: helpGlobalSection.name}
	for _, e := range helpGlobalSection.entries {
		if !covered[e.key] {
			global.entries = append(global.entries, e)
		}
	}
	if len(global.entries) > 0 {
		sections = append(sections, global)
	}
	return "Help — " + name, renderHelpSections(sections)
}

func (m Model) helpForView() (string, []helpSection) {
	switch m.view {
	case menuView:
		return "Menu", []helpSection{{name: "Navigation", entries: []helpEntry{
			{"j/k, ↑/↓", "move between entries"},
			{"enter", "select entry"},
		}}}

	case listView:
		return "Worktree List", []helpSection{
			{name: "Navigation", entries: []helpEntry{
				{"j/k, ↑/↓", "move cursor"},
				{"pgup/pgdn", "page"},
			}},
			{name: "Selection", entries: []helpEntry{
				{"space", "toggle worktree"},
				{"a", "select all"},
				{"enter", "delete selected"},
			}},
			{name: "Organize", entries: []helpEntry{
				{"/", "filter by name"},
				{"s / S", "cycle / reverse sort"},
				{"?", "details for highlighted worktree"},
			}},
		}

	case confirmView:
		return "Confirm Deletion", []helpSection{{name: "Actions", entries: []helpEntry{
			{"y", "delete the selected worktrees"},
			{"n", "go back to the list"},
		}}}

	case progressView:
		return "Removing Worktrees", progressHelp()
	case createProgressView:
		return "Creating Worktree", progressHelp()
	case repoProgressView, migrateProgressView:
		return "Repository Operation", progressHelp()
	case integrationProgressView:
		return "Applying Integrations", progressHelp()

	case summaryView, createSummaryView, repoSummaryView, migrateSummaryView, integrationSummaryView:
		return "Summary", []helpSection{{name: "Actions", entries: []helpEntry{
			{"enter", "continue"},
			{"esc", "back"},
		}}}

	case createBranchView, repoNameView, cloneInputView:
		return "Input", []helpSection{{name: "Editing", entries: []helpEntry{
			{"tab", "switch field"},
			{"enter", "continue"},
			{"esc", "back"},
		}}}

	case createOptionsView, repoOptionsView:
		return "Options", []helpSection{{name: "Actions", entries: []helpEntry{
			{"j/k", "move"},
			{"space", "toggle option"},
			{"enter", "continue"},
			{"esc", "back"},
		}}}

	case integrationListView, migrateIntegrationsView:
		return "Integrations", []helpSection{{name: "Actions", entries: []helpEntry{
			{"j/k", "move"},
			{"space", "toggle integration"},
			{"?", "integration info"},
			{"enter", "apply changes"},
			{"esc", "back"},
		}}}

	case cleanupPreviewView:
		return "Cleanup Preview", []helpSection{{name: "Actions", entries: []helpEntry{
			{"enter", "run safe cleanup"},
			{"a", "aggressive cleanup (when available)"},
			{"?", "full branch details"},
			{"esc", "back"},
		}}}

	case cleanupConfirmView, createConfirmView, cloneConfirmView, migrateConfirmView:
		return "Confirmation", []helpSection{{name: "Actions", entries: []helpEntry{
			{"enter", "confirm"},
			{"esc", "back"},
		}}}

	default:
		return "sentei", nil
	}
}

func progressHelp() []helpSection {
	return []helpSection{{name: "Actions", entries: []helpEntry{
		{"q / ctrl+c", "quit (operation keeps running in git)"},
	}}}
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
