package repo

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestCreate_LocalOnly(t *testing.T) {
	dir := t.TempDir()
	repoName := "my-project"
	repoPath := filepath.Join(dir, repoName)

	runner := &mockRunner{responses: map[string]mockResponse{
		fmt.Sprintf("%s/.bare:[init --bare]", repoPath):                                                    {output: ""},
		fmt.Sprintf("%s/.bare:[config remote.origin.fetch +refs/heads/*:refs/remotes/origin/*]", repoPath): {output: ""},
		fmt.Sprintf("%s:[worktree add main -b main]", repoPath):                                            {output: ""},
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
		fmt.Sprintf("%s:[worktree add main -b main]", repoPath):                                            {output: ""},
		fmt.Sprintf("%s/main:[add -A]", repoPath):                                                          {output: ""},
		fmt.Sprintf("%s/main:[commit -m Initial commit]", repoPath):                                        {output: ""},
		// GitHub phase
		fmt.Sprintf("%s:shell[gh api user --jq .login]", repoPath):                                                       {output: "abiswas97"},
		fmt.Sprintf("%s/main:shell[gh repo create my-project --private --description \"\" --source . --push]", repoPath): {output: ""},
		fmt.Sprintf("%s/.bare:[remote set-url origin git@github.com:abiswas97/my-project.git]", repoPath):                {output: ""},
		fmt.Sprintf("%s/main:[push -u origin main]", repoPath):                                                           {output: ""},
		fmt.Sprintf("%s/.bare:[remote set-head origin main]", repoPath):                                                  {output: ""},
	}}

	ec := &eventCollector{}
	opts := CreateOptions{
		Name:          repoName,
		Location:      dir,
		PublishGitHub: true,
		Visibility:    "private",
		Description:   "",
	}
	result := Create(runner, runner, opts, ec.emit)

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
		fmt.Sprintf("%s:[worktree add main -b main]", repoPath):                                            {output: ""},
		fmt.Sprintf("%s/main:[add -A]", repoPath):                                                          {output: ""},
		fmt.Sprintf("%s/main:[commit -m Initial commit]", repoPath):                                        {output: ""},
		// GitHub phase fails at user lookup
		fmt.Sprintf("%s:shell[gh api user --jq .login]", repoPath): {output: "", err: fmt.Errorf("gh: not authenticated")},
	}}

	ec := &eventCollector{}
	opts := CreateOptions{
		Name:          repoName,
		Location:      dir,
		PublishGitHub: true,
		Visibility:    "private",
	}
	result := Create(runner, runner, opts, ec.emit)

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
