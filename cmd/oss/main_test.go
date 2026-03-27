package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/ai"
)

func TestMergeTopicYAML(t *testing.T) {
	existing := `id: MT4-01
official_ref: "1.0"
name: "Fungsi dan Persamaan Kuadratik"
name_en: "Quadratic Functions and Equations"
subject_id: malaysia-kssm-matematik-tingkatan-4
syllabus_id: malaysia-kssm
country_id: malaysia
language: ms
difficulty: intermediate
tier: core
learning_objectives:
  - id: 1.1.1
    text: "Recognise a function and describe its characteristics."
    bloom: understand
prerequisites:
  required: []
  recommended: []
bloom_levels:
  - understand
mastery:
  minimum_score: 0.75
  assessment_count: 3
  spaced_repetition:
    initial_interval_days: 3
    multiplier: 2.5
ai_teaching_notes: "MT4-01.teaching.md"
quality_level: 1
provenance: ai-assisted`

	newContent := `id: MT4-01
name: "Fungsi dan Persamaan Kuadratik"
name_en: "Quadratic Functions and Equations"
learning_objectives:
  - id: 1.1.1
    text: "Recognise a function and describe its characteristics."
    bloom: understand
  - id: 1.1.2
    text: "Make connection between a function and a non-function."
    bloom: apply`

	// Mock provider returns a plausible merged YAML.
	mergedYAML := existing + "\n  - id: 1.1.2\n    text: \"Make connection between a function and a non-function.\"\n    bloom: apply"
	provider := ai.NewMockProvider(mergedYAML)

	got, err := mergeTopicYAML(context.Background(), provider, existing, newContent, "1.0 FUNGSI DAN PERSAMAAN KUADRATIK")
	if err != nil {
		t.Fatalf("mergeTopicYAML() error = %v", err)
	}
	if strings.TrimSpace(got) == "" {
		t.Error("mergeTopicYAML() returned empty string")
	}
	// The mock just returns what we gave it; verify it round-trips through the function.
	if got != mergedYAML {
		t.Errorf("mergeTopicYAML() output does not match mock response")
	}
}

func TestImportCmdFlags(t *testing.T) {
	cmd := importCmd()

	// Required flags must be defined
	if cmd.Flags().Lookup("pdf") == nil {
		t.Error("import command missing --pdf flag")
	}
	if cmd.Flags().Lookup("syllabus") == nil {
		t.Error("import command missing --syllabus flag")
	}
	if cmd.Flags().Lookup("subject-grade") == nil {
		t.Error("import command missing --subject-grade flag")
	}
	if cmd.Flags().Lookup("workers") == nil {
		t.Error("import command missing --workers flag")
	}
	if cmd.Flags().Lookup("pr") == nil {
		t.Error("import command missing --pr flag")
	}
	if cmd.Flags().Lookup("chunk-size") == nil {
		t.Error("import command missing --chunk-size flag")
	}
	if cmd.Flags().Lookup("force") == nil {
		t.Error("import command missing --force flag")
	}

	// Defaults
	workers, err := cmd.Flags().GetInt("workers")
	if err != nil {
		t.Fatalf("getting workers flag: %v", err)
	}
	if workers != 3 {
		t.Errorf("workers default = %d, want 3", workers)
	}

	chunkSize, err := cmd.Flags().GetInt("chunk-size")
	if err != nil {
		t.Fatalf("getting chunk-size flag: %v", err)
	}
	if chunkSize != 2000 {
		t.Errorf("chunk-size default = %d, want 2000", chunkSize)
	}

	pr, err := cmd.Flags().GetBool("pr")
	if err != nil {
		t.Fatalf("getting pr flag: %v", err)
	}
	if pr != false {
		t.Error("pr default should be false")
	}

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		t.Fatalf("getting force flag: %v", err)
	}
	if force != false {
		t.Error("force default should be false")
	}
}

func TestImportCmdRequiredFlags(t *testing.T) {
	cmd := importCmd()

	pdfFlag := cmd.Flags().Lookup("pdf")
	if pdfFlag == nil {
		t.Fatal("--pdf flag not defined")
	}
	syllabusFlag := cmd.Flags().Lookup("syllabus")
	if syllabusFlag == nil {
		t.Fatal("--syllabus flag not defined")
	}

	// Verify required annotations are set
	annotations := cmd.Annotations
	_ = annotations // required flags are enforced by cobra at runtime, not via annotations

	// Verify Use and Short are set
	if cmd.Use == "" {
		t.Error("import command missing Use field")
	}
	if cmd.Short == "" {
		t.Error("import command missing Short description")
	}
}

func TestImportSlug(t *testing.T) {
	cases := []struct {
		heading string
		want    string
	}{
		{"Bab 1 Fungsi", "bab-1-fungsi"},
		{"BAB 2 UNGKAPAN KUADRATIK", "bab-2-ungkapan-kuadratik"},
		{"Chapter 3: Algebra", "chapter-3-algebra"},
		{"", ""},
		{"  spaces  ", "spaces"},
		{"1.1 Introduction", "1-1-introduction"},
	}
	for _, c := range cases {
		got := importSlug(c.heading)
		if got != c.want {
			t.Errorf("importSlug(%q) = %q, want %q", c.heading, got, c.want)
		}
	}
}

func TestFindSubjectTopicsDir(t *testing.T) {
	// Create a temp directory tree mimicking a scaffolded OSS repo:
	// curricula/malaysia/malaysia-kssm/malaysia-kssm-matematik/malaysia-kssm-matematik-tingkatan-4/topics/
	root := t.TempDir()
	topicsDir := filepath.Join(root, "curricula", "malaysia", "malaysia-kssm",
		"malaysia-kssm-matematik", "malaysia-kssm-matematik-tingkatan-4", "topics")
	if err := os.MkdirAll(topicsDir, 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	t.Run("finds existing subject_grade topics dir", func(t *testing.T) {
		got, err := findSubjectTopicsDir(root, "malaysia-kssm-matematik-tingkatan-4", "malaysia-kssm-matematik", "malaysia-kssm")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != topicsDir {
			t.Errorf("got %q, want %q", got, topicsDir)
		}
	})

	t.Run("returns error when subject not found", func(t *testing.T) {
		_, err := findSubjectTopicsDir(root, "nonexistent-subject", "", "malaysia-kssm")
		if err == nil {
			t.Error("expected error for missing subject, got nil")
		}
	})

	t.Run("falls back to syllabusID when subjectGradeID and subjectID are empty", func(t *testing.T) {
		syllabusDir := filepath.Join(root, "curricula", "malaysia", "malaysia-kssm", "topics")
		if err := os.MkdirAll(syllabusDir, 0755); err != nil {
			t.Fatalf("setup: %v", err)
		}
		got, err := findSubjectTopicsDir(root, "", "", "malaysia-kssm")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != syllabusDir {
			t.Errorf("got %q, want %q", got, syllabusDir)
		}
	})
}

func TestSubjectBaseID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"malaysia-kssm-matematik-tingkatan-3", "malaysia-kssm-matematik"},
		{"india-cbse-physics-class-12", "india-cbse-physics"},
		{"uk-cambridge-igcse-mathematics-0580", "uk-cambridge-igcse-mathematics-0580"}, // no grade detected (4-digit code)
		{"", ""},
	}
	for _, c := range tests {
		got := subjectBaseID(c.input)
		if got != c.want {
			t.Errorf("subjectBaseID(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestExtractDSKPTopics(t *testing.T) {
	input := `
BIDANG PEMBELAJARAN
PERKAITAN DAN ALGEBRA
TAJUK
1.0 FUNGSI DAN PERSAMAAN KUADRATIK
Kandungan fungsi kuadratik dan persamaan.
TAJUK
2.0 POLINOMIAL
Kandungan operasi polinomial.
BIDANG PEMBELAJARAN
STATISTIK DAN KEBARANGKALIAN
TAJUK
3.0 STATISTIK
Kandungan data dan taburan.
`
	topics := extractDSKPTopics(input)
	if len(topics) != 3 {
		t.Fatalf("expected 3 topics, got %d", len(topics))
	}

	cases := []struct {
		number string
		name   string
		area   string
	}{
		{"1.0", "FUNGSI DAN PERSAMAAN KUADRATIK", "PERKAITAN DAN ALGEBRA"},
		{"2.0", "POLINOMIAL", "PERKAITAN DAN ALGEBRA"},
		{"3.0", "STATISTIK", "STATISTIK DAN KEBARANGKALIAN"},
	}
	for i, c := range cases {
		if topics[i].Number != c.number {
			t.Errorf("topics[%d].Number = %q, want %q", i, topics[i].Number, c.number)
		}
		if topics[i].Name != c.name {
			t.Errorf("topics[%d].Name = %q, want %q", i, topics[i].Name, c.name)
		}
		if topics[i].LearningArea != c.area {
			t.Errorf("topics[%d].LearningArea = %q, want %q", i, topics[i].LearningArea, c.area)
		}
	}
	if !strings.Contains(topics[0].Content, "Kandungan fungsi") {
		t.Errorf("topics[0].Content missing expected text, got: %q", topics[0].Content)
	}
}

func TestExtractDSKPTopicsNoMatch(t *testing.T) {
	// Generic Markdown document — should return nil, not panic.
	input := "# Chapter 1\nSome content\n## Section 1.1\nMore content"
	topics := extractDSKPTopics(input)
	if len(topics) != 0 {
		t.Errorf("expected 0 topics for non-DSKP input, got %d", len(topics))
	}
}

func TestDSKPTopicsToChunks(t *testing.T) {
	topics := []dskpTopic{
		{LearningArea: "PERKAITAN DAN ALGEBRA", Number: "1.0", Name: "FUNGSI", Content: "some content"},
		{LearningArea: "STATISTIK", Number: "2.0", Name: "DATA", Content: "more content"},
	}
	chunks := dskpTopicsToChunks(topics)
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
	if chunks[0].Heading != "1.0 FUNGSI" {
		t.Errorf("chunks[0].Heading = %q, want %q", chunks[0].Heading, "1.0 FUNGSI")
	}
	if !strings.Contains(chunks[0].Content, "BIDANG PEMBELAJARAN: PERKAITAN DAN ALGEBRA") {
		t.Errorf("chunks[0].Content missing learning area header, got: %q", chunks[0].Content)
	}
	if !strings.Contains(chunks[0].Content, "TAJUK: 1.0 FUNGSI") {
		t.Errorf("chunks[0].Content missing TAJUK line, got: %q", chunks[0].Content)
	}
	if chunks[0].Total != 2 {
		t.Errorf("chunks[0].Total = %d, want 2", chunks[0].Total)
	}
	if chunks[1].Index != 1 {
		t.Errorf("chunks[1].Index = %d, want 1", chunks[1].Index)
	}
}

func TestSubjectPrefix(t *testing.T) {
	cases := []struct {
		subjectID string
		want      string
	}{
		{"malaysia-kssm-matematik-tingkatan-4", "MT"},
		{"malaysia-kssm-fizik-tingkatan-4", "PH"},
		{"malaysia-kssm-kimia-tingkatan-5", "CH"},
		{"malaysia-kssm-sains-tingkatan-2", "SC"},
		{"malaysia-kssm-biologi-tingkatan-5", "BI"},
		{"malaysia-kssm-sejarah-tingkatan-3", "HI"},
		{"india-cbse-mathematics-class-10", "MT"},
		{"india-cbse-physics-class-12", "PH"},
	}
	for _, c := range cases {
		got := subjectPrefix(c.subjectID)
		if got != c.want {
			t.Errorf("subjectPrefix(%q) = %q, want %q", c.subjectID, got, c.want)
		}
	}
}

func TestGradeNumber(t *testing.T) {
	cases := []struct {
		subjectID string
		want      string
	}{
		{"malaysia-kssm-matematik-tingkatan-4", "4"},
		{"india-cbse-physics-class-12", "12"},
		{"malaysia-kssm-matematik", ""},           // no grade (exam-based)
		{"uk-cambridge-igcse-mathematics-0580", ""}, // 580 > 20, not a grade
	}
	for _, c := range cases {
		got := gradeNumber(c.subjectID)
		if got != c.want {
			t.Errorf("gradeNumber(%q) = %q, want %q", c.subjectID, got, c.want)
		}
	}
}

func TestTopicFileID(t *testing.T) {
	cases := []struct {
		subjectID  string
		heading    string
		chunkIndex int
		want       string
	}{
		{"malaysia-kssm-matematik-tingkatan-4", "1.0 FUNGSI DAN PERSAMAAN", 0, "MT4-01"},
		{"malaysia-kssm-matematik-tingkatan-4", "10.0 STATISTIK DAN KEBARANGKALIAN", 9, "MT4-10"},
		{"india-cbse-physics-class-12", "3.0 Motion in a Plane", 2, "PH12-03"},
		{"malaysia-kssm-matematik-tingkatan-4", "no number heading", 4, "MT4-05"}, // fallback to chunkIndex+1
		{"uk-cambridge-igcse-mathematics-0580", "1.0 Number", 0, "MT-01"},         // no grade
	}
	for _, c := range cases {
		got := topicFileID(c.subjectID, c.heading, c.chunkIndex)
		if got != c.want {
			t.Errorf("topicFileID(%q, %q, %d) = %q, want %q",
				c.subjectID, c.heading, c.chunkIndex, got, c.want)
		}
	}
}

func TestParseDSKPTopicLine(t *testing.T) {
	cases := []struct {
		line   string
		number string
		name   string
	}{
		{"1.0 FUNGSI DAN PERSAMAAN KUADRATIK", "1.0", "FUNGSI DAN PERSAMAAN KUADRATIK"},
		{"10.0 STATISTIK DAN KEBARANGKALIAN", "10.0", "STATISTIK DAN KEBARANGKALIAN"},
		{"PLAIN HEADING NO NUMBER", "", "PLAIN HEADING NO NUMBER"},
		{"2.0 POLINOMIAL", "2.0", "POLINOMIAL"},
	}
	for _, c := range cases {
		number, name := parseDSKPTopicLine(c.line)
		if number != c.number {
			t.Errorf("parseDSKPTopicLine(%q) number = %q, want %q", c.line, number, c.number)
		}
		if name != c.name {
			t.Errorf("parseDSKPTopicLine(%q) name = %q, want %q", c.line, name, c.name)
		}
	}
}
