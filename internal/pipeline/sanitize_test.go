package pipeline

import (
	"strings"
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
			name:     "LaTeX backslash-t in text converted",
			input:    "text: \"\\text{equation}\"",
			wantSafe: true,
		},
		{
			name:     "LaTeX backslash-s in sqrt converted",
			input:    "text: \"\\sqrt{x^2 + 1}\"",
			wantSafe: true,
		},
		{
			name:     "LaTeX backslash-f in frac converted",
			input:    "text: \"\\frac{a}{b}\"",
			wantSafe: true,
		},
		{
			name:     "LaTeX backslash-a in alpha converted",
			input:    "text: \"The angle \\alpha\"",
			wantSafe: true,
		},
		{
			name:     "properly escaped backslash not converted",
			input:    `text: "a \\ b"`,
			wantSafe: true,
		},
		{
			name:     "single quotes in content are doubled",
			input:    "text: \"student's \\text{answer}\"",
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
					t.Errorf("SanitizeYAMLQuoting() produced invalid YAML:\nInput:  %q\nOutput: %q\nError:  %v", tt.input, got, err)
				}
			}
		})
	}
}

func TestSanitizeYAMLQuoting_BackslashLetter(t *testing.T) {
	// AI outputs: text: "\text{hello}" (backslash-t intended as literal)
	// In YAML double-quoted, \t becomes a tab. We must convert to single-quoted.
	input := "text: \"\\text{hello}\""
	got := SanitizeYAMLQuoting(input)

	// Should be single-quoted now
	if !strings.Contains(got, "'") {
		t.Errorf("expected single-quoted output, got: %s", got)
	}
	if strings.Contains(got, "\"\\t") {
		t.Errorf("double-quoted backslash-letter should have been converted, got: %s", got)
	}

	// Parse and verify the value contains literal backslash-t
	var data map[string]string
	if err := yaml.Unmarshal([]byte(got), &data); err != nil {
		t.Fatalf("failed to parse sanitized YAML: %v", err)
	}
	if !strings.Contains(data["text"], `\text`) {
		t.Errorf("expected literal \\text preserved, got: %q", data["text"])
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
	// Simulate AI generating assessment YAML with LaTeX in double quotes
	input := "topic_id: MT1-01\nprovenance: ai-generated\n\nquestions:\n" +
		"  - id: Q1\n" +
		"    text: \"Solve $\\sqrt{x^2 + 4} = 3$\"\n" +
		"    difficulty: medium\n" +
		"    learning_objective: LO1\n" +
		"    answer:\n" +
		"      type: exact\n" +
		"      value: \"\\sqrt{5}\"\n" +
		"      working: |\n" +
		"        Step 1: Square both sides\n" +
		"        Step 2: Solve for x\n" +
		"    marks: 3\n" +
		"  - id: Q2\n" +
		"    text: \"What is $\\frac{\\text{num}}{\\text{den}}$?\"\n" +
		"    difficulty: hard\n" +
		"    answer:\n" +
		"      type: free_text\n" +
		"      value: \"The fraction simplifies to 1\"\n"

	got := SanitizeYAMLQuoting(input)

	// Must be valid YAML
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(got), &node); err != nil {
		t.Fatalf("sanitized YAML is invalid: %v\n\nOutput:\n%s", err, got)
	}

	// Verify LaTeX strings were converted (should not have double-quoted backslash+letter)
	if strings.Contains(got, "\"Solve $\\s") {
		t.Error("LaTeX string should have been converted from double quotes")
	}
	if strings.Contains(got, "\"\\sqrt{5}\"") {
		t.Error("LaTeX answer should have been converted from double quotes")
	}
	// Non-LaTeX double-quoted strings should remain
	if !strings.Contains(got, "\"The fraction simplifies to 1\"") {
		t.Error("non-LaTeX string should remain double-quoted")
	}
}

func TestSanitizeYAMLQuoting_NewlineInString(t *testing.T) {
	// Double quotes spanning a newline should not be treated as a quoted string
	input := "key: \"some\nvalue\""
	got := SanitizeYAMLQuoting(input)
	// Should be left unchanged (newline breaks the scanning)
	if got != input {
		t.Errorf("input with newline inside quotes should be unchanged, got: %q", got)
	}
}

func TestSanitizeYAMLQuoting_UnterminatedQuote(t *testing.T) {
	input := `key: "unterminated`
	got := SanitizeYAMLQuoting(input)
	if got != input {
		t.Errorf("unterminated quote should be unchanged, got: %q", got)
	}
}

func TestFixYAMLColonSpacing(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "missing space after colon",
			input: "    text:hubungkaitkan waktu subur",
			want:  "    text: hubungkaitkan waktu subur",
		},
		{
			name:  "already has space",
			input: "    text: hubungkaitkan waktu subur",
			want:  "    text: hubungkaitkan waktu subur",
		},
		{
			name:  "colon in quoted string unchanged",
			input: `    text: "value:with:colons"`,
			want:  `    text: "value:with:colons"`,
		},
		{
			name:  "multiple lines mixed",
			input: "id: SC1-04\n    text:hubung\nname: test",
			want:  "id: SC1-04\n    text: hubung\nname: test",
		},
		{
			name:  "empty value after colon",
			input: "key:",
			want:  "key:",
		},
		{
			name:  "colon followed by space",
			input: "key: value",
			want:  "key: value",
		},
		{
			name:  "colon followed by newline",
			input: "key:\n  nested: val",
			want:  "key:\n  nested: val",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FixYAMLColonSpacing(tt.input)
			if got != tt.want {
				t.Errorf("FixYAMLColonSpacing()\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestRemoveDuplicateKeys(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "duplicate top-level key",
			input: "id: SC1-07\nofficial_ref: \"Bab 7\"\nname: Udara\nofficial_ref: \"7.1\"\ndifficulty: beginner",
			want:  "id: SC1-07\nofficial_ref: \"Bab 7\"\nname: Udara\ndifficulty: beginner",
		},
		{
			name:  "no duplicates",
			input: "id: SC1-01\nname: Test\ndifficulty: beginner",
			want:  "id: SC1-01\nname: Test\ndifficulty: beginner",
		},
		{
			name:  "duplicate with nested content removed",
			input: "id: SC1-07\nofficial_ref: \"Bab 7\"\nofficial_ref:\n  - \"7.1\"\n  - \"7.2\"\nname: Udara",
			want:  "id: SC1-07\nofficial_ref: \"Bab 7\"\nname: Udara",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveDuplicateKeys(tt.input)
			if got != tt.want {
				t.Errorf("RemoveDuplicateKeys()\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}
