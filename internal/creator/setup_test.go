package creator

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/progress"
	"github.com/abiswas97/sentei/internal/testutil/mock"
)

func TestCreateWorktreeStep(t *testing.T) {
	tests := []struct {
		name         string
		branch       string
		baseBranch   string
		repoPath     string
		branchExists bool
		runnerErr    error
		wantStatus   progress.StepStatus
		wantPath     string
	}{
		{
			name:       "successful creation with new branch",
			branch:     "feature/auth",
			baseBranch: "main",
			repoPath:   "/repo",
			wantStatus: progress.StepDone,
			wantPath:   "/repo/feature-auth",
		},
		{
			name:         "checks out existing branch into new worktree",
			branch:       "feature/dup",
			baseBranch:   "main",
			repoPath:     "/repo",
			branchExists: true,
			wantStatus:   progress.StepDone,
			wantPath:     "/repo/feature-dup",
		},
		{
			name:       "worktree add failure is surfaced",
			branch:     "feature/broken",
			baseBranch: "main",
			repoPath:   "/repo",
			runnerErr:  fmt.Errorf("fatal: something went wrong"),
			wantStatus: progress.StepFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wtPath := git.WorktreePath(tt.repoPath, tt.branch)

			responses := map[string]mock.Response{
				fmt.Sprintf("%s:[show-ref --verify refs/heads/%s]", tt.repoPath, tt.branch): {
					Err: func() error {
						if tt.branchExists {
							return nil
						}
						return fmt.Errorf("not found")
					}(),
				},
			}
			if tt.branchExists {
				responses[fmt.Sprintf("%s:[worktree add %s %s]", tt.repoPath, wtPath, tt.branch)] = mock.Response{Err: tt.runnerErr}
			} else {
				responses[fmt.Sprintf("%s:[worktree add %s -b %s %s]", tt.repoPath, wtPath, tt.branch, tt.baseBranch)] = mock.Response{Err: tt.runnerErr}
			}
			runner := &mock.Runner{Responses: responses}

			ec := &mock.EventCollector[progress.Event]{}
			result, path := createWorktreeStep(runner, tt.repoPath, tt.branch, tt.baseBranch, ec.Emit)

			if result.Status != tt.wantStatus {
				t.Errorf("status = %v, want %v", result.Status, tt.wantStatus)
			}
			if tt.wantStatus == progress.StepDone && path != tt.wantPath {
				t.Errorf("path = %q, want %q", path, tt.wantPath)
			}
			if len(ec.Events) == 0 {
				t.Error("expected at least one event emitted")
			}
		})
	}
}

func TestMergeBaseStep(t *testing.T) {
	tests := []struct {
		name       string
		mergeBase  bool
		baseBranch string
		runnerErr  error
		wantStatus progress.StepStatus
	}{
		{
			name:       "successful merge",
			mergeBase:  true,
			baseBranch: "main",
			wantStatus: progress.StepDone,
		},
		{
			name:       "merge conflict continues",
			mergeBase:  true,
			baseBranch: "main",
			runnerErr:  fmt.Errorf("merge conflict"),
			wantStatus: progress.StepFailed,
		},
		{
			name:       "merge disabled",
			mergeBase:  false,
			baseBranch: "main",
			wantStatus: progress.StepSkipped,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mock.Runner{Responses: map[string]mock.Response{
				"/repo/feature-auth:[merge main --no-edit]": {
					Output: "",
					Err:    tt.runnerErr,
				},
			}}

			ec := &mock.EventCollector[progress.Event]{}
			result := mergeBaseStep(runner, "/repo/feature-auth", tt.baseBranch, tt.mergeBase, ec.Emit)

			if result.Status != tt.wantStatus {
				t.Errorf("status = %v, want %v", result.Status, tt.wantStatus)
			}
		})
	}
}

func TestCopyEnvFilesStep(t *testing.T) {
	tests := []struct {
		name       string
		envFiles   []string
		srcFiles   []string
		wantStatus progress.StepStatus
	}{
		{
			name:       "copies existing files",
			envFiles:   []string{".env", ".env.local"},
			srcFiles:   []string{".env"},
			wantStatus: progress.StepDone,
		},
		{
			name:       "no env files configured",
			envFiles:   nil,
			wantStatus: progress.StepSkipped,
		},
		{
			name:       "no source files exist",
			envFiles:   []string{".env"},
			srcFiles:   nil,
			wantStatus: progress.StepDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srcDir := t.TempDir()
			dstDir := t.TempDir()

			for _, f := range tt.srcFiles {
				os.WriteFile(filepath.Join(srcDir, f), []byte("SECRET=val"), 0644)
			}

			ec := &mock.EventCollector[progress.Event]{}
			result := copyEnvFilesStep(srcDir, dstDir, tt.envFiles, ec.Emit)

			if result.Status != tt.wantStatus {
				t.Errorf("status = %v, want %v", result.Status, tt.wantStatus)
			}

			for _, f := range tt.srcFiles {
				dstPath := filepath.Join(dstDir, f)
				if _, err := os.Stat(dstPath); os.IsNotExist(err) {
					t.Errorf("expected %s to be copied to dest", f)
				}
			}
		})
	}
}
