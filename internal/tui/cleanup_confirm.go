package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/cleanup"
	"github.com/abiswas97/sentei/internal/git"
)

// resolvedCleanupOpts returns the effective cleanup options, using defaults
// when no options have been explicitly set.
func (m Model) resolvedCleanupOpts() cleanup.Options {
	if m.cleanupOpts != nil {
		return *m.cleanupOpts
	}
	return cleanup.Options{Mode: cleanup.ModeSafe}
}

// cleanupConfirmationVM builds the ConfirmationViewModel for the cleanup flow.
func (m Model) cleanupConfirmationVM() ConfirmationViewModel {
	opts := m.resolvedCleanupOpts()

	mode := string(opts.Mode)
	if mode == "" {
		mode = string(cleanup.ModeSafe)
	}

	dryRun := "no"
	if opts.DryRun {
		dryRun = "yes"
	}

	flags := map[string]string{
		"mode": mode,
	}
	if opts.DryRun {
		flags["dry-run"] = "true"
	}

	return ConfirmationViewModel{
		Title: "Confirm Cleanup",
		Items: []ConfirmationItem{
			{Label: "Mode:", Value: mode},
			{Label: "Dry run:", Value: dryRun},
		},
		CLICommand: BuildCLICommand("cleanup", flags),
	}
}

func (m Model) updateCleanupConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = max(msg.Height-6, 5)
		return m, nil

	case ConfirmProceedMsg:
		opts := m.resolvedCleanupOpts()
		m.view = cleanupResultView
		m.remove.cleanupResult = nil
		return m, runCleanupWithOpts(m.runner, m.repoPath, opts)

	case ConfirmBackMsg:
		if m.cleanupOpts != nil {
			// Launched directly into cleanup confirmation — quit on back.
			return m, tea.Quit
		}
		m.view = menuView
		return m, nil
	}

	if cmd := UpdateConfirmation(msg); cmd != nil {
		return m, cmd
	}

	return m, nil
}

func (m Model) viewCleanupConfirm() string {
	return m.cleanupConfirmationVM().View()
}

// runCleanupWithOpts runs cleanup with the given options.
func runCleanupWithOpts(runner git.CommandRunner, repoPath string, opts cleanup.Options) tea.Cmd {
	return func() tea.Msg {
		result := cleanup.Run(runner, repoPath, opts, func(_ cleanup.Event) {})
		return standaloneCleanupDoneMsg{result: result}
	}
}
