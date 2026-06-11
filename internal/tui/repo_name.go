package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

func (m Model) updateRepoName(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, keys.Back):
			m.view = menuView
			return m, nil

		case key.Matches(msg, keys.Tab):
			if m.repo.focusedField == 0 {
				m.repo.focusedField = 1
				m.repo.nameInput.Blur()
				m.repo.locationInput.CursorEnd()
				return m, m.repo.locationInput.Focus()
			}
			m.repo.focusedField = 0
			m.repo.locationInput.Blur()
			m.repo.nameInput.CursorEnd()
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

			// Check parent directory exists
			parentDir := filepath.Dir(location)
			if _, err := os.Stat(parentDir); os.IsNotExist(err) {
				m.repo.validationErr = fmt.Sprintf("parent directory does not exist: %s", parentDir)
				return m, nil
			}

			// Check destination doesn't already exist
			if _, err := os.Stat(location); err == nil {
				m.repo.validationErr = fmt.Sprintf("directory already exists: %s", location)
				return m, nil
			}

			m.repo.validationErr = ""
			m.repo.focusedField = 0
			m.repo.locationInput.Blur()
			m.repo.nameInput.Blur()
			m.repo.createWorktree = true
			m.repo.publishGitHub = false
			m.repo.ghStatus = ""
			m.repo.optionsCursor = repoOptPublish
			m.view = repoOptionsView
			return m, checkGitHubAuth(m.shell)
		}

		// Forward to focused input
		m.repo.validationErr = ""
		var cmd tea.Cmd
		if m.repo.focusedField == 0 {
			m.repo.nameInput, cmd = m.repo.nameInput.Update(msg)
			// Auto-update location to show <repoPath>/<name> as user types
			name := strings.TrimSpace(m.repo.nameInput.Value())
			if name != "" {
				m.repo.locationInput.SetValue(filepath.Join(m.repoPath, name))
			} else {
				m.repo.locationInput.SetValue(m.repoPath)
			}
		} else {
			m.repo.locationInput, cmd = m.repo.locationInput.Update(msg)
		}
		return m, cmd
	}
	return m, nil
}

func (m Model) viewRepoName() string {
	var b strings.Builder

	b.WriteString(viewTitle(titleCreateRepo))
	b.WriteString("\n\n")
	b.WriteString(viewSeparator(m.width))
	b.WriteString("\n\n")

	if m.repo.focusedField == 0 {
		b.WriteString(styleAccent.Render("  Repository name") + "\n")
	} else {
		b.WriteString("  Repository name\n")
	}
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

	if m.repo.focusedField == 1 {
		b.WriteString(styleAccent.Render("  Location") + "\n")
	} else {
		b.WriteString("  Location\n")
	}
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
	b.WriteString(viewSeparator(m.width))
	b.WriteString("\n\n")
	b.WriteString(viewFooter(m.width, repoNameFooter))
	b.WriteString("\n")

	return b.String()
}
