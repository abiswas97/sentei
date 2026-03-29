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

type ghAuthStatusMsg struct {
	status string // "authenticated", "not authenticated", "gh not found"
}

func checkGitHubAuth(shell git.ShellRunner) tea.Cmd {
	return func() tea.Msg {
		_, err := shell.RunShell(".", "gh auth status")
		if err != nil {
			errStr := err.Error()
			if strings.Contains(errStr, "not found") || strings.Contains(errStr, "executable file not found") {
				return ghAuthStatusMsg{status: "gh not found"}
			}
			return ghAuthStatusMsg{status: "not authenticated"}
		}
		return ghAuthStatusMsg{status: "authenticated"}
	}
}

// repoOptionIndex constants for options cursor navigation.
const (
	repoOptWorktree    = 0 // always shown, display-only
	repoOptPublish     = 1 // toggle
	repoOptVisibility  = 2 // only shown when publish is on
	repoOptDescription = 3 // only shown when publish is on
)

func (m Model) repoVisibleOptions() []int {
	opts := []int{repoOptWorktree, repoOptPublish}
	if m.repo.publishGitHub {
		opts = append(opts, repoOptVisibility, repoOptDescription)
	}
	return opts
}

func (m Model) updateRepoOptions(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = max(msg.Height-6, 5)
		return m, nil

	case ghAuthStatusMsg:
		m.repo.ghStatus = msg.status
		return m, nil

	case tea.KeyMsg:
		visible := m.repoVisibleOptions()

		switch {
		case key.Matches(msg, keys.Back):
			m.view = repoNameView
			return m, m.repo.nameInput.Focus()

		case key.Matches(msg, keys.Down):
			// Find current index in visible list
			for i, opt := range visible {
				if opt == m.repo.optionsCursor {
					if i < len(visible)-1 {
						m.repo.optionsCursor = visible[i+1]
					}
					break
				}
			}

		case key.Matches(msg, keys.Up):
			for i, opt := range visible {
				if opt == m.repo.optionsCursor {
					if i > 0 {
						m.repo.optionsCursor = visible[i-1]
					}
					break
				}
			}

		case key.Matches(msg, keys.Toggle):
			switch m.repo.optionsCursor {
			case repoOptPublish:
				if m.repo.ghStatus == "authenticated" {
					m.repo.publishGitHub = !m.repo.publishGitHub
					if !m.repo.publishGitHub {
						// Reset cursor if it was on a now-hidden option
						m.repo.optionsCursor = repoOptPublish
					}
				}
			case repoOptVisibility:
				if m.repo.visibility == "private" {
					m.repo.visibility = "public"
				} else {
					m.repo.visibility = "private"
				}
			}

		case key.Matches(msg, keys.Confirm):
			name := strings.TrimSpace(m.repo.nameInput.Value())
			location := strings.TrimSpace(m.repo.locationInput.Value())

			opts := repo.CreateOptions{
				Name:          name,
				Location:      filepath.Dir(location),
				PublishGitHub: m.repo.publishGitHub,
				Visibility:    m.repo.visibility,
				Description:   strings.TrimSpace(m.repo.descInput.Value()),
			}

			m.repo.events = nil
			m.repo.result = nil
			m.repo.opType = "create"
			m.view = repoProgressView
			return m, m.startRepoPipeline(opts)

		default:
			// Forward to description input when focused on it
			if m.repo.optionsCursor == repoOptDescription && m.repo.publishGitHub {
				var cmd tea.Cmd
				m.repo.descInput, cmd = m.repo.descInput.Update(msg)
				return m, cmd
			}
		}
	}
	return m, nil
}

func (m Model) viewRepoOptions() string {
	var b strings.Builder

	name := m.repo.nameInput.Value()
	location := m.repo.locationInput.Value()

	b.WriteString(styleTitle.Render(fmt.Sprintf("  sentei %s Create Repository", "\u2500")))
	b.WriteString("\n\n")
	b.WriteString(styleDim.Render(fmt.Sprintf("  %s \u00b7 %s", name, location)))
	b.WriteString("\n\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")

	// Setup section
	b.WriteString("  " + styleTitle.Render("Setup"))
	b.WriteString("\n\n")

	// Worktree (display-only)
	cursor := "  "
	if m.repo.optionsCursor == repoOptWorktree {
		cursor = "> "
	}
	checkbox := styleCheckboxOn.Render("[x]")
	hint := "  " + styleDim.Render("main")
	if m.repo.optionsCursor == repoOptWorktree {
		b.WriteString(styleAccent.Render(cursor) + checkbox + " Create initial worktree" + hint)
	} else {
		b.WriteString("  " + checkbox + " Create initial worktree" + hint)
	}
	b.WriteString("\n\n")

	// GitHub section
	b.WriteString("  " + styleTitle.Render("GitHub"))

	// Auth status indicator
	var ghStatusStr string
	switch m.repo.ghStatus {
	case "authenticated":
		ghStatusStr = "  " + styleSuccess.Render("authenticated \u25cf")
	case "not authenticated":
		ghStatusStr = "  " + styleError.Render("not authenticated \u2717")
	case "gh not found":
		ghStatusStr = "  " + styleError.Render("gh not found \u2717")
	default:
		ghStatusStr = "  " + styleDim.Render("checking\u2026")
	}
	b.WriteString(ghStatusStr)
	b.WriteString("\n\n")

	// Publish toggle
	cursor = "  "
	if m.repo.optionsCursor == repoOptPublish {
		cursor = "> "
	}
	if m.repo.publishGitHub {
		checkbox = styleCheckboxOn.Render("[x]")
	} else {
		checkbox = styleCheckboxOff.Render("[ ]")
	}
	publishLabel := "Publish to GitHub"
	if m.repo.ghStatus != "authenticated" {
		publishLabel = styleDim.Render(publishLabel)
	}
	if m.repo.optionsCursor == repoOptPublish {
		b.WriteString(styleAccent.Render(cursor) + checkbox + " " + publishLabel)
	} else {
		b.WriteString("  " + checkbox + " " + publishLabel)
	}
	b.WriteString("\n")

	// Progressive disclosure — only when publishing
	if m.repo.publishGitHub {
		// Visibility
		cursor = "  "
		if m.repo.optionsCursor == repoOptVisibility {
			cursor = "> "
		}
		visVal := styleDim.Render(m.repo.visibility)
		if m.repo.optionsCursor == repoOptVisibility {
			b.WriteString(styleAccent.Render(cursor) + fmt.Sprintf("      %-15s %s", "Visibility", visVal))
		} else {
			b.WriteString(fmt.Sprintf("        %-15s %s", "Visibility", visVal))
		}
		b.WriteString("\n")

		// Description
		cursor = "  "
		if m.repo.optionsCursor == repoOptDescription {
			cursor = "> "
		}
		if m.repo.optionsCursor == repoOptDescription {
			b.WriteString(styleAccent.Render(cursor) + fmt.Sprintf("      %-15s ", "Description") + m.repo.descInput.View())
		} else {
			val := m.repo.descInput.Value()
			if val == "" {
				val = styleDim.Render("(optional)")
			}
			b.WriteString(fmt.Sprintf("        %-15s %s", "Description", val))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")
	b.WriteString(styleDim.Render("  space toggle \u00b7 enter create \u00b7 esc back"))
	b.WriteString("\n")

	return b.String()
}
