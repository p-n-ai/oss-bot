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

func TestTranslate(t *testing.T) {
	mockTranslation := `name: "Pemboleh ubah & Ungkapan Algebra"

learning_objectives:
  - id: 1.0.1
    text: "Menggunakan huruf untuk mewakili kuantiti yang tidak diketahui"
`

	mock := ai.NewMockProvider(mockTranslation)

	topic := generator.Topic{
		ID:         "F1-01",
		Name:       "Variables & Algebraic Expressions",
		SyllabusID: "test-syllabus",
		Difficulty: "beginner",
		LearningObjectives: []generator.LearningObjective{
			{ID: "1.0.1", Text: "Use letters to represent unknown quantities", Bloom: "remember"},
		},
	}

	result, err := generator.Translate(context.Background(), mock, &topic, "ms")
	if err != nil {
		t.Fatalf("Translate() error = %v", err)
	}

	if !strings.Contains(result.Content, "Pemboleh ubah") {
		t.Error("Translation should contain BM terminology")
	}
}

func TestTranslate_UnsupportedLanguage(t *testing.T) {
	mock := ai.NewMockProvider("irrelevant")

	topic := generator.Topic{ID: "F1-01", Name: "Test"}

	_, err := generator.Translate(context.Background(), mock, &topic, "xx")
	if err == nil {
		t.Error("Translate() should error for unsupported language")
	}
}

func TestWriteTranslationFile(t *testing.T) {
	tmpDir := t.TempDir()

	content := "name: \"Garis Lurus\"\nlearning_objectives:\n  - id: 1.0.1\n    text: \"Menentukan kecerunan\"\n"

	err := generator.WriteTranslationFile(tmpDir, "ms", "MT3-09.yaml", content)
	if err != nil {
		t.Fatalf("WriteTranslationFile() error = %v", err)
	}

	// Verify file was written to translations/ms/MT3-09.yaml
	outPath := filepath.Join(tmpDir, "translations", "ms", "MT3-09.yaml")
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("translation file not created at %s: %v", outPath, err)
	}
	if !strings.Contains(string(data), "Garis Lurus") {
		t.Error("translation file should contain translated content")
	}
}

func TestWriteTranslationFile_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	err := generator.WriteTranslationFile(tmpDir, "zh", "PH4-01.yaml", "name: \"测试\"\n")
	if err != nil {
		t.Fatalf("WriteTranslationFile() error = %v", err)
	}

	// Verify directory structure: translations/zh/
	langDir := filepath.Join(tmpDir, "translations", "zh")
	info, err := os.Stat(langDir)
	if err != nil {
		t.Fatalf("translations/zh/ directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("translations/zh/ should be a directory")
	}
}

func TestWriteTranslationFile_CompanionFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Write multiple companion file translations
	files := map[string]string{
		"MT3-09.teaching.md":      "# Garis Lurus\n\nNota pengajaran...\n",
		"MT3-09.assessments.yaml": "assessments:\n  - id: Q1\n    question: \"Tentukan kecerunan\"\n",
		"MT3-09.examples.yaml":    "examples:\n  - id: WE-01\n    title: \"Contoh 1\"\n",
	}

	for fileName, content := range files {
		if err := generator.WriteTranslationFile(tmpDir, "ms", fileName, content); err != nil {
			t.Fatalf("WriteTranslationFile(%s) error = %v", fileName, err)
		}
	}

	// Verify all files exist in translations/ms/
	for fileName, expected := range files {
		outPath := filepath.Join(tmpDir, "translations", "ms", fileName)
		data, err := os.ReadFile(outPath)
		if err != nil {
			t.Errorf("companion file not created at %s: %v", outPath, err)
			continue
		}
		if string(data) != expected {
			t.Errorf("%s: content mismatch\ngot:  %q\nwant: %q", fileName, string(data), expected)
		}
	}
}

func TestTranslateFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a teaching notes file
	teachingFile := filepath.Join(tmpDir, "MT3-09.teaching.md")
	os.WriteFile(teachingFile, []byte("# Straight Lines\n\nTeaching notes here.\n"), 0o644)

	mockTranslation := "# Garis Lurus\n\nNota pengajaran di sini.\n"
	mock := ai.NewMockProvider(mockTranslation)

	result, err := generator.TranslateFile(context.Background(), mock, "MT3-09", teachingFile, "ms")
	if err != nil {
		t.Fatalf("TranslateFile() error = %v", err)
	}

	if !strings.Contains(result.Content, "Garis Lurus") {
		t.Error("TranslateFile should return translated content")
	}
}

func TestTranslateFile_UnsupportedLanguage(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "MT3-09.teaching.md")
	os.WriteFile(filePath, []byte("# Test"), 0o644)

	mock := ai.NewMockProvider("irrelevant")
	_, err := generator.TranslateFile(context.Background(), mock, "MT3-09", filePath, "xx")
	if err == nil {
		t.Error("TranslateFile() should error for unsupported language")
	}
}

func TestBuildTranslationPrompt(t *testing.T) {
	topic := generator.Topic{
		ID:         "F1-01",
		Name:       "Test Topic",
		SyllabusID: "test-syllabus",
		LearningObjectives: []generator.LearningObjective{
			{ID: "1.0.1", Text: "Test objective", Bloom: "understand"},
		},
	}

	prompt := generator.BuildTranslationPrompt(&topic, "Bahasa Melayu")
	if !strings.Contains(prompt, "Bahasa Melayu") {
		t.Error("Prompt should contain target language name")
	}
	if !strings.Contains(prompt, "F1-01") {
		t.Error("Prompt should contain topic ID")
	}
}
