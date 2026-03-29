package creator

import "github.com/abiswas97/sentei/internal/git"

func runDeps(runner git.CommandRunner, wtPath string, opts Options, emit func(Event)) Phase {
	return Phase{Name: "Dependencies"}
}
