package cli

import (
	"errors"
	"testing"
)

func newTestRegistry() *Registry {
	r := NewRegistry()

	r.Register(&Command{
		Name: "ecosystems",
		Type: Output,
		RunCLI: func(args []string) error {
			return nil
		},
	})

	r.Register(&Command{
		Name:        "cleanup",
		Type:        Decision,
		Destructive: true,
		RunCLI: func(args []string) error {
			return nil
		},
	})

	r.Register(&Command{
		Name: "create",
		Type: Decision,
		RunCLI: func(args []string) error {
			return nil
		},
	})

	return r
}

func TestDispatch_Root(t *testing.T) {
	r := newTestRegistry()

	result, err := r.Dispatch(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsRoot {
		t.Error("expected IsRoot=true for nil args")
	}

	result, err = r.Dispatch([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsRoot {
		t.Error("expected IsRoot=true for empty args")
	}
}

func TestDispatch_RootWithFlags(t *testing.T) {
	r := newTestRegistry()

	result, err := r.Dispatch([]string{"--version"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsRoot {
		t.Error("expected IsRoot=true when first arg is a flag")
	}
	if len(result.Args) != 1 || result.Args[0] != "--version" {
		t.Errorf("expected args=[--version], got %v", result.Args)
	}
}

func TestDispatch_KnownOutputCommand(t *testing.T) {
	r := newTestRegistry()

	result, err := r.Dispatch([]string{"ecosystems"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsRoot {
		t.Error("expected IsRoot=false for known command")
	}
	if result.Command.Name != "ecosystems" {
		t.Errorf("expected command=ecosystems, got %s", result.Command.Name)
	}
	if result.Command.Type != Output {
		t.Error("expected Output type")
	}
}

func TestDispatch_KnownDecisionCommand(t *testing.T) {
	r := newTestRegistry()

	result, err := r.Dispatch([]string{"create", "--branch", "foo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Command.Name != "create" {
		t.Errorf("expected command=create, got %s", result.Command.Name)
	}
	if result.Command.Type != Decision {
		t.Error("expected Decision type")
	}
	if len(result.Args) != 2 || result.Args[0] != "--branch" || result.Args[1] != "foo" {
		t.Errorf("expected remaining args=[--branch foo], got %v", result.Args)
	}
}

func TestDispatch_UnknownCommand(t *testing.T) {
	r := newTestRegistry()

	_, err := r.Dispatch([]string{"foobar"})
	if !errors.Is(err, ErrUnknownCommand) {
		t.Errorf("expected ErrUnknownCommand, got %v", err)
	}
}

func TestDispatch_NonInteractiveFlag(t *testing.T) {
	r := newTestRegistry()

	result, err := r.Dispatch([]string{"create", "--non-interactive", "--branch", "foo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.NonInteractive {
		t.Error("expected NonInteractive=true")
	}
	if len(result.Args) != 2 || result.Args[0] != "--branch" || result.Args[1] != "foo" {
		t.Errorf("--non-interactive should be stripped from args, got %v", result.Args)
	}
}

func TestDispatch_ForceFlag(t *testing.T) {
	r := newTestRegistry()

	result, err := r.Dispatch([]string{"cleanup", "--force", "--non-interactive"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Force {
		t.Error("expected Force=true")
	}
	if !result.NonInteractive {
		t.Error("expected NonInteractive=true")
	}
}

func TestDispatch_DestructiveWithoutForce(t *testing.T) {
	r := newTestRegistry()

	_, err := r.Dispatch([]string{"cleanup", "--non-interactive"})
	if !errors.Is(err, ErrMissingForce) {
		t.Errorf("expected ErrMissingForce, got %v", err)
	}
}

func TestDispatch_DestructiveWithForce(t *testing.T) {
	r := newTestRegistry()

	result, err := r.Dispatch([]string{"cleanup", "--non-interactive", "--force"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.NonInteractive || !result.Force {
		t.Error("expected both NonInteractive and Force to be true")
	}
}

func TestDispatch_DestructiveInteractiveWithoutForce(t *testing.T) {
	r := newTestRegistry()

	// Without --non-interactive, --force requirement doesn't apply.
	result, err := r.Dispatch([]string{"cleanup"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NonInteractive {
		t.Error("expected NonInteractive=false")
	}
}

func TestDispatch_NonDestructiveNonInteractive(t *testing.T) {
	r := newTestRegistry()

	// create is not destructive, so --non-interactive alone is fine.
	result, err := r.Dispatch([]string{"create", "--non-interactive"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.NonInteractive {
		t.Error("expected NonInteractive=true")
	}
}

func TestLookup(t *testing.T) {
	r := newTestRegistry()

	if cmd := r.Lookup("ecosystems"); cmd == nil {
		t.Error("expected to find ecosystems")
	}
	if cmd := r.Lookup("nope"); cmd != nil {
		t.Error("expected nil for unknown command")
	}
}

func TestCommandNames(t *testing.T) {
	r := newTestRegistry()

	names := r.CommandNames()
	if len(names) != 3 {
		t.Errorf("expected 3 names, got %d", len(names))
	}

	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}
	for _, want := range []string{"ecosystems", "cleanup", "create"} {
		if !nameSet[want] {
			t.Errorf("expected %q in command names", want)
		}
	}
}
