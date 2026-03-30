package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/repo"
)

// Screen 1: Backup cleanup decision
func (m Model) updateMigrateSummary(msg tea.Msg) (tea.Model, tea.Cmd) {
	result, ok := m.repo.result.(repo.MigrateResult)
	if !ok {
		return m, tea.Quit
	}

	// Check for critical migration failures
	for _, phase := range result.Phases {
		if phase.Name == "Migrate" && phase.HasFailures() {
			// Critical failure — only q to quit
			if msg, ok := msg.(tea.KeyMsg); ok && key.Matches(msg, keys.Quit) {
				return m, tea.Quit
			}
			return m, nil
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = max(msg.Height-6, 5)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Yes):
			// Delete backup, then show integration selection screen
			if result.BackupPath != "" {
				_ = repo.DeleteBackup(result.BackupPath)
			}
			m.view = migrateIntegrationsView
			return m, loadMigrateIntegrations(m.migrateWorktreePath(result))

		case key.Matches(msg, keys.No):
			// Keep backup, show integration selection screen
			m.view = migrateIntegrationsView
			return m, loadMigrateIntegrations(m.migrateWorktreePath(result))

		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) viewMigrateSummary() string {
	var b strings.Builder

	result, ok := m.repo.result.(repo.MigrateResult)
	if !ok {
		return "  Migration result unavailable\n"
	}

	// Check for critical failures
	hasCriticalFailure := false
	var failErr error
	for _, phase := range result.Phases {
		if phase.Name == "Migrate" && phase.HasFailures() {
			hasCriticalFailure = true
			for _, step := range phase.Steps {
				if step.Error != nil {
					failErr = step.Error
					break
				}
			}
			break
		}
	}

	b.WriteString(styleTitle.Render(fmt.Sprintf("  sentei %s Migration Complete", "\u2500")))
	b.WriteString("\n\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")

	if hasCriticalFailure {
		errMsg := "unknown error"
		if failErr != nil {
			errMsg = failErr.Error()
		}
		errWidth := max(m.width-8, 30)
		b.WriteString(fmt.Sprintf("  %s Migration failed\n\n",
			styleIndicatorFailed.Render(indicatorFailed)))
		b.WriteString("    " + styleError.Width(errWidth).Render(errMsg))
		b.WriteString("\n\n")
		if result.BackupPath != "" {
			b.WriteString("  Your original repo is backed up at:\n")
			b.WriteString(fmt.Sprintf("    %s\n\n", styleDim.Render(result.BackupPath)))
			b.WriteString("  To restore:\n")
			b.WriteString(fmt.Sprintf("    rm -rf %s && mv %s %s\n",
				result.BareRoot, result.BackupPath, result.BareRoot))
		}
		b.WriteString("\n")
		b.WriteString(separator(m.width))
		b.WriteString("\n\n")
		b.WriteString(styleDim.Render("  q quit"))
		b.WriteString("\n")
		return b.String()
	}

	repoName := filepath.Base(result.BareRoot)
	b.WriteString(fmt.Sprintf("  %s %s migrated\n\n",
		styleIndicatorDone.Render(indicatorDone), repoName))

	b.WriteString(fmt.Sprintf("    %-10s %s\n", styleDim.Render("Path"), result.BareRoot))
	b.WriteString(fmt.Sprintf("    %-10s %s\n", styleDim.Render("Branch"), result.Branch))
	if result.BackupPath != "" {
		backupName := filepath.Base(result.BackupPath)
		sizeHint := ""
		if result.BackupSize != "" {
			sizeHint = "  " + styleDim.Render(result.BackupSize)
		}
		b.WriteString(fmt.Sprintf("    %-10s %s%s\n", styleDim.Render("Backup"), backupName, sizeHint))
	}

	b.WriteString("\n")
	b.WriteString("  Delete backup?\n")
	b.WriteString("\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")
	b.WriteString(styleDim.Render("  y delete \u00b7 n keep \u00b7 q quit"))
	b.WriteString("\n")

	return b.String()
}

// Screen 2: What next — continue in sentei or exit
func (m Model) updateMigrateNext(msg tea.Msg) (tea.Model, tea.Cmd) {
	result, ok := m.repo.result.(repo.MigrateResult)
	if !ok {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = max(msg.Height-6, 5)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Confirm):
			// Re-launch sentei at the migrated repo
			return m, relaunchSentei(result.BareRoot)

		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) viewMigrateNext() string {
	var b strings.Builder

	result, ok := m.repo.result.(repo.MigrateResult)
	if !ok {
		return "  Migration result unavailable\n"
	}

	repoName := filepath.Base(result.BareRoot)
	worktreePath := result.WorktreePath
	if worktreePath == "" {
		worktreePath = filepath.Join(result.BareRoot, result.Branch)
	}

	b.WriteString(styleTitle.Render(fmt.Sprintf("  sentei %s Migration Complete", "\u2500")))
	b.WriteString("\n\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("  %s %s ready\n\n",
		styleIndicatorDone.Render(indicatorDone), repoName))

	b.WriteString(fmt.Sprintf("    cd %s\n", styleDim.Render(worktreePath)))
	b.WriteString("\n")

	b.WriteString(styleDim.Render("  Your repo is ready for worktrees."))
	b.WriteString("\n")
	b.WriteString(styleDim.Render("  Continue in sentei to create worktrees"))
	b.WriteString("\n")
	b.WriteString(styleDim.Render("  and set up your workspace, or exit"))
	b.WriteString("\n")
	b.WriteString(styleDim.Render("  to your shell."))
	b.WriteString("\n")

	b.WriteString("\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")
	b.WriteString(styleDim.Render("  enter open in sentei \u00b7 q exit"))
	b.WriteString("\n")

	return b.String()
}
