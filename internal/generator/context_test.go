package generator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/generator"
)

func TestBuildContext(t *testing.T) {
	repoDir := setupTestRepo(t)

	ctx, err := generator.BuildContext(repoDir, "F1-01")
	if err != nil {
		t.Fatalf("BuildContext() error = %v", err)
	}

	if ctx.Topic.ID != "F1-01" {
		t.Errorf("Topic.ID = %q, want %q", ctx.Topic.ID, "F1-01")
	}
	if ctx.Topic.Name == "" {
		t.Error("Topic.Name is empty")
	}
}

func TestBuildContext_WithPrerequisites(t *testing.T) {
	repoDir := setupTestRepo(t)

	ctx, err := generator.BuildContext(repoDir, "F1-02")
	if err != nil {
		t.Fatalf("BuildContext() error = %v", err)
	}

	if len(ctx.Prerequisites) == 0 {
		t.Error("Prerequisites should not be empty for F1-02")
	}
}

func TestBuildContext_NotFound(t *testing.T) {
	repoDir := setupTestRepo(t)

	_, err := generator.BuildContext(repoDir, "NONEXISTENT")
	if err == nil {
		t.Error("BuildContext() should error for non-existent topic")
	}
}

// setupTestRepo creates a minimal OSS repo structure for testing.
func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	topicsDir := filepath.Join(dir, "curricula", "test", "test-algebra", "test-algebra-1", "topics")
	os.MkdirAll(topicsDir, 0o755)

	// Topic F1-01 (no prerequisites)
	os.WriteFile(filepath.Join(topicsDir, "01-test.yaml"), []byte(`
id: F1-01
name: "Test Topic One"
subject_id: algebra
syllabus_id: test-syllabus
difficulty: beginner
learning_objectives:
  - id: 1.0.1
    text: "Test objective"
    bloom: understand
prerequisites:
  required: []
quality_level: 1
provenance: human
`), 0o644)

	// Topic F1-02 (requires F1-01)
	os.WriteFile(filepath.Join(topicsDir, "02-test.yaml"), []byte(`
id: F1-02
name: "Test Topic Two"
subject_id: algebra
syllabus_id: test-syllabus
difficulty: beginner
learning_objectives:
  - id: 1.0.1
    text: "Test objective two"
    bloom: apply
prerequisites:
  required:
    - F1-01
quality_level: 1
provenance: human
`), 0o644)

	return dir
}
