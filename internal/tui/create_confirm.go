package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/cmd"
)

// SetCreateOpts sets the create options and starts at the appropriate view.
// If branch and base are both set, starts at createConfirmView.
// If only branch is set, starts at createOptionsView (base selection done).
// If nothing is set, starts at createBranchView (normal flow).
func (m *Model) SetCreateOpts(opts *cmd.CreateOptions) {
	m.createOpts = opts

	if opts.Branch != "" {
		m.create.branchInput.SetValue(opts.Branch)
	}
	if opts.Base != "" {
		m.create.baseInput.SetValue(opts.Base)
	}
	if opts.MergeBase {
		m.create.mergeBase = true
	}
	if opts.CopyEnv {
		m.create.copyEnvFiles = true
	}
	if len(opts.Ecosystems) > 0 {
		for _, eco := range opts.Ecosystems {
			m.create.ecoEnabled[eco] = true
		}
	}

	switch {
	case opts.Branch != "" && opts.Base != "":
		m.view = createConfirmView
	case opts.Branch != "":
		m.prepareCreateOptions()
		m.view = createOptionsView
	default:
		m.view = createBranchView
	}
}

// createConfirmationVM builds the ConfirmationViewModel for the create flow.
func (m Model) createConfirmationVM() ConfirmationViewModel {
	branch := m.create.branchInput.Value()
	base := m.create.baseInput.Value()

	items := []ConfirmationItem{
		{Label: "Branch:", Value: branch},
		{Label: "Base:", Value: base},
	}

	var enabledEcos []string
	for _, eco := range m.create.ecosystems {
		if m.create.ecoEnabled[eco.Name] {
			enabledEcos = append(enabledEcos, eco.Name)
		}
	}
	if len(enabledEcos) > 0 {
		items = append(items, ConfirmationItem{
			Label: "Ecosystems:",
			Value: strings.Join(enabledEcos, ", "),
		})
	}

	mergeBase := "no"
	if m.create.mergeBase {
		mergeBase = "yes"
	}
	items = append(items, ConfirmationItem{Label: "Merge base:", Value: mergeBase})

	copyEnv := "no"
	if m.create.copyEnvFiles {
		copyEnv = "yes"
	}
	items = append(items, ConfirmationItem{Label: "Copy env:", Value: copyEnv})

	// Build CLI command from options.
	opts := m.resolvedCreateOpts()
	cliCmd := cmd.CreateCLICommand(opts)

	return ConfirmationViewModel{
		Title:      "Confirm Create",
		Items:      items,
		CLICommand: cliCmd,
	}
}

// resolvedCreateOpts returns the effective CreateOptions from the current model state.
func (m Model) resolvedCreateOpts() *cmd.CreateOptions {
	opts := &cmd.CreateOptions{
		Branch:    m.create.branchInput.Value(),
		Base:      m.create.baseInput.Value(),
		MergeBase: m.create.mergeBase,
		CopyEnv:   m.create.copyEnvFiles,
	}
	for _, eco := range m.create.ecosystems {
		if m.create.ecoEnabled[eco.Name] {
			opts.Ecosystems = append(opts.Ecosystems, eco.Name)
		}
	}
	return opts
}

func (m Model) updateCreateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = max(msg.Height-6, 5)
		return m, nil

	case ConfirmProceedMsg:
		m.startCreation()
		m.view = createProgressView
		return m, m.waitForCreateEvent()

	case ConfirmBackMsg:
		if m.createOpts != nil {
			return m, tea.Quit
		}
		m.view = createOptionsView
		return m, nil
	}

	if cmd := UpdateConfirmation(msg); cmd != nil {
		return m, cmd
	}

	return m, nil
}

func (m Model) viewCreateConfirm() string {
	return m.createConfirmationVM().View()
}
