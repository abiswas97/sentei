package integration

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/abiswas97/sentei/internal/fileutil"
	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/progress"
)

const prerequisitesPhase progress.PhaseID = "prerequisites"

func SetupStepName(integ Integration) string          { return "Setup " + integ.Name }
func InstallStepName(integ Integration) string        { return "Install " + integ.Name }
func InstallDependencyStepName(dep Dependency) string { return "Install dependency " + dep.Name }
func TeardownStepName(integ Integration) string       { return "Teardown " + integ.Name }
func RemoveDirStepName(dir, wtPath string) string {
	return fmt.Sprintf("Remove %s in %s", strings.TrimSuffix(dir, "/"), filepath.Base(wtPath))
}

type applyOperationKind uint8

const (
	applyShellCommand applyOperationKind = iota
	applyFailure
	applyRemove
)

type applyOperation struct {
	phaseID   progress.PhaseID
	phaseName string
	stepID    progress.StepID
	label     string
	kind      applyOperationKind
	dir       string
	command   string
	failure   error
	dependsOn []string

	seedSource   string
	seedDest     string
	gitignore    []string
	gitignoreDir string
}

func (op applyOperation) key() string { return op.phaseID + "\x00" + op.stepID }

// PreparedApply is an exact progress declaration plus the frozen operations
// that implement it. Run performs no detection.
type PreparedApply struct {
	plan       progress.Plan
	operations []applyOperation
}

// Plan returns a defensive copy of the frozen progress declaration.
func (p PreparedApply) Plan() progress.Plan { return clonePlan(p.plan) }

// Empty reports whether preparation froze no work to execute.
func (p PreparedApply) Empty() bool { return len(p.operations) == 0 }

// PrepareApply performs the read-only detection pass once in the first target
// worktree, then freezes both the declaration and every execution decision.
func PrepareApply(shell git.ShellRunner, repoPath, mainWT string, toEnable, toDisable []Integration, wtPaths []string) (PreparedApply, error) {
	if (len(toEnable) > 0 || len(toDisable) > 0) && len(wtPaths) == 0 {
		return PreparedApply{}, errors.New("preparing integrations: no target worktree")
	}
	probeDir := ""
	if len(wtPaths) > 0 {
		probeDir = wtPaths[0]
	}

	type enabledDecision struct {
		integration Integration
		key         string
		installed   bool
		toolOpKey   string
	}
	decisions := make([]enabledDecision, 0, len(toEnable))
	dependencyPresent := map[string]bool{}
	dependencySpecs := map[string]Dependency{}
	dependencyOp := map[string]applyOperation{}
	var prerequisites []applyOperation

	for i, integ := range toEnable {
		decision := enabledDecision{integration: integ, key: fmt.Sprintf("%d:%s", i, integ.Name)}
		_, detectErr := shell.RunShell(probeDir, detectionCommand(integ))
		decision.installed = detectErr == nil
		for _, dep := range integ.Dependencies {
			if _, checked := dependencyPresent[dep.Name]; checked {
				continue
			}
			_, err := shell.RunShell(probeDir, dep.Detect)
			dependencyPresent[dep.Name] = err == nil
			dependencySpecs[dep.Name] = dep
		}
		decisions = append(decisions, decision)
	}

	for _, decision := range decisions {
		if decision.installed {
			continue
		}
		for _, dependency := range decision.integration.Dependencies {
			if dependencyPresent[dependency.Name] {
				continue
			}
			if _, planned := dependencyOp[dependency.Name]; planned {
				continue
			}
			dep := dependencySpecs[dependency.Name]
			op := applyOperation{
				phaseID: prerequisitesPhase, phaseName: "Prerequisites",
				stepID: stableStepID("dependency", dep.Name), label: InstallDependencyStepName(dep),
				dir: probeDir,
			}
			if dep.Install == "" {
				op.kind = applyFailure
				op.failure = fmt.Errorf("dependency %s is missing and has no install command", dep.Name)
			} else {
				op.kind = applyShellCommand
				op.command = dep.Install
			}
			dependencyOp[dep.Name] = op
			prerequisites = append(prerequisites, op)
		}
	}

	for i := range decisions {
		decision := &decisions[i]
		if decision.installed {
			continue
		}
		op := applyOperation{
			phaseID: prerequisitesPhase, phaseName: "Prerequisites",
			stepID: stableStepID("install", decision.key), label: InstallStepName(decision.integration),
			kind: applyShellCommand, dir: probeDir, command: decision.integration.Install.Command,
		}
		if op.command == "" {
			op.kind = applyFailure
			op.failure = fmt.Errorf("%s is missing and has no install command", decision.integration.Name)
		}
		for _, dep := range decision.integration.Dependencies {
			if depOp, missing := dependencyOp[dep.Name]; missing {
				op.dependsOn = append(op.dependsOn, depOp.key())
			}
		}
		decision.toolOpKey = op.key()
		prerequisites = append(prerequisites, op)
	}

	operations := append([]applyOperation(nil), prerequisites...)
	for wtIndex, wtPath := range wtPaths {
		phaseID := progress.PhaseID("worktree:" + stableToken(fmt.Sprintf("%d:%s", wtIndex, wtPath)))
		for _, decision := range decisions {
			integ := decision.integration
			workDir := wtPath
			if integ.Setup.WorkingDir == "repo" {
				workDir = repoPath
			}
			op := applyOperation{
				phaseID: phaseID, phaseName: wtPath,
				stepID: stableStepID("setup", decision.key), label: SetupStepName(integ),
				kind: applyShellCommand, dir: workDir,
				command:      strings.ReplaceAll(integ.Setup.Command, "{path}", git.ShellQuote(wtPath)),
				gitignore:    append([]string(nil), integ.GitignoreEntries...),
				gitignoreDir: wtPath,
			}
			if decision.toolOpKey != "" {
				op.dependsOn = append(op.dependsOn, decision.toolOpKey)
			}
			if integ.IndexCopyDir != "" && mainWT != "" && wtPath != mainWT {
				source := filepath.Join(mainWT, integ.IndexCopyDir)
				if _, err := os.Stat(source); err == nil {
					op.seedSource = source
					op.seedDest = filepath.Join(wtPath, integ.IndexCopyDir)
				}
			}
			operations = append(operations, op)
		}
		for disableIndex, integ := range toDisable {
			key := fmt.Sprintf("%d:%s", disableIndex, integ.Name)
			if integ.Teardown.Command != "" {
				operations = append(operations, applyOperation{
					phaseID: phaseID, phaseName: wtPath,
					stepID: stableStepID("teardown", key), label: TeardownStepName(integ),
					kind: applyShellCommand, dir: wtPath, command: integ.Teardown.Command,
				})
			}
			for dirIndex, dir := range integ.Teardown.Dirs {
				operations = append(operations, applyOperation{
					phaseID: phaseID, phaseName: wtPath,
					stepID: stableStepID("remove", fmt.Sprintf("%s:%d:%s", key, dirIndex, dir)), label: RemoveDirStepName(dir, wtPath),
					kind: applyRemove, dir: filepath.Join(wtPath, strings.TrimSuffix(dir, "/")),
				})
			}
		}
	}

	return PreparedApply{plan: planForOperations(operations), operations: operations}, nil
}

func (p PreparedApply) Run(shell git.ShellRunner, emit func(progress.Event)) ([]progress.Phase, error) {
	execution, err := progress.Start(p.plan, emit)
	if err != nil {
		return nil, fmt.Errorf("starting integration apply: %w", err)
	}
	results := make(map[string]progress.StepResult, len(p.operations))
	var runErr error
	for _, op := range p.operations {
		blockedBy := ""
		for _, dependency := range op.dependsOn {
			result := results[dependency]
			if result.Status == progress.StepFailed || result.Status == progress.StepSkipped {
				blockedBy = result.Name
				break
			}
		}
		if blockedBy != "" {
			result, err := execution.Skip(op.phaseID, op.stepID, "blocked by "+blockedBy)
			if err != nil {
				runErr = fmt.Errorf("skipping %s: %w", op.label, err)
				break
			}
			results[op.key()] = result
			continue
		}
		var result progress.StepResult
		var transitionErr error
		switch op.kind {
		case applyFailure:
			result, transitionErr = execution.Fail(op.phaseID, op.stepID, op.failure)
		case applyRemove:
			result, transitionErr = execution.Run(op.phaseID, op.stepID, func() (string, error) { return "", os.RemoveAll(op.dir) })
		default:
			result, transitionErr = execution.Run(op.phaseID, op.stepID, func() (string, error) {
				if op.seedSource != "" {
					_ = os.RemoveAll(op.seedDest)
					_ = fileutil.CopyDir(op.seedSource, op.seedDest)
				}
				message, err := shell.RunShell(op.dir, op.command)
				if err == nil {
					_ = appendGitignoreEntries(op.gitignoreDir, op.gitignore)
				}
				return message, err
			})
		}
		if transitionErr != nil {
			runErr = fmt.Errorf("executing %s: %w", op.label, transitionErr)
			break
		}
		results[op.key()] = result
	}
	finishErr := execution.Finish("integration apply finished")
	if finishErr != nil {
		finishErr = fmt.Errorf("finishing integration apply: %w", finishErr)
	}
	return phasesFromResults(p.plan, results), errors.Join(runErr, finishErr)
}

func detectionCommand(integ Integration) string {
	if integ.Detect.Command != "" {
		return integ.Detect.Command
	}
	if integ.Detect.BinaryName != "" {
		return "command -v " + integ.Detect.BinaryName
	}
	return "false"
}

func stableToken(value string) string {
	sum := sha256.Sum256([]byte(value))
	return fmt.Sprintf("%x", sum[:8])
}

func stableStepID(kind, value string) progress.StepID {
	return progress.StepID(kind + ":" + stableToken(value))
}

func clonePlan(plan progress.Plan) progress.Plan {
	clone := progress.Plan{Phases: make([]progress.PlannedPhase, len(plan.Phases))}
	copy(clone.Phases, plan.Phases)
	for i := range clone.Phases {
		clone.Phases[i].Steps = append([]progress.PlannedStep(nil), plan.Phases[i].Steps...)
	}
	return clone
}

func planForOperations(operations []applyOperation) progress.Plan {
	indices := map[progress.PhaseID]int{}
	var plan progress.Plan
	for _, op := range operations {
		index, exists := indices[op.phaseID]
		if !exists {
			index = len(plan.Phases)
			indices[op.phaseID] = index
			plan.Phases = append(plan.Phases, progress.PlannedPhase{ID: op.phaseID, Label: op.phaseName})
		}
		plan.Phases[index].Steps = append(plan.Phases[index].Steps, progress.PlannedStep{ID: op.stepID, Label: op.label})
	}
	return plan
}

func phasesFromResults(plan progress.Plan, results map[string]progress.StepResult) []progress.Phase {
	phases := make([]progress.Phase, 0, len(plan.Phases))
	for _, plannedPhase := range plan.Phases {
		phase := progress.Phase{Name: plannedPhase.Label}
		for _, step := range plannedPhase.Steps {
			phase.Steps = append(phase.Steps, results[plannedPhase.ID+"\x00"+step.ID])
		}
		phases = append(phases, phase)
	}
	return phases
}
