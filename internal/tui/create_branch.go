package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/creator"
	"github.com/abiswas97/sentei/internal/ecosystem"
	"github.com/abiswas97/sentei/internal/integration"
)

type branchValidationError struct {
	message string
}

func validateBranchName(name string, existingWorktrees []string) *branchValidationError {
	if name == "" {
		return &branchValidationError{message: "branch name is required"}
	}
	if strings.Contains(name, " ") {
		return &branchValidationError{message: "branch name cannot contain spaces"}
	}
	if strings.Contains(name, "..") {
		return &branchValidationError{message: "branch name cannot contain '..'"}
	}
	sanitized := creator.SanitizeBranchPath(name)
	for _, wt := range existingWorktrees {
		if strings.HasSuffix(wt, "/"+sanitized) || wt == sanitized {
			return &branchValidationError{message: fmt.Sprintf("worktree %q already exists", sanitized)}
		}
	}
	return nil
}

func (m Model) updateCreateBranch(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = max(msg.Height-6, 5)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Back):
			m.view = menuView
			return m, nil

		case key.Matches(msg, keys.Tab):
			if m.create.focusedField == 0 {
				m.create.focusedField = 1
				m.create.branchInput.Blur()
				m.create.baseInput.Focus()
				return m, m.create.baseInput.Cursor.BlinkCmd()
			}
			m.create.focusedField = 0
			m.create.baseInput.Blur()
			m.create.branchInput.Focus()
			return m, m.create.branchInput.Cursor.BlinkCmd()

		case key.Matches(msg, keys.Confirm):
			branch := m.create.branchInput.Value()
			var existingPaths []string
			for _, wt := range m.remove.worktrees {
				existingPaths = append(existingPaths, wt.Path)
			}
			if err := validateBranchName(branch, existingPaths); err != nil {
				m.create.validationErr = err.message
				return m, nil
			}

			m.create.validationErr = ""
			m.prepareCreateOptions()
			m.view = createOptionsView
			return m, nil
		}

		// Forward key to focused input
		m.create.validationErr = ""
		var cmd tea.Cmd
		if m.create.focusedField == 0 {
			m.create.branchInput, cmd = m.create.branchInput.Update(msg)
		} else {
			m.create.baseInput, cmd = m.create.baseInput.Update(msg)
		}
		return m, cmd
	}
	return m, nil
}

func (m *Model) prepareCreateOptions() {
	if m.cfg == nil {
		return
	}

	// Detect ecosystems from a source worktree
	sourceWT := m.findSourceWorktree()
	if sourceWT != "" {
		registry := ecosystem.NewRegistry(m.cfg.Ecosystems)
		detected, _ := registry.Detect(sourceWT)
		m.create.ecosystems = nil
		for _, eco := range detected {
			m.create.ecosystems = append(m.create.ecosystems, eco.Config)
			m.create.ecoEnabled[eco.Name] = true
		}
	}

	// Load integrations
	m.create.integrations = nil
	enabledSet := make(map[string]bool)
	for _, name := range m.cfg.IntegrationsEnabled {
		enabledSet[name] = true
	}
	for _, integ := range integration.All() {
		m.create.integrations = append(m.create.integrations, integ)
		m.create.intEnabled[integ.Name] = enabledSet[integ.Name]
	}
}

func (m Model) findSourceWorktree() string {
	for _, wt := range m.remove.worktrees {
		branch := stripBranchPrefix(wt.Branch)
		if branch == "main" || branch == "master" {
			return wt.Path
		}
	}
	if len(m.remove.worktrees) > 0 {
		return m.remove.worktrees[0].Path
	}
	return ""
}

func (m Model) viewCreateBranch() string {
	var b strings.Builder

	b.WriteString(styleTitle.Render(fmt.Sprintf("  sentei %s Create Worktree", "\u2500")))
	b.WriteString("\n\n")

	b.WriteString(styleDim.Render(fmt.Sprintf("  %s %s %s", strings.TrimPrefix(m.repoPath, ""), "\u00b7", m.repoPath)))
	b.WriteString("\n\n")

	b.WriteString(separator(m.width))
	b.WriteString("\n\n")

	b.WriteString("  Branch name\n")
	if m.create.focusedField == 0 {
		b.WriteString("  " + m.create.branchInput.View())
	} else {
		val := m.create.branchInput.Value()
		if val == "" {
			val = styleDim.Render("(empty)")
		}
		b.WriteString("    " + val)
	}
	b.WriteString("\n")
	if m.create.validationErr != "" {
		b.WriteString("  " + styleError.Render(m.create.validationErr))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	b.WriteString("  Base branch\n")
	if m.create.focusedField == 1 {
		b.WriteString("  " + m.create.baseInput.View())
	} else {
		val := m.create.baseInput.Value()
		if val == "" {
			val = "main"
		}
		b.WriteString("    " + val)
	}
	b.WriteString("\n\n")

	b.WriteString(separator(m.width))
	b.WriteString("\n\n")
	b.WriteString(styleDim.Render("  enter continue \u00b7 tab switch field \u00b7 esc back"))
	b.WriteString("\n")

	return b.String()
}
