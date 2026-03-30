package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/cmd"
	"github.com/abiswas97/sentei/internal/repo"
)

// SetCloneOpts sets the clone options and starts at the appropriate view.
// If URL is set, starts at cloneConfirmView.
// If nothing is set, starts at cloneInputView (normal flow).
func (m *Model) SetCloneOpts(opts *cmd.CloneOptions) {
	m.cloneOpts = opts

	if opts.URL != "" {
		m.repo.urlInput.SetValue(opts.URL)
	}
	if opts.Name != "" {
		m.repo.cloneNameInput.SetValue(opts.Name)
		m.repo.nameManuallyEdited = true
	}

	if opts.URL != "" {
		m.view = cloneConfirmView
	} else {
		m.view = cloneInputView
	}
}

// cloneConfirmationVM builds the ConfirmationViewModel for the clone flow.
func (m Model) cloneConfirmationVM() ConfirmationViewModel {
	url := m.repo.urlInput.Value()
	name := m.repo.cloneNameInput.Value()

	items := []ConfirmationItem{
		{Label: "URL:", Value: url},
	}
	if name != "" {
		items = append(items, ConfirmationItem{Label: "Name:", Value: name})
	}

	opts := &cmd.CloneOptions{
		URL:  url,
		Name: name,
	}
	cliCmd := cmd.CloneCLICommand(opts)

	return ConfirmationViewModel{
		Title:      "Confirm Clone",
		Items:      items,
		CLICommand: cliCmd,
	}
}

func (m Model) updateCloneConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = max(msg.Height-6, 5)
		return m, nil

	case ConfirmProceedMsg:
		url := m.repo.urlInput.Value()
		name := m.repo.cloneNameInput.Value()

		m.repo.validationErr = ""
		m.repo.events = nil
		m.repo.result = nil
		m.repo.opType = "clone"
		m.view = repoProgressView

		cloneOpts := repo.CloneOptions{
			URL:      url,
			Location: m.repoPath,
			Name:     name,
		}
		return m, m.startRepoPipeline(cloneOpts)

	case ConfirmBackMsg:
		if m.cloneOpts != nil {
			return m, tea.Quit
		}
		m.view = cloneInputView
		return m, nil
	}

	if cmd := UpdateConfirmation(msg); cmd != nil {
		return m, cmd
	}

	return m, nil
}

func (m Model) viewCloneConfirm() string {
	return m.cloneConfirmationVM().View()
}
