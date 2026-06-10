package cli

import (
	"errors"
	"strings"
	"testing"
)

func TestIsUnknownCommand_DispatchError(t *testing.T) {
	r := newTestRegistry()
	_, err := r.Dispatch([]string{"no-such-command"})
	if !IsUnknownCommand(err) {
		t.Errorf("expected IsUnknownCommand=true for %v", err)
	}
}

func TestIsUnknownCommand_OtherError(t *testing.T) {
	if IsUnknownCommand(errors.New("boom")) {
		t.Error("expected IsUnknownCommand=false for unrelated error")
	}
}

func TestIsUnknownCommand_NilError(t *testing.T) {
	if IsUnknownCommand(nil) {
		t.Error("expected IsUnknownCommand=false for nil")
	}
}

func TestUsageString_ListsCommandsWithTypeLabels(t *testing.T) {
	r := newTestRegistry()
	usage := r.UsageString()

	if !strings.Contains(usage, "Usage: sentei [command] [options]") {
		t.Errorf("expected usage header, got:\n%s", usage)
	}
	if !strings.Contains(usage, "ecosystems") || !strings.Contains(usage, "(output)") {
		t.Errorf("expected output command with label, got:\n%s", usage)
	}
	if !strings.Contains(usage, "cleanup") || !strings.Contains(usage, "(interactive)") {
		t.Errorf("expected decision command with label, got:\n%s", usage)
	}
}

func TestUsageString_SortsCommands(t *testing.T) {
	r := newTestRegistry()
	usage := r.UsageString()

	cleanupIdx := strings.Index(usage, "cleanup")
	createIdx := strings.Index(usage, "create")
	ecosystemsIdx := strings.Index(usage, "ecosystems")
	if cleanupIdx >= createIdx || createIdx >= ecosystemsIdx {
		t.Errorf("expected alphabetical command order, got:\n%s", usage)
	}
}

func TestBuildFlagString(t *testing.T) {
	tests := []struct {
		name  string
		base  string
		flags map[string]string
		want  string
	}{
		{"no flags", "sentei remove", nil, "sentei remove"},
		{"bool flag rendered without value", "sentei remove", map[string]string{"dry-run": "true"}, "sentei remove --dry-run"},
		{"value flag", "sentei cleanup", map[string]string{"mode": "safe"}, "sentei cleanup --mode safe"},
		{
			"flags sorted by key",
			"sentei remove",
			map[string]string{"stale": "30d", "all": "true", "merged": "true"},
			"sentei remove --all --merged --stale 30d",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildFlagString(tt.base, tt.flags); got != tt.want {
				t.Errorf("BuildFlagString(%q, %v) = %q, want %q", tt.base, tt.flags, got, tt.want)
			}
		})
	}
}
