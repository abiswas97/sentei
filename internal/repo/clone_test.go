package repo

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestDeriveRepoName(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{url: "git@github.com:user/repo.git", want: "repo"},
		{url: "https://github.com/user/repo.git", want: "repo"},
		{url: "https://github.com/user/repo", want: "repo"},
		{url: "git@gitlab.com:org/sub/project.git", want: "project"},
		{url: "https://github.com/user/my-repo.git", want: "my-repo"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got := DeriveRepoName(tt.url)
			if got != tt.want {
				t.Errorf("DeriveRepoName(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestClone_Successful(t *testing.T) {
	dir := t.TempDir()
	repoName := "repo"
	repoPath := filepath.Join(dir, repoName)
	barePath := filepath.Join(repoPath, ".bare")

	runner := &mockRunner{responses: map[string]mockResponse{
		// Clone phase
		fmt.Sprintf("%s:[clone --bare git@github.com:user/repo.git %s]", dir, barePath): {output: ""},
		// Structure phase
		fmt.Sprintf("%s:[config remote.origin.fetch +refs/heads/*:refs/remotes/origin/*]", barePath): {output: ""},
		// Worktree phase
		fmt.Sprintf("%s:[symbolic-ref refs/remotes/origin/HEAD]", barePath):     {output: "refs/remotes/origin/main"},
		fmt.Sprintf("%s:[worktree add main main]", repoPath):                    {output: ""},
		fmt.Sprintf("%s/main:[branch --set-upstream-to=origin/main]", repoPath): {output: ""},
	}}

	ec := &eventCollector{}
	opts := CloneOptions{
		URL:      "git@github.com:user/repo.git",
		Location: dir,
		Name:     repoName,
	}
	result := Clone(runner, opts, ec.emit)

	if result.RepoPath != repoPath {
		t.Errorf("RepoPath = %q, want %q", result.RepoPath, repoPath)
	}
	if result.DefaultBranch != "main" {
		t.Errorf("DefaultBranch = %q, want %q", result.DefaultBranch, "main")
	}
	for _, phase := range result.Phases {
		for _, step := range phase.Steps {
			if step.Status == StepFailed {
				t.Errorf("step %q failed: %v", step.Name, step.Error)
			}
		}
	}
}

func TestClone_DefaultBranchFallback(t *testing.T) {
	dir := t.TempDir()
	repoName := "repo"
	repoPath := filepath.Join(dir, repoName)
	barePath := filepath.Join(repoPath, ".bare")

	runner := &mockRunner{responses: map[string]mockResponse{
		fmt.Sprintf("%s:[clone --bare git@github.com:user/repo.git %s]", dir, barePath):              {output: ""},
		fmt.Sprintf("%s:[config remote.origin.fetch +refs/heads/*:refs/remotes/origin/*]", barePath): {output: ""},
		// symbolic-ref fails — fallback to main
		fmt.Sprintf("%s:[symbolic-ref refs/remotes/origin/HEAD]", barePath):     {output: "", err: fmt.Errorf("not found")},
		fmt.Sprintf("%s:[show-ref --verify refs/heads/main]", barePath):         {output: "abc123"},
		fmt.Sprintf("%s:[worktree add main main]", repoPath):                    {output: ""},
		fmt.Sprintf("%s/main:[branch --set-upstream-to=origin/main]", repoPath): {output: ""},
	}}

	ec := &eventCollector{}
	opts := CloneOptions{URL: "git@github.com:user/repo.git", Location: dir, Name: repoName}
	result := Clone(runner, opts, ec.emit)

	if result.DefaultBranch != "main" {
		t.Errorf("DefaultBranch = %q, want %q (fallback)", result.DefaultBranch, "main")
	}
}

func TestClone_NetworkError(t *testing.T) {
	dir := t.TempDir()
	barePath := filepath.Join(dir, "repo", ".bare")

	runner := &mockRunner{responses: map[string]mockResponse{
		fmt.Sprintf("%s:[clone --bare git@github.com:user/repo.git %s]", dir, barePath): {
			output: "", err: fmt.Errorf("fatal: Could not read from remote repository"),
		},
	}}

	ec := &eventCollector{}
	opts := CloneOptions{URL: "git@github.com:user/repo.git", Location: dir, Name: "repo"}
	result := Clone(runner, opts, ec.emit)

	if len(result.Phases) == 0 || !result.Phases[0].HasFailures() {
		t.Error("expected clone phase to fail on network error")
	}
}
