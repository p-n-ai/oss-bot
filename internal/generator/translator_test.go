package generator_test

import (
	"context"
	"strings"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"github.com/p-n-ai/oss-bot/internal/generator"
)

func TestTranslate(t *testing.T) {
	mockTranslation := `name: "Pemboleh ubah & Ungkapan Algebra"

learning_objectives:
  - id: LO1
    text: "Menggunakan huruf untuk mewakili kuantiti yang tidak diketahui"
`

	mock := ai.NewMockProvider(mockTranslation)

	topic := generator.Topic{
		ID:         "F1-01",
		Name:       "Variables & Algebraic Expressions",
		SyllabusID: "test-syllabus",
		Difficulty: "beginner",
		LearningObjectives: []generator.LearningObjective{
			{ID: "LO1", Text: "Use letters to represent unknown quantities", Bloom: "remember"},
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

func TestBuildTranslationPrompt(t *testing.T) {
	topic := generator.Topic{
		ID:         "F1-01",
		Name:       "Test Topic",
		SyllabusID: "test-syllabus",
		LearningObjectives: []generator.LearningObjective{
			{ID: "LO1", Text: "Test objective", Bloom: "understand"},
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
