package repo

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type mockGhRunner struct {
	responses map[string]mockResponse
}

func (m *mockGhRunner) RunGh(dir string, args ...string) (string, error) {
	key := fmt.Sprintf("%s:gh[%s]", dir, strings.Join(args, " "))
	if resp, ok := m.responses[key]; ok {
		return resp.output, resp.err
	}
	return "", fmt.Errorf("unexpected gh call: %s", key)
}

func TestCreate_LocalOnly(t *testing.T) {
	dir := t.TempDir()
	repoName := "my-project"
	repoPath := filepath.Join(dir, repoName)

	runner := &mockRunner{responses: map[string]mockResponse{
		fmt.Sprintf("%s/.bare:[init --bare]", repoPath):                                                    {output: ""},
		fmt.Sprintf("%s/.bare:[config remote.origin.fetch +refs/heads/*:refs/remotes/origin/*]", repoPath): {output: ""},
		fmt.Sprintf("%s:[worktree add %s/main -b main]", repoPath, repoPath):                               {output: ""},
		fmt.Sprintf("%s/main:[add -A]", repoPath):                                                          {output: ""},
		fmt.Sprintf("%s/main:[commit -m Initial commit]", repoPath):                                        {output: ""},
	}}

	ec := &eventCollector{}
	opts := CreateOptions{
		Name:          repoName,
		Location:      dir,
		PublishGitHub: false,
	}
	result := Create(runner, runner, opts, ec.emit)

	if result.RepoPath != repoPath {
		t.Errorf("RepoPath = %q, want %q", result.RepoPath, repoPath)
	}
	if len(result.Phases) != 1 {
		t.Errorf("want 1 phase (setup only), got %d", len(result.Phases))
	}
	for _, phase := range result.Phases {
		for _, step := range phase.Steps {
			if step.Status == StepFailed {
				t.Errorf("step %q failed: %v", step.Name, step.Error)
			}
		}
	}
	if len(ec.events) == 0 {
		t.Error("expected events to be emitted")
	}
}

func TestCreate_WithGitHub(t *testing.T) {
	dir := t.TempDir()
	repoName := "my-project"
	repoPath := filepath.Join(dir, repoName)

	runner := &mockRunner{responses: map[string]mockResponse{
		// Setup phase
		fmt.Sprintf("%s/.bare:[init --bare]", repoPath):                                                    {output: ""},
		fmt.Sprintf("%s/.bare:[config remote.origin.fetch +refs/heads/*:refs/remotes/origin/*]", repoPath): {output: ""},
		fmt.Sprintf("%s:[worktree add %s/main -b main]", repoPath, repoPath):                               {output: ""},
		fmt.Sprintf("%s/main:[add -A]", repoPath):                                                          {output: ""},
		fmt.Sprintf("%s/main:[commit -m Initial commit]", repoPath):                                        {output: ""},
		// GitHub phase (git commands) — gh git_protocol is https (gh default).
		fmt.Sprintf("%s/.bare:[remote set-url origin https://github.com/abiswas97/my-project.git]", repoPath): {output: ""},
		fmt.Sprintf("%s/main:[push -u origin main]", repoPath):                                                {output: ""},
		fmt.Sprintf("%s/.bare:[remote set-head origin main]", repoPath):                                       {output: ""},
	}}
	ghRunner := &mockGhRunner{responses: map[string]mockResponse{
		fmt.Sprintf("%s:gh[api user --jq .login]", repoPath):             {output: "abiswas97"},
		fmt.Sprintf("%s:gh[repo create my-project --private]", repoPath): {output: ""},
		fmt.Sprintf("%s:gh[config get git_protocol]", repoPath):          {output: "https"},
	}}

	ec := &eventCollector{}
	opts := CreateOptions{
		Name:          repoName,
		Location:      dir,
		PublishGitHub: true,
		Visibility:    "private",
		Description:   "",
	}
	result := CreateWithGh(runner, runner, ghRunner, opts, ec.emit)

	if len(result.Phases) != 2 {
		t.Errorf("want 2 phases (setup + github), got %d", len(result.Phases))
	}
	for _, phase := range result.Phases {
		for _, step := range phase.Steps {
			if step.Status == StepFailed {
				t.Errorf("step %q failed: %v", step.Name, step.Error)
			}
		}
	}
}

func TestGhRemoteURL_RespectsConfiguredProtocol(t *testing.T) {
	sshGh := &mockGhRunner{responses: map[string]mockResponse{
		"/r:gh[config get git_protocol]": {output: "ssh"},
	}}
	if got := ghRemoteURL(sshGh, "/r", "u", "n"); got != "git@github.com:u/n.git" {
		t.Errorf("ssh protocol: got %q", got)
	}

	httpsGh := &mockGhRunner{responses: map[string]mockResponse{
		"/r:gh[config get git_protocol]": {output: "https"},
	}}
	if got := ghRemoteURL(httpsGh, "/r", "u", "n"); got != "https://github.com/u/n.git" {
		t.Errorf("https protocol: got %q", got)
	}

	// On error, default to HTTPS — gh's own default and what token auth uses.
	brokenGh := &mockGhRunner{responses: map[string]mockResponse{}}
	if got := ghRemoteURL(brokenGh, "/r", "u", "n"); got != "https://github.com/u/n.git" {
		t.Errorf("default: got %q", got)
	}
}

func TestCreate_DirAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	repoName := "existing"
	os.MkdirAll(filepath.Join(dir, repoName), 0755)

	runner := &mockRunner{responses: map[string]mockResponse{}}
	ec := &eventCollector{}
	opts := CreateOptions{
		Name:     repoName,
		Location: dir,
	}
	result := Create(runner, runner, opts, ec.emit)

	if len(result.Phases) == 0 {
		t.Fatal("expected at least one phase")
	}
	if !result.Phases[0].HasFailures() {
		t.Error("expected setup phase to have failures when dir already exists")
	}
}

func TestCreate_GitHubPhaseFailure_LocalStillUsable(t *testing.T) {
	dir := t.TempDir()
	repoName := "my-project"
	repoPath := filepath.Join(dir, repoName)

	runner := &mockRunner{responses: map[string]mockResponse{
		// Setup phase succeeds
		fmt.Sprintf("%s/.bare:[init --bare]", repoPath):                                                    {output: ""},
		fmt.Sprintf("%s/.bare:[config remote.origin.fetch +refs/heads/*:refs/remotes/origin/*]", repoPath): {output: ""},
		fmt.Sprintf("%s:[worktree add %s/main -b main]", repoPath, repoPath):                               {output: ""},
		fmt.Sprintf("%s/main:[add -A]", repoPath):                                                          {output: ""},
		fmt.Sprintf("%s/main:[commit -m Initial commit]", repoPath):                                        {output: ""},
	}}
	ghRunner := &mockGhRunner{responses: map[string]mockResponse{
		// GitHub phase fails at user lookup
		fmt.Sprintf("%s:gh[api user --jq .login]", repoPath): {output: "", err: fmt.Errorf("gh: not authenticated")},
	}}

	ec := &eventCollector{}
	opts := CreateOptions{
		Name:          repoName,
		Location:      dir,
		PublishGitHub: true,
		Visibility:    "private",
	}
	result := CreateWithGh(runner, runner, ghRunner, opts, ec.emit)

	// Local repo still usable despite GitHub failure
	if result.RepoPath != repoPath {
		t.Errorf("RepoPath should still be set: got %q", result.RepoPath)
	}
	if len(result.Phases) != 2 {
		t.Fatalf("want 2 phases, got %d", len(result.Phases))
	}
	if result.Phases[0].HasFailures() {
		t.Error("setup phase should not have failures")
	}
	if !result.Phases[1].HasFailures() {
		t.Error("github phase should have failures")
	}
}

func TestCreate_PushFailure_ReportsOrphanedRepo(t *testing.T) {
	dir := t.TempDir()
	repoName := "my-project"
	repoPath := filepath.Join(dir, repoName)
	pushErr := fmt.Errorf("Permission denied (publickey)")

	runner := &mockRunner{responses: map[string]mockResponse{
		fmt.Sprintf("%s/.bare:[init --bare]", repoPath):                                                       {output: ""},
		fmt.Sprintf("%s/.bare:[config remote.origin.fetch +refs/heads/*:refs/remotes/origin/*]", repoPath):    {output: ""},
		fmt.Sprintf("%s:[worktree add %s/main -b main]", repoPath, repoPath):                                  {output: ""},
		fmt.Sprintf("%s/main:[add -A]", repoPath):                                                             {output: ""},
		fmt.Sprintf("%s/main:[commit -m Initial commit]", repoPath):                                           {output: ""},
		fmt.Sprintf("%s/.bare:[remote set-url origin https://github.com/abiswas97/my-project.git]", repoPath): {output: ""},
		fmt.Sprintf("%s/main:[push -u origin main]", repoPath):                                                {output: "", err: pushErr},
	}}
	ghRunner := &mockGhRunner{responses: map[string]mockResponse{
		fmt.Sprintf("%s:gh[api user --jq .login]", repoPath):             {output: "abiswas97"},
		fmt.Sprintf("%s:gh[repo create my-project --private]", repoPath): {output: ""},
		fmt.Sprintf("%s:gh[config get git_protocol]", repoPath):          {output: "https"},
	}}

	ec := &eventCollector{}
	opts := CreateOptions{Name: repoName, Location: dir, PublishGitHub: true, Visibility: "private"}
	result := CreateWithGh(runner, runner, ghRunner, opts, ec.emit)

	var pushStep *StepResult
	for i := range result.Phases {
		if result.Phases[i].Name != "GitHub" {
			continue
		}
		for j := range result.Phases[i].Steps {
			if result.Phases[i].Steps[j].Name == "Push to GitHub" {
				pushStep = &result.Phases[i].Steps[j]
			}
		}
	}
	if pushStep == nil || pushStep.Status != StepFailed {
		t.Fatal("expected a failed 'Push to GitHub' step")
	}
	msg := pushStep.Error.Error()
	if !strings.Contains(msg, opts.Name) {
		t.Errorf("push error should name the orphaned repo %q: %v", opts.Name, msg)
	}
	if !strings.Contains(msg, "delete it or push") {
		t.Errorf("push error should guide the user on the orphaned remote repo: %v", msg)
	}
	if !errors.Is(pushStep.Error, pushErr) {
		t.Errorf("push error must wrap the original error: %v", msg)
	}
}

func TestCreateResult_SetupFailed(t *testing.T) {
	setupBroken := CreateResult{Phases: []Phase{
		{Name: "Setup", Steps: []StepResult{{Name: "Initial commit", Status: StepFailed, Error: fmt.Errorf("exit 128")}}},
	}}
	if failed, err := setupBroken.SetupFailed(); !failed || err == nil {
		t.Errorf("a Setup-phase failure must be hard, got failed=%v err=%v", failed, err)
	}

	githubOnly := CreateResult{Phases: []Phase{
		{Name: "Setup", Steps: []StepResult{{Status: StepDone}}},
		{Name: PhaseGitHub, Steps: []StepResult{{Status: StepFailed, Error: fmt.Errorf("push")}}},
	}}
	if failed, _ := githubOnly.SetupFailed(); failed {
		t.Error("a GitHub-only failure must be soft (local repo is fine)")
	}

	clean := CreateResult{Phases: []Phase{{Name: "Setup", Steps: []StepResult{{Status: StepDone}}}}}
	if failed, _ := clean.SetupFailed(); failed {
		t.Error("a clean result must not report a setup failure")
	}
}
