package cleanup

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

type eventCollector struct {
	events []Event
}

func collectEvents(t *testing.T) *eventCollector {
	t.Helper()
	return &eventCollector{}
}

func (c *eventCollector) emit(e Event) {
	c.events = append(c.events, e)
}

func TestResolveConfigPath(t *testing.T) {
	tests := []struct {
		name       string
		commonDir  string
		wantSuffix string
	}{
		{
			name:       "absolute path (bare repo)",
			commonDir:  "/repo/.bare",
			wantSuffix: "/repo/.bare/config",
		},
		{
			name:       "relative path (normal repo)",
			commonDir:  ".git",
			wantSuffix: ".git/config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockRunner{responses: map[string]mockResponse{
				"/repo:[rev-parse --git-common-dir]": {output: tt.commonDir},
			}}

			path, err := resolveConfigPath(runner, "/repo")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.HasSuffix(path, tt.wantSuffix) {
				t.Errorf("path = %q, want suffix %q", path, tt.wantSuffix)
			}
		})
	}
}
