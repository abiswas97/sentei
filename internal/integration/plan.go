package integration

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/abiswas97/sentei/internal/progress"
)

// Step names are built here and nowhere else: the apply plan and the manager
// emit the same strings by construction.

// SetupStepName is the setup step for an integration; it always runs on
// enable, so plans declare it upfront.
func SetupStepName(integ Integration) string { return "Setup " + integ.Name }

// InstallStepName is the conditional install step (only when detection finds
// the tool missing); plans do not declare it.
func InstallStepName(integ Integration) string { return "Install " + integ.Name }

// InstallDependencyStepName is the conditional dependency-install step.
func InstallDependencyStepName(dep Dependency) string { return "Install dependency " + dep.Name }

// TeardownStepName is the teardown-command step on disable.
func TeardownStepName(integ Integration) string { return "Teardown " + integ.Name }

// RemoveDirStepName is the artifact-directory removal step on disable.
func RemoveDirStepName(dir, wtPath string) string {
	return fmt.Sprintf("Remove %s in %s", strings.TrimSuffix(dir, "/"), filepath.Base(wtPath))
}

// ApplyPlan declares the work known upfront for an apply: one Open phase per
// worktree (install steps appear dynamically when detection finds a tool
// missing) containing every setup and teardown step that will certainly run.
// The driver closes each phase as that worktree's work completes.
func ApplyPlan(toEnable, toDisable []Integration, wtPaths []string) progress.Plan {
	var phases []progress.PlannedPhase
	for _, wtPath := range wtPaths {
		var steps []progress.PlannedStep
		for _, integ := range toEnable {
			steps = append(steps, progress.PlannedStep{Name: SetupStepName(integ)})
		}
		for _, integ := range toDisable {
			if integ.Teardown.Command != "" {
				steps = append(steps, progress.PlannedStep{Name: TeardownStepName(integ)})
			}
			for _, dir := range integ.Teardown.Dirs {
				steps = append(steps, progress.PlannedStep{Name: RemoveDirStepName(dir, wtPath)})
			}
		}
		if len(steps) == 0 {
			continue // a worktree with no certain work declares nothing
		}
		phases = append(phases, progress.PlannedPhase{Name: wtPath, Steps: steps, Open: true})
	}
	return progress.Plan{Phases: phases}
}
