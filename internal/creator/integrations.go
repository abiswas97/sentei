package creator

import "github.com/abiswas97/sentei/internal/git"

func runIntegrations(runner git.CommandRunner, wtPath string, opts Options, emit func(Event)) Phase {
	return Phase{Name: "Integrations"}
}
