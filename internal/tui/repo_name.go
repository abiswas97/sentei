package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateRepoName(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.repo.focusedField == 0 {
				m.repo.focusedField = 1
				m.repo.nameInput.Blur()
				return m, m.repo.locationInput.Focus()
			}
			m.repo.focusedField = 0
			m.repo.locationInput.Blur()
			return m, m.repo.nameInput.Focus()

		case key.Matches(msg, keys.Confirm):
			name := strings.TrimSpace(m.repo.nameInput.Value())
			location := strings.TrimSpace(m.repo.locationInput.Value())

			if name == "" {
				m.repo.validationErr = "repository name is required"
				return m, nil
			}
			if strings.Contains(name, " ") {
				m.repo.validationErr = "repository name cannot contain spaces"
				return m, nil
			}
			if location == "" {
				m.repo.validationErr = "location is required"
				return m, nil
			}

			// Expand ~ in location
			if strings.HasPrefix(location, "~/") {
				home, err := os.UserHomeDir()
				if err == nil {
					location = filepath.Join(home, location[2:])
					m.repo.locationInput.SetValue(location)
				}
			}

			// Check location exists
			if _, err := os.Stat(location); os.IsNotExist(err) {
				m.repo.validationErr = fmt.Sprintf("location does not exist: %s", location)
				return m, nil
			}

			// Check destination doesn't already exist
			dest := filepath.Join(location, name)
			if _, err := os.Stat(dest); err == nil {
				m.repo.validationErr = fmt.Sprintf("directory already exists: %s", dest)
				return m, nil
			}

			m.repo.validationErr = ""
			m.repo.focusedField = 0
			m.repo.locationInput.Blur()
			m.repo.nameInput.Blur()
			m.repo.publishGitHub = false
			m.repo.ghStatus = ""
			m.repo.optionsCursor = 0
			m.view = repoOptionsView
			return m, checkGitHubAuth(m.shell)
		}

		// Forward to focused input
		m.repo.validationErr = ""
		var cmd tea.Cmd
		if m.repo.focusedField == 0 {
			m.repo.nameInput, cmd = m.repo.nameInput.Update(msg)
		} else {
			m.repo.locationInput, cmd = m.repo.locationInput.Update(msg)
		}
		return m, cmd
	}
	return m, nil
}

func (m Model) viewRepoName() string {
	var b strings.Builder

	b.WriteString(styleTitle.Render(fmt.Sprintf("  sentei %s Create Repository", "\u2500")))
	b.WriteString("\n\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")

	b.WriteString("  Repository name\n")
	if m.repo.focusedField == 0 {
		b.WriteString("  " + m.repo.nameInput.View())
	} else {
		val := m.repo.nameInput.Value()
		if val == "" {
			val = styleDim.Render("(empty)")
		}
		b.WriteString("    " + val)
	}
	b.WriteString("\n\n")

	b.WriteString("  Location\n")
	if m.repo.focusedField == 1 {
		b.WriteString("  " + m.repo.locationInput.View())
	} else {
		val := m.repo.locationInput.Value()
		if val == "" {
			val = styleDim.Render("(empty)")
		}
		b.WriteString("    " + val)
	}
	b.WriteString("\n")

	if m.repo.validationErr != "" {
		b.WriteString("\n  " + styleError.Render(m.repo.validationErr))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")
	b.WriteString(styleDim.Render("  enter continue \u00b7 tab switch field \u00b7 esc back"))
	b.WriteString("\n")

	return b.String()
}
