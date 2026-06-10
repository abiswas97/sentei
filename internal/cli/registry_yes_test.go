package cli

import (
	"errors"
	"testing"
)

func yesTestRegistry() *Registry {
	r := NewRegistry()
	r.Register(&Command{Name: "cleanup", Type: Decision, Destructive: true})
	r.Register(&Command{Name: "remove", Type: Decision, Destructive: true})
	return r
}

func TestDispatch_YesFlagExtracted(t *testing.T) {
	r := yesTestRegistry()
	for _, flag := range []string{"--yes", "-y"} {
		result, err := r.Dispatch([]string{"cleanup", flag, "--mode", "safe"})
		if err != nil {
			t.Fatalf("Dispatch(%s) error: %v", flag, err)
		}
		if !result.Yes {
			t.Errorf("%s must set Yes", flag)
		}
		for _, a := range result.Args {
			if a == flag {
				t.Errorf("%s must be consumed, found in remaining args %v", flag, result.Args)
			}
		}
	}
}

func TestDispatch_YesDoesNotRequireForce(t *testing.T) {
	r := yesTestRegistry()
	if _, err := r.Dispatch([]string{"cleanup", "--yes"}); err != nil {
		t.Errorf("--yes alone must dispatch destructive commands (command safeties apply): %v", err)
	}
	// --non-interactive keeps its stricter contract.
	if _, err := r.Dispatch([]string{"cleanup", "--non-interactive"}); !errors.Is(err, ErrMissingForce) {
		t.Errorf("--non-interactive without --force must still be rejected, got %v", err)
	}
}
