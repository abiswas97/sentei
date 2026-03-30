package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestConfirmationView_ContainsTitle(t *testing.T) {
	vm := ConfirmationViewModel{
		Title: "Confirm Cleanup",
	}

	output := stripAnsi(vm.View())

	if !strings.Contains(output, "Confirm Cleanup") {
		t.Error("view should contain the title")
	}
}

func TestConfirmationView_ContainsAllItems(t *testing.T) {
	vm := ConfirmationViewModel{
		Title: "Review",
		Items: []ConfirmationItem{
			{Label: "Mode", Value: "aggressive"},
			{Label: "Force", Value: "true"},
			{Label: "Path", Value: "/home/user/repo"},
		},
	}

	output := stripAnsi(vm.View())

	for _, item := range vm.Items {
		if !strings.Contains(output, item.Label) {
			t.Errorf("view should contain label %q", item.Label)
		}
		if !strings.Contains(output, item.Value) {
			t.Errorf("view should contain value %q", item.Value)
		}
	}
}

func TestConfirmationView_ContainsCLICommand(t *testing.T) {
	vm := ConfirmationViewModel{
		Title:      "Confirm",
		CLICommand: "sentei cleanup --mode aggressive",
	}

	output := stripAnsi(vm.View())

	if !strings.Contains(output, "sentei cleanup --mode aggressive") {
		t.Error("view should contain the CLI command")
	}
}

func TestConfirmationView_ContainsKeybindings(t *testing.T) {
	vm := ConfirmationViewModel{Title: "Test"}

	output := stripAnsi(vm.View())

	if !strings.Contains(output, "enter confirm") {
		t.Error("view should show enter keybinding")
	}
	if !strings.Contains(output, "esc back") {
		t.Error("view should show esc keybinding")
	}
	if !strings.Contains(output, "q quit") {
		t.Error("view should show q keybinding")
	}
}

func TestConfirmationView_EmptyItems(t *testing.T) {
	vm := ConfirmationViewModel{
		Title:      "Empty",
		CLICommand: "sentei test",
	}

	output := stripAnsi(vm.View())

	if !strings.Contains(output, "Empty") {
		t.Error("view should still render title with no items")
	}
	if !strings.Contains(output, "sentei test") {
		t.Error("view should still render CLI command with no items")
	}
}

func TestUpdateConfirmation_EnterReturnsProceedMsg(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyEnter}

	cmd := UpdateConfirmation(msg)
	if cmd == nil {
		t.Fatal("enter should return a command")
	}

	result := cmd()
	if _, ok := result.(ConfirmProceedMsg); !ok {
		t.Errorf("enter should return ConfirmProceedMsg, got %T", result)
	}
}

func TestUpdateConfirmation_EscReturnsBackMsg(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyEscape}

	cmd := UpdateConfirmation(msg)
	if cmd == nil {
		t.Fatal("esc should return a command")
	}

	result := cmd()
	if _, ok := result.(ConfirmBackMsg); !ok {
		t.Errorf("esc should return ConfirmBackMsg, got %T", result)
	}
}

func TestUpdateConfirmation_QKeyReturnsQuit(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}

	cmd := UpdateConfirmation(msg)
	if cmd == nil {
		t.Fatal("q should return a command")
	}

	result := cmd()
	// tea.Quit returns a tea.QuitMsg
	if _, ok := result.(tea.QuitMsg); !ok {
		t.Errorf("q should return tea.QuitMsg, got %T", result)
	}
}

func TestUpdateConfirmation_CtrlCReturnsQuit(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}

	cmd := UpdateConfirmation(msg)
	if cmd == nil {
		t.Fatal("ctrl+c should return a command")
	}

	result := cmd()
	if _, ok := result.(tea.QuitMsg); !ok {
		t.Errorf("ctrl+c should return tea.QuitMsg, got %T", result)
	}
}

func TestUpdateConfirmation_UnhandledKeyReturnsNil(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}

	cmd := UpdateConfirmation(msg)
	if cmd != nil {
		t.Error("unhandled key should return nil")
	}
}

func TestUpdateConfirmation_NonKeyMsgReturnsNil(t *testing.T) {
	msg := tea.WindowSizeMsg{Width: 80, Height: 24}

	cmd := UpdateConfirmation(msg)
	if cmd != nil {
		t.Error("non-key message should return nil")
	}
}

func TestBuildCLICommand_NoFlags(t *testing.T) {
	result := BuildCLICommand("cleanup", nil)
	expected := "sentei cleanup"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestBuildCLICommand_EmptyFlags(t *testing.T) {
	result := BuildCLICommand("cleanup", map[string]string{})
	expected := "sentei cleanup"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestBuildCLICommand_WithFlags(t *testing.T) {
	result := BuildCLICommand("cleanup", map[string]string{
		"mode":  "aggressive",
		"force": "true",
	})
	expected := "sentei cleanup --force --mode aggressive"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestBuildCLICommand_FlagsSortedAlphabetically(t *testing.T) {
	result := BuildCLICommand("run", map[string]string{
		"zebra": "z",
		"alpha": "a",
		"mid":   "m",
	})
	expected := "sentei run --alpha a --mid m --zebra z"
	if result != expected {
		t.Errorf("flags should be sorted alphabetically: got %q, want %q", result, expected)
	}
}
