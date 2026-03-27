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
		SyllabusID:     "india-jee",
		SubjectID:      "india-jee-mathematics",
		SubjectGradeID: "india-jee-mathematics-class-11",
		Country:        "india",
		SourceText:     "Mathematics syllabus covering Algebra, Calculus, and Trigonometry.",
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

	t.Run("subject-grade-yaml-created", func(t *testing.T) {
		found := false
		for path := range result.Files {
			if strings.Contains(path, "subject-grade.yaml") {
				found = true
				break
			}
		}
		if !found {
			t.Error("ScaffoldSubject() should create a subject-grade.yaml file")
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

	t.Run("paths-follow-new-structure", func(t *testing.T) {
		for path := range result.Files {
			if strings.Contains(path, "topics/") {
				// Topic paths should be: .../india-jee-mathematics/india-jee-mathematics-class-11/topics/...
				if !strings.Contains(path, "india-jee-mathematics/india-jee-mathematics-class-11/topics/") {
					t.Errorf("topic path %q should contain subject/subject_grade/topics/ structure", path)
				}
			}
		}
	})
}

func TestScaffoldSubject_MalayNameStripping(t *testing.T) {
	// Mock returns Malay name with grade — subject.yaml should get grade-less name.
	mockResponse := "NAME: Matematik Tingkatan 4\nTOPICS:\n- Fungsi Dan Persamaan Kuadratik\n- Asas Nombor"
	mock := ai.NewMockProvider(mockResponse)
	s := generator.NewScaffolder(mock)

	result, err := s.ScaffoldSubject(context.Background(), generator.ScaffoldRequest{
		SyllabusID:     "malaysia-kssm",
		SubjectID:      "malaysia-kssm-matematik",
		SubjectGradeID: "malaysia-kssm-matematik-tingkatan-4",
		Country:        "malaysia",
		SourceText:     "Matematik Tingkatan 4 DSKP",
	})
	if err != nil {
		t.Fatalf("ScaffoldSubject() error = %v", err)
	}

	t.Run("subject-yaml-has-grade-less-name", func(t *testing.T) {
		for path, content := range result.Files {
			if strings.HasSuffix(path, "subject.yaml") {
				// Check that the name: field (not name_en) has the grade-less name.
				if !strings.Contains(content, "\nname: \"Matematik\"\n") {
					t.Errorf("subject.yaml name should be 'Matematik', got:\n%s", content)
				}
			}
		}
	})

	t.Run("subject-grade-yaml-has-full-name", func(t *testing.T) {
		for path, content := range result.Files {
			if strings.HasSuffix(path, "subject-grade.yaml") {
				if !strings.Contains(content, `name: "Matematik Tingkatan 4"`) {
					t.Errorf("subject-grade.yaml should have full name with grade, got:\n%s", content)
				}
			}
		}
	})
}

func TestStripGradeFromName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Matematik Tingkatan 4", "Matematik"},
		{"Fizik Tingkatan 4", "Fizik"},
		{"Matematika Kelas 10", "Matematika"},
		{"Physics Class 12", "Physics"},
		{"Mathematics Form 3", "Mathematics"},
		{"Mathematics Year 10", "Mathematics"},
		{"Mathematics", "Mathematics"},                         // no grade
		{"Bahasa Melayu Tingkatan 3", "Bahasa Melayu"},        // multi-word subject
		{"Sains Komputer Tingkatan 4", "Sains Komputer"},      // multi-word subject
	}
	for _, c := range tests {
		got := generator.StripGradeFromName(c.input)
		if got != c.want {
			t.Errorf("StripGradeFromName(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestScaffoldSubject_RequiresIDs(t *testing.T) {
	s := generator.NewScaffolder(ai.NewMockProvider(""))

	t.Run("requires-syllabus-id", func(t *testing.T) {
		_, err := s.ScaffoldSubject(context.Background(), generator.ScaffoldRequest{SubjectID: "jee-math"})
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
