// Package mock provides shared test doubles for sentei's runner interfaces.
// It lives below internal/testutil (which imports production packages) so
// that any package's tests — including internal/git's — can use it without
// an import cycle.
package mock

import (
	"fmt"
	"sync"
)

// Response is the canned result for one expected call.
type Response struct {
	Output string
	Err    error
}

// Runner fakes git.CommandRunner and git.ShellRunner. Expected calls are
// keyed "dir:[arg1 arg2 ...]" for Run and "dir:shell[command]" for RunShell;
// an unmatched key returns an error naming the call so the failing fixture is
// obvious from the test output.
type Runner struct {
	mu        sync.Mutex
	Responses map[string]Response
	Calls     []string
	// OnRun, if set, runs before each Run call — lets a test simulate the
	// filesystem side effects of a git command (e.g. clone --bare creating
	// the bare dir).
	OnRun func(dir string, args []string)
}

func (m *Runner) Run(dir string, args ...string) (string, error) {
	if m.OnRun != nil {
		m.OnRun(dir, args)
	}
	key := fmt.Sprintf("%s:%v", dir, args)
	m.record(key)
	if resp, ok := m.Responses[key]; ok {
		return resp.Output, resp.Err
	}
	return "", fmt.Errorf("unexpected call: %s", key)
}

func (m *Runner) RunShell(dir string, command string) (string, error) {
	key := fmt.Sprintf("%s:shell[%s]", dir, command)
	m.record(key)
	if resp, ok := m.Responses[key]; ok {
		return resp.Output, resp.Err
	}
	return "", fmt.Errorf("unexpected shell call: %s", key)
}

func (m *Runner) record(key string) {
	m.mu.Lock()
	m.Calls = append(m.Calls, key)
	m.mu.Unlock()
}

// EventCollector accumulates emitted events for assertions.
type EventCollector[E any] struct {
	Events []E
}

func (c *EventCollector[E]) Emit(e E) {
	c.Events = append(c.Events, e)
}
