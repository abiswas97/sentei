package creator

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type mockRunner struct {
	responses map[string]mockResponse
	calls     []string
}

type mockResponse struct {
	output string
	err    error
}

func (m *mockRunner) Run(dir string, args ...string) (string, error) {
	key := fmt.Sprintf("%s:%v", dir, args)
	m.calls = append(m.calls, key)
	if resp, ok := m.responses[key]; ok {
		return resp.output, resp.err
	}
	return "", fmt.Errorf("unexpected call: %s", key)
}

func (m *mockRunner) RunShell(dir string, command string) (string, error) {
	key := fmt.Sprintf("%s:shell[%s]", dir, command)
	m.calls = append(m.calls, key)
	if resp, ok := m.responses[key]; ok {
		return resp.output, resp.err
	}
	return "", fmt.Errorf("unexpected shell call: %s", key)
}

type eventCollector struct {
	events []Event
}

func (c *eventCollector) emit(e Event) {
	c.events = append(c.events, e)
}

func TestSanitizeBranchPath(t *testing.T) {
	tests := []struct {
		name   string
		branch string
		want   string
	}{
		{
			name:   "slash replaced with dash",
			branch: "feature/auth",
			want:   "feature-auth",
		},
		{
			name:   "multiple slashes",
			branch: "bugfix/login/redirect",
			want:   "bugfix-login-redirect",
		},
		{
			name:   "no slash unchanged",
			branch: "hotfix",
			want:   "hotfix",
		},
		{
			name:   "trailing slash stripped",
			branch: "feature/",
			want:   "feature-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeBranchPath(tt.branch)
			if got != tt.want {
				t.Errorf("SanitizeBranchPath(%q) = %q, want %q", tt.branch, got, tt.want)
			}
		})
	}
}

func TestCreateWorktreeStep(t *testing.T) {
	tests := []struct {
		name       string
		branch     string
		baseBranch string
		repoPath   string
		runnerErr  error
		wantStatus StepStatus
		wantPath   string
	}{
		{
			name:       "successful creation",
			branch:     "feature/auth",
			baseBranch: "main",
			repoPath:   "/repo",
			wantStatus: StepDone,
			wantPath:   "/repo/feature-auth",
		},
		{
			name:       "branch already exists",
			branch:     "feature/dup",
			baseBranch: "main",
			repoPath:   "/repo",
			runnerErr:  fmt.Errorf("fatal: branch already exists"),
			wantStatus: StepFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitized := SanitizeBranchPath(tt.branch)
			wtPath := filepath.Join(tt.repoPath, sanitized)

			runner := &mockRunner{responses: map[string]mockResponse{
				fmt.Sprintf("%s:[worktree add %s -b %s %s]", tt.repoPath, wtPath, tt.branch, tt.baseBranch): {
					output: "",
					err:    tt.runnerErr,
				},
			}}

			ec := &eventCollector{}
			result, path := createWorktreeStep(runner, tt.repoPath, tt.branch, tt.baseBranch, ec.emit)

			if result.Status != tt.wantStatus {
				t.Errorf("status = %v, want %v", result.Status, tt.wantStatus)
			}
			if tt.wantStatus == StepDone && path != tt.wantPath {
				t.Errorf("path = %q, want %q", path, tt.wantPath)
			}
			if len(ec.events) == 0 {
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
		wantStatus StepStatus
	}{
		{
			name:       "successful merge",
			mergeBase:  true,
			baseBranch: "main",
			wantStatus: StepDone,
		},
		{
			name:       "merge conflict continues",
			mergeBase:  true,
			baseBranch: "main",
			runnerErr:  fmt.Errorf("merge conflict"),
			wantStatus: StepFailed,
		},
		{
			name:       "merge disabled",
			mergeBase:  false,
			baseBranch: "main",
			wantStatus: StepSkipped,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockRunner{responses: map[string]mockResponse{
				"/repo/feature-auth:[merge main --no-edit]": {
					output: "",
					err:    tt.runnerErr,
				},
			}}

			ec := &eventCollector{}
			result := mergeBaseStep(runner, "/repo/feature-auth", tt.baseBranch, tt.mergeBase, ec.emit)

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
		wantStatus StepStatus
	}{
		{
			name:       "copies existing files",
			envFiles:   []string{".env", ".env.local"},
			srcFiles:   []string{".env"},
			wantStatus: StepDone,
		},
		{
			name:       "no env files configured",
			envFiles:   nil,
			wantStatus: StepSkipped,
		},
		{
			name:       "no source files exist",
			envFiles:   []string{".env"},
			srcFiles:   nil,
			wantStatus: StepDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srcDir := t.TempDir()
			dstDir := t.TempDir()

			for _, f := range tt.srcFiles {
				os.WriteFile(filepath.Join(srcDir, f), []byte("SECRET=val"), 0644)
			}

			ec := &eventCollector{}
			result := copyEnvFilesStep(srcDir, dstDir, tt.envFiles, ec.emit)

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
