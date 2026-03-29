package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/repo"
)

type migrateInfoMsg struct {
	branch  string
	isDirty bool
	err     error
}

func loadMigrateInfo(runner git.CommandRunner, repoPath string) tea.Cmd {
	return func() tea.Msg {
		branch, err := runner.Run(repoPath, "branch", "--show-current")
		if err != nil {
			return migrateInfoMsg{err: err}
		}
		status, err := runner.Run(repoPath, "status", "--porcelain")
		if err != nil {
			return migrateInfoMsg{err: err}
		}
		isDirty := strings.TrimSpace(status) != ""
		return migrateInfoMsg{branch: branch, isDirty: isDirty}
	}
}

func (m Model) updateMigrateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = max(msg.Height-6, 5)
		return m, nil

	case migrateInfoMsg:
		if msg.err != nil {
			m.repo.validationErr = fmt.Sprintf("failed to load repo info: %v", msg.err)
		} else {
			m.repo.migrateInfo = MigrateInfo{
				Branch:  msg.branch,
				IsDirty: msg.isDirty,
			}
		}
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Back):
			m.view = menuView
			return m, nil

		case key.Matches(msg, keys.Confirm):
			opts := repo.MigrateOptions{
				RepoPath: m.repoPath,
			}
			m.repo.events = nil
			m.repo.result = nil
			m.repo.opType = "migrate"
			m.view = migrateProgressView
			return m, m.startRepoPipeline(opts)
		}
	}
	return m, nil
}

func (m Model) viewMigrateConfirm() string {
	var b strings.Builder

	b.WriteString(styleTitle.Render(fmt.Sprintf("  sentei %s Migrate to Bare Repository", "\u2500")))
	b.WriteString("\n\n")
	b.WriteString(styleDim.Render(fmt.Sprintf("  %s", m.repoPath)))
	b.WriteString("\n\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")

	if m.repo.validationErr != "" {
		b.WriteString(styleError.Render("  " + m.repo.validationErr))
		b.WriteString("\n\n")
	}

	branch := m.repo.migrateInfo.Branch
	if branch == "" {
		branch = styleDim.Render("detecting\u2026")
	}

	b.WriteString(fmt.Sprintf("  %-18s %s\n", styleDim.Render("Current branch"), branch))

	if m.repo.migrateInfo.IsDirty {
		b.WriteString(fmt.Sprintf("  %-18s %s\n", styleDim.Render("Status"),
			styleIndicatorWarning.Render(indicatorWarning+" uncommitted changes")))
	} else {
		b.WriteString(fmt.Sprintf("  %-18s %s\n", styleDim.Render("Status"), styleSuccess.Render("clean")))
	}

	b.WriteString("\n")
	b.WriteString("  This will:\n")
	b.WriteString(fmt.Sprintf("    %s Back up current repo\n", styleDim.Render("\u25cf")))
	b.WriteString(fmt.Sprintf("    %s Convert to bare repository structure\n", styleDim.Render("\u25cf")))
	b.WriteString(fmt.Sprintf("    %s Create worktree for %s\n", styleDim.Render("\u25cf"),
		filepath.Base(m.repoPath)+"/"+branch))

	if m.repo.migrateInfo.IsDirty {
		b.WriteString("\n")
		b.WriteString(styleIndicatorWarning.Render(fmt.Sprintf("  %s Uncommitted changes will be preserved in the backup", indicatorWarning)))
		b.WriteString("\n")
		b.WriteString(styleDim.Render("    but not in the new worktree"))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")
	b.WriteString(styleDim.Render("  enter migrate \u00b7 esc back"))
	b.WriteString("\n")

	return b.String()
}
