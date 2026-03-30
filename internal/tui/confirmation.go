package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// ConfirmProceedMsg is sent when the user presses Enter to confirm.
type ConfirmProceedMsg struct{}

// ConfirmBackMsg is sent when the user presses Esc to go back.
type ConfirmBackMsg struct{}

// ConfirmationItem represents a key-value row in the confirmation view.
type ConfirmationItem struct {
	Label string
	Value string
}

// ConfirmationViewModel holds the state for a reusable confirmation view.
type ConfirmationViewModel struct {
	Title      string
	Items      []ConfirmationItem
	CLICommand string
}

// UpdateConfirmation handles key messages for the confirmation view and returns
// the appropriate command. Returns nil if the message was not handled.
func UpdateConfirmation(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Confirm):
			return func() tea.Msg { return ConfirmProceedMsg{} }
		case key.Matches(msg, keys.Back):
			return func() tea.Msg { return ConfirmBackMsg{} }
		case key.Matches(msg, keys.Quit):
			return tea.Quit
		}
	}
	return nil
}

// View renders the confirmation view.
func (c ConfirmationViewModel) View() string {
	var b strings.Builder

	b.WriteString(styleTitle.Render("  sentei ─ " + c.Title))
	b.WriteString("\n\n")

	if len(c.Items) > 0 {
		labelWidth := 0
		for _, item := range c.Items {
			if len(item.Label) > labelWidth {
				labelWidth = len(item.Label)
			}
		}

		for _, item := range c.Items {
			fmt.Fprintf(&b, "  %-*s  %s\n", labelWidth, item.Label, item.Value)
		}

		b.WriteString("\n")
		b.WriteString(separator(50))
		b.WriteString("\n\n")
	}

	if c.CLICommand != "" {
		b.WriteString(styleDim.Render("  "+c.CLICommand) + "\n\n")
	}

	b.WriteString(styleDim.Render("  enter confirm  •  esc back  •  q quit") + "\n")

	return styleDialogBox.Render(b.String())
}

// BuildCLICommand constructs a CLI command string from a command name and flags.
// Flags are sorted by key for deterministic output.
func BuildCLICommand(command string, flags map[string]string) string {
	if len(flags) == 0 {
		return "sentei " + command
	}

	keys := make([]string, 0, len(flags))
	for k := range flags {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	parts = append(parts, "sentei", command)
	for _, k := range keys {
		v := flags[k]
		if v == "true" {
			parts = append(parts, "--"+k)
		} else {
			parts = append(parts, "--"+k, v)
		}
	}

	return strings.Join(parts, " ")
}
