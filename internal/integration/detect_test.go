package integration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectPresent_DirExists(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".code-review-graph"), 0o755); err != nil {
		t.Fatal(err)
	}
	if !DetectPresent(dir, codeReviewGraph()) {
		t.Error("expected DetectPresent to return true when .code-review-graph exists")
	}
}

func TestDetectPresent_DirMissing(t *testing.T) {
	dir := t.TempDir()
	if DetectPresent(dir, codeReviewGraph()) {
		t.Error("expected DetectPresent to return false in empty directory")
	}
}

func TestDetectPresent_MultipleEntries(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".cocoindex_code"), 0o755); err != nil {
		t.Fatal(err)
	}
	if !DetectPresent(dir, cocoindexCode()) {
		t.Error("expected DetectPresent to return true when .cocoindex_code exists")
	}
}

func TestDetectAllPresent(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".code-review-graph"), 0o755); err != nil {
		t.Fatal(err)
	}

	result := DetectAllPresent(dir, All())

	if !result["code-review-graph"] {
		t.Error("expected code-review-graph to be detected as present")
	}
	if result["cocoindex-code"] {
		t.Error("expected cocoindex-code to be detected as absent")
	}
}
