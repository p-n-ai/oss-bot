package generator_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"github.com/p-n-ai/oss-bot/internal/generator"
)

func TestScaffoldSyllabus(t *testing.T) {
	mockResponse := "NAME: Joint Entrance Examination\nDESCRIPTION: India's engineering entrance exam\nSUBJECTS:\n- Physics\n- Chemistry\n- Mathematics"
	mock := ai.NewMockProvider(mockResponse)
	s := generator.NewScaffolder(mock)

	result, err := s.ScaffoldSyllabus(context.Background(), generator.ScaffoldRequest{
		SyllabusID: "india-jee",
		Country:    "india",
		SourceText: "The JEE curriculum covers Physics, Chemistry, and Mathematics.",
	})
	if err != nil {
		t.Fatalf("ScaffoldSyllabus() error = %v", err)
	}

	t.Run("files-generated", func(t *testing.T) {
		if len(result.Files) == 0 {
			t.Error("ScaffoldSyllabus() should generate at least one file")
		}
	})

	t.Run("syllabus-yaml-created", func(t *testing.T) {
		found := false
		for path := range result.Files {
			if strings.Contains(path, "syllabus.yaml") {
				found = true
				break
			}
		}
		if !found {
			t.Error("ScaffoldSyllabus() should create a syllabus.yaml file")
		}
	})

	t.Run("summary-non-empty", func(t *testing.T) {
		if result.Summary == "" {
			t.Error("ScaffoldSyllabus() should return a non-empty summary")
		}
	})
}

func TestScaffoldSyllabus_RequiresSyllabusID(t *testing.T) {
	s := generator.NewScaffolder(ai.NewMockProvider(""))
	_, err := s.ScaffoldSyllabus(context.Background(), generator.ScaffoldRequest{})
	if err == nil {
		t.Error("ScaffoldSyllabus() should error when SyllabusID is empty")
	}
}

func TestScaffoldSyllabus_NoAIProvider(t *testing.T) {
	// Without AI provider, should still produce a minimal stub
	s := generator.NewScaffolder(nil)
	result, err := s.ScaffoldSyllabus(context.Background(), generator.ScaffoldRequest{
		SyllabusID: "test-syllabus",
		Country:    "test",
	})
	if err != nil {
		t.Fatalf("ScaffoldSyllabus() error = %v (should use fallback)", err)
	}
	if len(result.Files) == 0 {
		t.Error("ScaffoldSyllabus() should produce a stub even without AI provider")
	}
}

func TestScaffoldSubject(t *testing.T) {
	mockResponse := "NAME: Mathematics\nTOPICS:\n- Algebra\n- Calculus\n- Trigonometry"
	mock := ai.NewMockProvider(mockResponse)
	s := generator.NewScaffolder(mock)

	result, err := s.ScaffoldSubject(context.Background(), generator.ScaffoldRequest{
		SyllabusID: "india-jee",
		SubjectID:  "mathematics",
		Country:    "india",
		SourceText: "Mathematics syllabus covering Algebra, Calculus, and Trigonometry.",
	})
	if err != nil {
		t.Fatalf("ScaffoldSubject() error = %v", err)
	}

	t.Run("subject-yaml-created", func(t *testing.T) {
		found := false
		for path := range result.Files {
			if strings.Contains(path, "subject.yaml") {
				found = true
				break
			}
		}
		if !found {
			t.Error("ScaffoldSubject() should create a subject.yaml file")
		}
	})

	t.Run("topic-stubs-created", func(t *testing.T) {
		topicCount := 0
		for path := range result.Files {
			if strings.Contains(path, "topics/") && strings.HasSuffix(path, ".yaml") {
				topicCount++
			}
		}
		if topicCount == 0 {
			t.Error("ScaffoldSubject() should create topic stub files")
		}
	})
}

func TestScaffoldSubject_RequiresIDs(t *testing.T) {
	s := generator.NewScaffolder(ai.NewMockProvider(""))

	t.Run("requires-syllabus-id", func(t *testing.T) {
		_, err := s.ScaffoldSubject(context.Background(), generator.ScaffoldRequest{SubjectID: "math"})
		if err == nil {
			t.Error("ScaffoldSubject() should error when SyllabusID is empty")
		}
	})

	t.Run("requires-subject-id", func(t *testing.T) {
		_, err := s.ScaffoldSubject(context.Background(), generator.ScaffoldRequest{SyllabusID: "jee"})
		if err == nil {
			t.Error("ScaffoldSubject() should error when SubjectID is empty")
		}
	})
}

func TestScaffoldWriteFiles(t *testing.T) {
	s := generator.NewScaffolder(nil)
	result, err := s.ScaffoldSyllabus(context.Background(), generator.ScaffoldRequest{
		SyllabusID: "write-test",
		Country:    "test",
	})
	if err != nil {
		t.Fatalf("ScaffoldSyllabus() error = %v", err)
	}

	outputDir := t.TempDir()
	if err := s.WriteFiles(result, outputDir); err != nil {
		t.Fatalf("WriteFiles() error = %v", err)
	}

	for relPath := range result.Files {
		full := filepath.Join(outputDir, relPath)
		if _, err := os.Stat(full); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist after WriteFiles()", full)
		}
	}
}
