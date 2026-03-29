package creator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/integration"
)

// ArtifactInfo describes the artifact directories found for an integration.
type ArtifactInfo struct {
	IntegrationName string
	Dirs            []string
}

// ScanArtifacts checks which integrations have artifact directories present in wtPath.
func ScanArtifacts(wtPath string, integrations []integration.Integration) []ArtifactInfo {
	var found []ArtifactInfo

	for _, integ := range integrations {
		var presentDirs []string
		for _, dir := range integ.Teardown.Dirs {
			cleanDir := strings.TrimSuffix(dir, "/")
			fullPath := filepath.Join(wtPath, cleanDir)
			if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
				presentDirs = append(presentDirs, dir)
			}
		}
		if len(presentDirs) > 0 {
			found = append(found, ArtifactInfo{
				IntegrationName: integ.Name,
				Dirs:            presentDirs,
			})
		}
	}

	return found
}

// Teardown removes integration artifacts from wtPath. For each integration that has
// artifacts present, it runs the teardown command if configured; if the command fails
// or is absent, it falls back to deleting the artifact directories directly.
func Teardown(shell git.ShellRunner, wtPath string, integrations []integration.Integration, emit func(Event)) []StepResult {
	artifacts := ScanArtifacts(wtPath, integrations)
	if len(artifacts) == 0 {
		return nil
	}

	var results []StepResult

	for _, artifact := range artifacts {
		integ := findIntegration(integrations, artifact.IntegrationName)
		if integ == nil {
			continue
		}

		stepName := fmt.Sprintf("Teardown %s", integ.Name)
		emit(Event{Phase: "Teardown", Step: stepName, Status: StepRunning})

		if integ.Teardown.Command != "" {
			_, err := shell.RunShell(wtPath, integ.Teardown.Command)
			if err == nil {
				emit(Event{Phase: "Teardown", Step: stepName, Status: StepDone})
				results = append(results, StepResult{Name: stepName, Status: StepDone})
				continue
			}
		}

		allRemoved := true
		for _, dir := range artifact.Dirs {
			cleanDir := strings.TrimSuffix(dir, "/")
			fullPath := filepath.Join(wtPath, cleanDir)
			if err := os.RemoveAll(fullPath); err != nil {
				allRemoved = false
			}
		}

		if allRemoved {
			emit(Event{Phase: "Teardown", Step: stepName, Status: StepDone, Message: "removed artifact dirs"})
			results = append(results, StepResult{Name: stepName, Status: StepDone, Message: "removed artifact dirs"})
		} else {
			emit(Event{Phase: "Teardown", Step: stepName, Status: StepFailed, Error: fmt.Errorf("failed to remove some artifact dirs")})
			results = append(results, StepResult{
				Name:   stepName,
				Status: StepFailed,
				Error:  fmt.Errorf("failed to remove some artifact dirs for %s", integ.Name),
			})
		}
	}

	return results
}

func findIntegration(integrations []integration.Integration, name string) *integration.Integration {
	for i := range integrations {
		if integrations[i].Name == name {
			return &integrations[i]
		}
	}
	return nil
}
