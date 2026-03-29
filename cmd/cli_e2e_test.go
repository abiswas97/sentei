package cmd_test

import (
	"os/exec"
	"strings"
	"testing"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	tmpBin := t.TempDir() + "/sentei"
	build := exec.Command("go", "build", "-o", tmpBin, ".")
	build.Dir = ".."
	out, err := build.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}
	return tmpBin
}

func TestEcosystemsCLI(t *testing.T) {
	bin := buildBinary(t)

	cmd := exec.Command(bin, "ecosystems")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sentei ecosystems failed: %v\n%s", err, out)
	}

	output := string(out)
	for _, want := range []string{"Ecosystems (", "pnpm", "go", "SOURCE"} {
		if !strings.Contains(output, want) {
			t.Errorf("output missing %q\nfull output:\n%s", want, output)
		}
	}
}

func TestIntegrationsCLI(t *testing.T) {
	bin := buildBinary(t)

	cmd := exec.Command(bin, "integrations")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sentei integrations failed: %v\n%s", err, out)
	}

	output := string(out)
	for _, want := range []string{"Integrations (2", "code-review-graph", "cocoindex-code", "https://github.com/"} {
		if !strings.Contains(output, want) {
			t.Errorf("output missing %q\nfull output:\n%s", want, output)
		}
	}
}
