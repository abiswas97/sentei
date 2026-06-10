package repo

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/abiswas97/sentei/internal/pipeline"
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
		{url: "https://github.com/user/repo/", want: "repo"},         // trailing slash
		{url: "https://github.com/user/repo.git/", want: "repo"},     // trailing slash + .git
		{url: "https://github.com/user/repo?ref=main", want: "repo"}, // query string
		{url: "git@github.com:user/repo.git#frag", want: "repo"},     // fragment
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
		fmt.Sprintf("%s:[fetch origin]", barePath):                                                   {output: ""},
		// Worktree phase
		fmt.Sprintf("%s:[symbolic-ref --short HEAD]", barePath):                 {output: "main"},
		fmt.Sprintf("%s:[show-ref --verify refs/heads/main]", barePath):         {output: "abc123"},
		fmt.Sprintf("%s:[worktree add %s/main main]", repoPath, repoPath):       {output: ""},
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
			if step.Status == pipeline.StepFailed {
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
		fmt.Sprintf("%s:[fetch origin]", barePath):                                                   {output: ""},
		// symbolic-ref fails — fallback to main
		fmt.Sprintf("%s:[symbolic-ref --short HEAD]", barePath):                 {output: "", err: fmt.Errorf("not found")},
		fmt.Sprintf("%s:[show-ref --verify refs/heads/main]", barePath):         {output: "abc123"},
		fmt.Sprintf("%s:[worktree add %s/main main]", repoPath, repoPath):       {output: ""},
		fmt.Sprintf("%s/main:[branch --set-upstream-to=origin/main]", repoPath): {output: ""},
	}}

	ec := &eventCollector{}
	opts := CloneOptions{URL: "git@github.com:user/repo.git", Location: dir, Name: repoName}
	result := Clone(runner, opts, ec.emit)

	if result.DefaultBranch != "main" {
		t.Errorf("DefaultBranch = %q, want %q (fallback)", result.DefaultBranch, "main")
	}
}

func TestClone_NonStandardDefaultBranch(t *testing.T) {
	dir := t.TempDir()
	repoName := "repo"
	repoPath := filepath.Join(dir, repoName)
	barePath := filepath.Join(repoPath, ".bare")

	runner := &mockRunner{responses: map[string]mockResponse{
		fmt.Sprintf("%s:[clone --bare git@github.com:user/repo.git %s]", dir, barePath):              {output: ""},
		fmt.Sprintf("%s:[config remote.origin.fetch +refs/heads/*:refs/remotes/origin/*]", barePath): {output: ""},
		fmt.Sprintf("%s:[fetch origin]", barePath):                                                   {output: ""},
		// Default branch is neither main nor master; a bare clone records it in HEAD.
		fmt.Sprintf("%s:[symbolic-ref --short HEAD]", barePath):                             {output: "production"},
		fmt.Sprintf("%s:[show-ref --verify refs/heads/production]", barePath):               {output: "abc123"},
		fmt.Sprintf("%s:[worktree add %s/production production]", repoPath, repoPath):       {output: ""},
		fmt.Sprintf("%s/production:[branch --set-upstream-to=origin/production]", repoPath): {output: ""},
	}}

	ec := &eventCollector{}
	opts := CloneOptions{URL: "git@github.com:user/repo.git", Location: dir, Name: repoName}
	result := Clone(runner, opts, ec.emit)

	if result.DefaultBranch != "production" {
		t.Errorf("DefaultBranch = %q, want %q", result.DefaultBranch, "production")
	}
	for _, phase := range result.Phases {
		for _, step := range phase.Steps {
			if step.Status == pipeline.StepFailed {
				t.Errorf("step %q failed: %v", step.Name, step.Error)
			}
		}
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

func TestClone_FetchFailure_StillSucceedsWithoutTracking(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "repo")
	barePath := filepath.Join(repoPath, ".bare")

	// Everything succeeds except the tracking fetch (simulating a network blip
	// after the bare clone already succeeded). The worktree must still be created.
	runner := &mockRunner{responses: map[string]mockResponse{
		fmt.Sprintf("%s:[clone --bare git@h:u/repo.git %s]", dir, barePath):                          {output: ""},
		fmt.Sprintf("%s:[config remote.origin.fetch +refs/heads/*:refs/remotes/origin/*]", barePath): {output: ""},
		fmt.Sprintf("%s:[symbolic-ref --short HEAD]", barePath):                                      {output: "main"},
		fmt.Sprintf("%s:[show-ref --verify refs/heads/main]", barePath):                              {output: "abc"},
		fmt.Sprintf("%s:[worktree add %s/main main]", repoPath, repoPath):                            {output: ""},
		fmt.Sprintf("%s:[fetch origin]", barePath):                                                   {output: "", err: fmt.Errorf("could not read from remote")},
	}}

	ec := &eventCollector{}
	result := Clone(runner, CloneOptions{URL: "git@h:u/repo.git", Location: dir, Name: "repo"}, ec.emit)

	if result.HasFailures() {
		t.Error("a tracking fetch failure must not fail the clone; the worktree is usable")
	}
	if result.WorktreePath != filepath.Join(repoPath, "main") {
		t.Errorf("worktree should still be created, got WorktreePath=%q", result.WorktreePath)
	}
}

func TestClone_EmptyName_RejectedBeforeAnyGitCall(t *testing.T) {
	dir := t.TempDir()
	// Empty responses: any git call would return an error, proving none happens.
	runner := &mockRunner{responses: map[string]mockResponse{}}

	ec := &eventCollector{}
	opts := CloneOptions{URL: "https://host/user/repo/", Location: dir, Name: ""}
	result := Clone(runner, opts, ec.emit)

	if len(result.Phases) != 1 || result.Phases[0].Name != "Validate" {
		t.Fatalf("expected only a Validate phase, got %+v", result.Phases)
	}
	if !result.Phases[0].HasFailures() {
		t.Error("empty name must fail validation")
	}
	if result.DefaultBranch != "" {
		t.Error("no clone should have happened")
	}
}

func TestClone_PathLikeName_Rejected(t *testing.T) {
	dir := t.TempDir()
	runner := &mockRunner{responses: map[string]mockResponse{}}
	ec := &eventCollector{}

	for _, name := range []string{"/abs/target", "../escaped", "nested/name", ".."} {
		result := Clone(runner, CloneOptions{URL: "u", Location: dir, Name: name}, ec.emit)
		if len(result.Phases) != 1 || !result.Phases[0].HasFailures() {
			t.Errorf("name %q should be rejected by validation, got %+v", name, result.Phases)
		}
	}
}

func TestClone_ExistingTarget_RejectedAndPreserved(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "repo")
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(repoPath, "important.txt")
	if err := os.WriteFile(sentinel, []byte("user data"), 0o644); err != nil {
		t.Fatal(err)
	}

	runner := &mockRunner{responses: map[string]mockResponse{}}
	ec := &eventCollector{}
	result := Clone(runner, CloneOptions{URL: "u", Location: dir, Name: "repo"}, ec.emit)

	if len(result.Phases) != 1 || !result.Phases[0].HasFailures() {
		t.Fatalf("existing target must be rejected, got %+v", result.Phases)
	}
	if _, err := os.Stat(sentinel); err != nil {
		t.Error("pre-existing user data must be left untouched")
	}
}

func TestClone_WorktreeFailure_RollsBackPartialDir(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "repo")
	barePath := filepath.Join(repoPath, ".bare")

	// All phases succeed until worktree add fails (empty remote: show-ref fails).
	runner := &mockRunner{
		responses: map[string]mockResponse{
			fmt.Sprintf("%s:[clone --bare u %s]", dir, barePath):                                         {output: ""},
			fmt.Sprintf("%s:[config remote.origin.fetch +refs/heads/*:refs/remotes/origin/*]", barePath): {output: ""},
			fmt.Sprintf("%s:[fetch origin]", barePath):                                                   {output: ""},
			fmt.Sprintf("%s:[symbolic-ref --short HEAD]", barePath):                                      {output: "main"},
			fmt.Sprintf("%s:[show-ref --verify refs/heads/main]", barePath):                              {output: "", err: fmt.Errorf("not found")},
		},
		// Simulate clone --bare creating the bare dir, so rollback has something
		// to remove.
		onRun: func(d string, args []string) {
			if len(args) > 0 && args[0] == "clone" {
				_ = os.MkdirAll(barePath, 0o755)
			}
		},
	}

	ec := &eventCollector{}
	result := Clone(runner, CloneOptions{URL: "u", Location: dir, Name: "repo"}, ec.emit)

	if !result.HasFailures() {
		t.Fatal("expected the clone to fail on an empty remote")
	}
	if result.WorktreePath != "" {
		t.Error("no worktree was created; WorktreePath must be empty")
	}
	if _, err := os.Stat(repoPath); !os.IsNotExist(err) {
		t.Error("the half-built repo directory must be rolled back")
	}
}

func TestClone_TrackingSkip_PreservesRepoDir(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "repo")
	barePath := filepath.Join(repoPath, ".bare")

	// Worktree add succeeds; the best-effort fetch fails. The clone must still
	// succeed (tracking is pipeline.StepSkipped) and must NOT roll back the repo dir.
	runner := &mockRunner{
		responses: map[string]mockResponse{
			fmt.Sprintf("%s:[clone --bare u %s]", dir, barePath):                                         {output: ""},
			fmt.Sprintf("%s:[config remote.origin.fetch +refs/heads/*:refs/remotes/origin/*]", barePath): {output: ""},
			fmt.Sprintf("%s:[symbolic-ref --short HEAD]", barePath):                                      {output: "main"},
			fmt.Sprintf("%s:[show-ref --verify refs/heads/main]", barePath):                              {output: "abc"},
			fmt.Sprintf("%s:[worktree add %s/main main]", repoPath, repoPath):                            {output: ""},
			fmt.Sprintf("%s:[fetch origin]", barePath):                                                   {output: "", err: fmt.Errorf("network down")},
		},
		onRun: func(d string, args []string) {
			if len(args) >= 2 && args[0] == "worktree" && args[1] == "add" {
				_ = os.MkdirAll(filepath.Join(repoPath, "main"), 0o755)
			}
		},
	}

	ec := &eventCollector{}
	result := Clone(runner, CloneOptions{URL: "u", Location: dir, Name: "repo"}, ec.emit)

	if result.HasFailures() {
		t.Error("a tracking-only skip must not fail the clone")
	}
	if result.WorktreePath != filepath.Join(repoPath, "main") {
		t.Errorf("worktree should be created, got WorktreePath=%q", result.WorktreePath)
	}
	if _, err := os.Stat(repoPath); err != nil {
		t.Errorf("repo dir must be preserved on a tracking skip (not rolled back): %v", err)
	}
}
