package cmd

import (
	"flag"
	"fmt"

	"github.com/abiswas97/sentei/internal/cleanup"
	"github.com/abiswas97/sentei/internal/cli"
)

// ParseCleanupFlags parses cleanup-specific flags and returns CleanupOptions.
// Returns an error if validation fails (e.g., invalid mode).
func ParseCleanupFlags(args []string) (*cleanup.Options, error) {
	fs := flag.NewFlagSet("cleanup", flag.ContinueOnError)
	mode := fs.String("mode", "", "Cleanup mode: safe or aggressive")
	dryRun := fs.Bool("dry-run", false, "Show what would be done without making changes")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	opts := &cleanup.Options{
		DryRun: *dryRun,
	}

	if *mode != "" {
		m := cleanup.Mode(*mode)
		if m != cleanup.ModeSafe && m != cleanup.ModeAggressive {
			return nil, fmt.Errorf("invalid value for --mode: must be 'safe' or 'aggressive'")
		}
		opts.Mode = m
	}

	return opts, nil
}

// ValidateCleanupForNonInteractive checks that all required flags are present
// for non-interactive execution.
func ValidateCleanupForNonInteractive(opts *cleanup.Options) error {
	if opts.Mode == "" {
		return fmt.Errorf("missing required flag: --mode (safe|aggressive)")
	}
	return nil
}

// CleanupCLICommand generates the equivalent CLI command string from options.
func CleanupCLICommand(opts *cleanup.Options) string {
	flags := make(map[string]string)
	if opts.Mode != "" {
		flags["mode"] = string(opts.Mode)
	}
	if opts.DryRun {
		flags["dry-run"] = "true"
	}
	return buildFlagString("sentei cleanup", flags)
}

func buildFlagString(base string, flags map[string]string) string {
	return cli.BuildFlagString(base, flags)
}
