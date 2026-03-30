package cli

import (
	"errors"
	"flag"
	"fmt"
	"strings"
)

// CommandType classifies commands as output (read-only, always CLI) or decision
// (requires user choices, defaults to TUI).
type CommandType int

const (
	Output   CommandType = iota // Read-only commands that print to stdout and exit.
	Decision                    // Commands that require choices; default to TUI.
)

// Command defines a registered CLI command with its type and handlers.
type Command struct {
	Name        string
	Type        CommandType
	Destructive bool // When true, --non-interactive requires --force.

	// RunCLI executes the command in non-interactive mode.
	// For output commands, this is the only execution path.
	// For decision commands, this runs when --non-interactive is provided.
	RunCLI func(args []string) error

	// ParseFlags parses command-specific flags into an opaque options value.
	// Only used for decision commands; nil for output commands.
	ParseFlags func(args []string) (any, error)

	// BuildTUI configures the TUI model for this command's flow.
	// Receives the parsed options (may be nil if no flags provided).
	// Only used for decision commands; nil for output commands.
	BuildTUI func(opts any) error
}

// Registry holds registered commands and dispatches based on os.Args.
type Registry struct {
	commands map[string]*Command
}

// NewRegistry creates an empty command registry.
func NewRegistry() *Registry {
	return &Registry{commands: make(map[string]*Command)}
}

// Register adds a command to the registry.
func (r *Registry) Register(cmd *Command) {
	r.commands[cmd.Name] = cmd
}

// DispatchResult tells the caller what action to take after dispatch.
type DispatchResult struct {
	// IsRoot is true when no command was provided (launch TUI menu).
	IsRoot bool

	// Command is the matched command, if any.
	Command *Command

	// Args are the remaining arguments after the command name.
	Args []string

	// NonInteractive is true when --non-interactive was provided.
	NonInteractive bool

	// Force is true when --force was provided.
	Force bool
}

var (
	ErrUnknownCommand           = errors.New("unknown command")
	ErrMissingForce             = errors.New("destructive operation requires --force with --non-interactive")
	ErrNonInteractiveNotAllowed = errors.New("--non-interactive is not supported for output commands")
)

// IsUnknownCommand returns true if the error wraps ErrUnknownCommand.
func IsUnknownCommand(err error) bool {
	return errors.Is(err, ErrUnknownCommand)
}

// Dispatch parses the command name from args and returns a DispatchResult.
// It extracts --non-interactive and --force from the args before returning.
func (r *Registry) Dispatch(args []string) (*DispatchResult, error) {
	if len(args) == 0 {
		return &DispatchResult{IsRoot: true}, nil
	}

	name := args[0]

	// Check if the first arg looks like a flag (not a command).
	if strings.HasPrefix(name, "-") {
		return &DispatchResult{IsRoot: true, Args: args}, nil
	}

	cmd, ok := r.commands[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownCommand, name)
	}

	remaining := args[1:]
	nonInteractive, force, remaining := extractGlobalFlags(remaining)

	result := &DispatchResult{
		Command:        cmd,
		Args:           remaining,
		NonInteractive: nonInteractive,
		Force:          force,
	}

	// Validate flag combinations.
	if nonInteractive && cmd.Destructive && !force {
		return nil, ErrMissingForce
	}

	return result, nil
}

// Lookup returns a command by name, or nil if not found.
func (r *Registry) Lookup(name string) *Command {
	return r.commands[name]
}

// CommandNames returns all registered command names.
func (r *Registry) CommandNames() []string {
	names := make([]string, 0, len(r.commands))
	for name := range r.commands {
		names = append(names, name)
	}
	return names
}

// UsageString returns a formatted help string listing all commands.
func (r *Registry) UsageString() string {
	var b strings.Builder
	b.WriteString("Usage: sentei [command] [options]\n\n")
	b.WriteString("Commands:\n")

	for _, cmd := range r.commands {
		label := "output"
		if cmd.Type == Decision {
			label = "interactive"
		}
		fmt.Fprintf(&b, "  %-14s (%s)\n", cmd.Name, label)
	}

	b.WriteString("\nRun 'sentei <command> --help' for command-specific options.\n")
	return b.String()
}

// extractGlobalFlags pulls --non-interactive and --force from the args slice,
// returning the flag values and the remaining args. This uses a simple scan
// rather than flag.FlagSet to avoid conflicting with command-specific flags.
func extractGlobalFlags(args []string) (nonInteractive bool, force bool, remaining []string) {
	fs := flag.NewFlagSet("global", flag.ContinueOnError)
	fs.BoolVar(&nonInteractive, "non-interactive", false, "")
	fs.BoolVar(&force, "force", false, "")

	// We need to extract our flags but leave command-specific flags untouched.
	// Scan args manually since flag.FlagSet doesn't support mixed parsing well.
	for _, arg := range args {
		switch arg {
		case "--non-interactive":
			nonInteractive = true
		case "--force":
			force = true
		default:
			remaining = append(remaining, arg)
		}
	}
	return
}
