package repo

import (
	"fmt"
	"strings"
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

func TestDetectContext(t *testing.T) {
	tests := []struct {
		name      string
		responses map[string]mockResponse
		setupDir  func(t *testing.T, dir string)
		want      RepoContext
	}{
		{
			name: "bare repo detected via git",
			responses: map[string]mockResponse{
				"{dir}:[rev-parse --is-bare-repository]": {output: "true"},
			},
			want: ContextBareRepo,
		},
		{
			name: "worktree inside bare repo with .bare directory",
			responses: map[string]mockResponse{
				"{dir}:[rev-parse --is-bare-repository]": {output: "false"},
				"{dir}:[rev-parse --git-dir]":            {output: "/repo/.bare"},
				"{dir}:[rev-parse --git-common-dir]":     {output: "/repo/.bare"},
			},
			want: ContextBareRepo,
		},
		{
			name: "non-bare regular repo",
			responses: map[string]mockResponse{
				"{dir}:[rev-parse --is-bare-repository]": {output: "false"},
				"{dir}:[rev-parse --git-dir]":            {output: ".git"},
				"{dir}:[rev-parse --git-common-dir]":     {output: ".git"},
			},
			want: ContextNonBareRepo,
		},
		{
			name: "no repo at all",
			responses: map[string]mockResponse{
				"{dir}:[rev-parse --is-bare-repository]": {output: "", err: fmt.Errorf("not a git repository")},
				"{dir}:[rev-parse --git-dir]":            {output: "", err: fmt.Errorf("not a git repository")},
			},
			want: ContextNoRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			// Replace {dir} placeholders in response keys
			resolved := make(map[string]mockResponse)
			for k, v := range tt.responses {
				resolvedKey := strings.ReplaceAll(k, "{dir}", dir)
				resolvedVal := mockResponse{
					output: strings.ReplaceAll(v.output, "{dir}", dir),
					err:    v.err,
				}
				resolved[resolvedKey] = resolvedVal
			}

			if tt.setupDir != nil {
				tt.setupDir(t, dir)
			}

			runner := &mockRunner{responses: resolved}
			got := DetectContext(runner, dir)
			if got != tt.want {
				t.Errorf("DetectContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveBareRoot(t *testing.T) {
	tests := []struct {
		name      string
		responses map[string]mockResponse
		wantExact string // if set, expect this exact path
		wantSelf  bool   // if true, expect the input dir back
	}{
		{
			name: "from worktree resolves to bare root",
			responses: map[string]mockResponse{
				"{dir}:[rev-parse --git-common-dir]": {output: "/code/myproject/.bare"},
			},
			wantExact: "/code/myproject",
		},
		{
			name: "git fails returns original path",
			responses: map[string]mockResponse{
				"{dir}:[rev-parse --git-common-dir]": {output: "", err: fmt.Errorf("not a git repo")},
			},
			wantSelf: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			resolved := make(map[string]mockResponse)
			for k, v := range tt.responses {
				resolved[strings.ReplaceAll(k, "{dir}", dir)] = mockResponse{
					output: strings.ReplaceAll(v.output, "{dir}", dir),
					err:    v.err,
				}
			}

			runner := &mockRunner{responses: resolved}
			got := ResolveBareRoot(runner, dir)

			if tt.wantExact != "" && got != tt.wantExact {
				t.Errorf("ResolveBareRoot() = %q, want %q", got, tt.wantExact)
			}
			if tt.wantSelf && got != dir {
				t.Errorf("ResolveBareRoot() = %q, want %q", got, dir)
			}
		})
	}
}
