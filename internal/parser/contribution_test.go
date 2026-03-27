package parser_test

import (
	"context"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"github.com/p-n-ai/oss-bot/internal/parser"
)

func TestParseContribution_ValidatesInput(t *testing.T) {
	tests := []struct {
		name    string
		input   parser.ContributionInput
		wantErr bool
	}{
		{
			name: "empty teacher input",
			input: parser.ContributionInput{
				TeacherInput: "",
				TopicPath:    "math/01",
				ContentType:  "misconception",
			},
			wantErr: true,
		},
		{
			name: "empty topic path",
			input: parser.ContributionInput{
				TeacherInput: "Students confuse X with Y",
				TopicPath:    "",
				ContentType:  "misconception",
			},
			wantErr: true,
		},
		{
			name: "empty content type",
			input: parser.ContributionInput{
				TeacherInput: "Students confuse X with Y",
				TopicPath:    "math/01",
				ContentType:  "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.ParseContribution(context.Background(), nil, tt.input, "")
			if tt.wantErr && err == nil {
				t.Error("expected validation error, got nil")
			}
		})
	}
}

func TestParseContribution_WithMockProvider(t *testing.T) {
	mockYAML := `misconception:
  description: "Students write -x+2 instead of -x-2 when expanding -(x+2)"
  cause: "Distributing only the leading sign, ignoring the inner plus"
  correction: "Stress that the minus distributes to every term inside"
  example:
    incorrect: "-(x+2) = -x+2"
    correct: "-(x+2) = -x-2"`

	mock := ai.NewMockProvider(mockYAML)

	input := parser.ContributionInput{
		TeacherInput: "My students always confuse the negative sign when expanding brackets like -(x+2). They write -x+2 instead of -x-2.",
		TopicPath:    "mathematics/algebra/03-expanding-brackets",
		ContentType:  "misconception",
		SyllabusID:   "malaysia-kssm",
	}

	result, err := parser.ParseContribution(context.Background(), mock, input, "testdata/")
	if err != nil {
		t.Fatalf("ParseContribution() error = %v", err)
	}
	if result == "" {
		t.Error("ParseContribution() returned empty result")
	}
}

func TestContributionInput_Fields(t *testing.T) {
	validTypes := []string{"misconception", "teaching_note", "example"}
	for _, ct := range validTypes {
		t.Run(ct, func(t *testing.T) {
			input := parser.ContributionInput{
				TeacherInput: "some teacher observation",
				TopicPath:    "math/01",
				ContentType:  ct,
				SyllabusID:   "test-syllabus",
			}
			if input.TeacherInput == "" || input.TopicPath == "" || input.ContentType == "" {
				t.Error("valid input fields must not be empty")
			}
		})
	}
}
