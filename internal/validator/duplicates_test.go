package validator_test

import (
	"testing"

	"github.com/p-n-ai/oss-bot/internal/validator"
)

func TestTokenize(t *testing.T) {
	text := "Solve the equation 2x + 3 = 7"
	tokens := validator.Tokenize(text)
	if len(tokens) == 0 {
		t.Error("Tokenize() returned empty")
	}
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name      string
		a, b      string
		threshold float64
		similar   bool
	}{
		{
			name:      "identical",
			a:         "Solve the equation 2x + 3 = 7",
			b:         "Solve the equation 2x + 3 = 7",
			threshold: 0.85,
			similar:   true,
		},
		{
			name:      "very-different",
			a:         "Solve the equation 2x + 3 = 7",
			b:         "Describe the process of photosynthesis in plants",
			threshold: 0.85,
			similar:   false,
		},
		{
			name:      "similar-but-different",
			a:         "Find the value of x in 3x + 5 = 20",
			b:         "Find the value of x in 4x - 2 = 14",
			threshold: 0.85,
			similar:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sim := validator.CosineSimilarity(tt.a, tt.b)
			got := sim >= tt.threshold
			if got != tt.similar {
				t.Errorf("CosineSimilarity(%q, %q) = %f, similar=%v want %v",
					tt.a, tt.b, sim, got, tt.similar)
			}
		})
	}
}

func TestFindDuplicates(t *testing.T) {
	questions := []string{
		"Solve 2x + 3 = 7",
		"Simplify 3a + 2b - a",
		"Solve 2x + 3 = 7", // duplicate of first
	}

	dupes := validator.FindDuplicates(questions, 0.85)
	if len(dupes) == 0 {
		t.Error("FindDuplicates() found no duplicates, expected at least one pair")
	}
}
