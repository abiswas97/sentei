package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/repo"
)

func (m Model) updateCloneInput(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.repo.cloneFocusedField == 0 {
				m.repo.cloneFocusedField = 1
				m.repo.urlInput.Blur()
				return m, m.repo.cloneNameInput.Focus()
			}
			m.repo.cloneFocusedField = 0
			m.repo.cloneNameInput.Blur()
			return m, m.repo.urlInput.Focus()

		case key.Matches(msg, keys.Confirm):
			url := strings.TrimSpace(m.repo.urlInput.Value())
			if url == "" {
				m.repo.validationErr = "repository URL is required"
				return m, nil
			}

			name := strings.TrimSpace(m.repo.cloneNameInput.Value())
			if name == "" {
				name = repo.DeriveRepoName(url)
			}

			// Use the repo path as location (parent directory for non-bare)
			location := m.repoPath
			if location == "." {
				var err error
				location, err = os.Getwd()
				if err != nil {
					location = "."
				}
			}

			dest := filepath.Join(location, name)
			if _, err := os.Stat(dest); err == nil {
				m.repo.validationErr = fmt.Sprintf("directory already exists: %s", dest)
				return m, nil
			}

			m.repo.validationErr = ""
			opts := repo.CloneOptions{
				URL:      url,
				Location: location,
				Name:     name,
			}
			m.repo.events = nil
			m.repo.result = nil
			m.repo.opType = "clone"
			m.view = repoProgressView
			return m, m.startRepoPipeline(opts)
		}

		// Forward to focused input
		m.repo.validationErr = ""
		var cmd tea.Cmd
		if m.repo.cloneFocusedField == 0 {
			prevURL := m.repo.urlInput.Value()
			m.repo.urlInput, cmd = m.repo.urlInput.Update(msg)
			newURL := m.repo.urlInput.Value()
			// Auto-derive name when URL changes, unless user edited name field
			if newURL != prevURL && !m.repo.nameManuallyEdited {
				derived := repo.DeriveRepoName(newURL)
				m.repo.cloneNameInput.SetValue(derived)
			}
		} else {
			m.repo.nameManuallyEdited = true
			m.repo.cloneNameInput, cmd = m.repo.cloneNameInput.Update(msg)
		}
		return m, cmd
	}
	return m, nil
}

func (m Model) viewCloneInput() string {
	var b strings.Builder

	b.WriteString(styleTitle.Render(fmt.Sprintf("  sentei %s Clone Repository", "\u2500")))
	b.WriteString("\n\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")

	b.WriteString("  Repository URL\n")
	if m.repo.cloneFocusedField == 0 {
		b.WriteString("  " + m.repo.urlInput.View())
	} else {
		val := m.repo.urlInput.Value()
		if val == "" {
			val = styleDim.Render("(empty)")
		}
		b.WriteString("    " + val)
	}
	b.WriteString("\n\n")

	// Derive clone location for display
	location := m.repoPath
	if location == "." {
		if wd, err := os.Getwd(); err == nil {
			location = wd
		}
	}
	name := m.repo.cloneNameInput.Value()
	if name == "" {
		name = repo.DeriveRepoName(m.repo.urlInput.Value())
	}
	cloneDest := filepath.Join(location, name)

	b.WriteString("  Clone to\n")
	if m.repo.cloneFocusedField == 1 {
		b.WriteString("  " + m.repo.cloneNameInput.View())
		b.WriteString(styleDim.Render(fmt.Sprintf("  → %s", cloneDest)))
	} else {
		b.WriteString("    " + styleDim.Render(cloneDest))
	}
	b.WriteString("\n")

	if m.repo.validationErr != "" {
		b.WriteString("\n  " + styleError.Render(m.repo.validationErr))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")
	b.WriteString(styleDim.Render("  enter clone \u00b7 tab switch field \u00b7 esc back"))
	b.WriteString("\n")

	return b.String()
}
