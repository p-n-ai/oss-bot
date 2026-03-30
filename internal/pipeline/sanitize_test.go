package pipeline

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestSanitizeYAMLQuoting(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantSafe bool // true means the output should be parseable YAML with literal backslashes
	}{
		{
			name:     "no quotes unchanged",
			input:    "key: value\n",
			wantSafe: true,
		},
		{
			name:     "double quotes without backslash unchanged",
			input:    `key: "hello world"`,
			wantSafe: true,
		},
		{
			name:     "LaTeX \\text converted to single quotes",
			input:    `text: "Solve \\text{equation}"`,
			wantSafe: true,
		},
		{
			name:     "LaTeX \\sqrt converted to single quotes",
			input:    `text: "Find \\sqrt{x^2 + 1}"`,
			wantSafe: true,
		},
		{
			name:     "LaTeX \\frac converted to single quotes",
			input:    `text: "Simplify \\frac{a}{b}"`,
			wantSafe: true,
		},
		{
			name:     "LaTeX \\alpha converted",
			input:    `text: "The angle \\alpha"`,
			wantSafe: true,
		},
		{
			name:     "multiple LaTeX in one string",
			input:    `text: "If $\\sqrt{x} = \\frac{1}{2}$, find $\\text{x}$"`,
			wantSafe: true,
		},
		{
			name:     "escaped backslash not converted",
			input:    `text: "a \\\\ b"`,
			wantSafe: true,
		},
		{
			name:     "single quotes in content are doubled",
			input:    `text: "student's \\text{answer}"`,
			wantSafe: true,
		},
		{
			name:     "multiline YAML with mixed quoting",
			input: `topic_id: MT1-01
questions:
  - id: Q1
    text: "Solve $\\text{the equation}$"
    difficulty: easy
    answer:
      type: exact
      value: "42"`,
			wantSafe: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeYAMLQuoting(tt.input)

			if tt.wantSafe {
				// Verify the output is valid YAML
				var node yaml.Node
				if err := yaml.Unmarshal([]byte(got), &node); err != nil {
					t.Errorf("SanitizeYAMLQuoting() produced invalid YAML:\nInput:  %s\nOutput: %s\nError:  %v", tt.input, got, err)
				}
			}
		})
	}
}

func TestSanitizeYAMLQuoting_BackslashLetter(t *testing.T) {
	// Verify that backslash+letter sequences are preserved literally
	input := `text: "\\text{hello}"`
	got := SanitizeYAMLQuoting(input)

	// Should be single-quoted now
	if got != `text: '\\text{hello}'` {
		t.Errorf("expected single-quoted output, got: %s", got)
	}

	// Parse and verify the value contains literal backslash
	var data map[string]string
	if err := yaml.Unmarshal([]byte(got), &data); err != nil {
		t.Fatalf("failed to parse sanitized YAML: %v", err)
	}
	if data["text"] != `\\text{hello}` {
		t.Errorf("expected literal backslash preserved, got: %q", data["text"])
	}
}

func TestSanitizeYAMLQuoting_PreservesValidDoubleQuotes(t *testing.T) {
	// Double-quoted strings without backslash+letter should be unchanged
	input := `key: "normal value"`
	got := SanitizeYAMLQuoting(input)

	if got != input {
		t.Errorf("expected unchanged output %q, got %q", input, got)
	}
}

func TestSanitizeYAMLQuoting_FullAssessmentYAML(t *testing.T) {
	input := `topic_id: MT1-01
provenance: ai-generated

questions:
  - id: Q1
    text: "Solve $\\sqrt{x^2 + 4} = 3$"
    difficulty: medium
    learning_objective: LO1
    answer:
      type: exact
      value: "\\sqrt{5}"
      working: |
        Step 1: Square both sides
        Step 2: Solve for x
    marks: 3
  - id: Q2
    text: "What is the value of $\\frac{\\text{numerator}}{\\text{denominator}}$?"
    difficulty: hard
    answer:
      type: free_text
      value: "The fraction simplifies to 1"
`

	got := SanitizeYAMLQuoting(input)

	// Must be valid YAML
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(got), &node); err != nil {
		t.Fatalf("sanitized YAML is invalid: %v\n\nOutput:\n%s", err, got)
	}

	// Verify LaTeX strings are now single-quoted (not double-quoted)
	if contains(got, `"Solve $\\sqrt`) {
		t.Error("LaTeX string should have been converted to single quotes")
	}
	if contains(got, `"\\sqrt{5}"`) {
		t.Error("LaTeX answer should have been converted to single quotes")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
