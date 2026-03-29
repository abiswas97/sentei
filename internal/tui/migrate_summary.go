package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/repo"
)

func (m Model) updateMigrateSummary(msg tea.Msg) (tea.Model, tea.Cmd) {
	result, ok := m.repo.result.(repo.MigrateResult)
	if !ok {
		return m, tea.Quit
	}

	// Check if migration had critical failures (nothing to launch)
	hasCriticalFailure := false
	for _, phase := range result.Phases {
		if phase.Name == "Migrate" && phase.HasFailures() {
			hasCriticalFailure = true
			break
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
			if hasCriticalFailure {
				return m, tea.Quit
			}
			if result.BackupPath != "" {
				_ = repo.DeleteBackup(result.BackupPath)
			}
			return m, relaunchSentei(result.BareRoot)

		case key.Matches(msg, keys.No):
			if hasCriticalFailure {
				return m, tea.Quit
			}
			return m, relaunchSentei(result.BareRoot)

		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		}
	}
	return m, nil
}

func relaunchSentei(repoPath string) tea.Cmd {
	senteiPath, err := os.Executable()
	if err != nil {
		senteiPath = "sentei"
	}
	c := exec.Command(senteiPath, repoPath)
	c.Env = os.Environ()
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return tea.Quit()
	})
}

func (m Model) viewMigrateSummary() string {
	var b strings.Builder

	result, ok := m.repo.result.(repo.MigrateResult)
	if !ok {
		return "  Migration result unavailable\n"
	}

	// Check for critical failures in Migrate phase
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
		b.WriteString(fmt.Sprintf("  %s Migration failed: %s\n\n",
			styleIndicatorFailed.Render(indicatorFailed), styleError.Render(errMsg)))
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
		b.WriteString(fmt.Sprintf("    %-10s %s\n", styleDim.Render("Backup"), result.BackupPath))
	}

	b.WriteString("\n")
	if result.BackupPath != "" && result.BackupSize != "" {
		b.WriteString(fmt.Sprintf("  Delete backup? (saves %s)\n", result.BackupSize))
	} else {
		b.WriteString("  Delete backup?\n")
	}
	b.WriteString("\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")
	b.WriteString(styleDim.Render("  y delete backup \u00b7 n keep and open in sentei \u00b7 q quit"))
	b.WriteString("\n")

	return b.String()
}
