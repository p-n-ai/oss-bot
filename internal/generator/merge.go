package generator

import (
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/p-n-ai/oss-bot/internal/validator"
)

// Assessment is a parsed assessment question used for content merging.
type Assessment struct {
	ID         string `yaml:"id"`
	Text       string `yaml:"text"`
	Difficulty string `yaml:"difficulty"`
}

// Example is a parsed worked example used for content merging.
type Example struct {
	ID         string `yaml:"id"`
	Topic      string `yaml:"topic"`
	Difficulty string `yaml:"difficulty"`
	Scenario   string `yaml:"scenario,omitempty"`
}

// TeachingNotes is Markdown content representing teaching notes.
type TeachingNotes = string

// MergeReport summarises what changed during a content merge, used in PR descriptions.
type MergeReport struct {
	Added    int      // Number of new items added
	Skipped  int      // Number of duplicates skipped
	Sections []string // Section names added (teaching notes only)
}

// String returns a human-readable one-line summary.
func (r MergeReport) String() string {
	if r.Added == 0 && r.Skipped == 0 {
		return "no changes (content identical)"
	}
	return fmt.Sprintf("added %d, skipped %d duplicate(s)", r.Added, r.Skipped)
}

// skipThreshold is the cosine similarity above which an item is treated as a duplicate.
const skipThreshold = 0.95

// MergeAssessments appends new assessments to existing ones, skipping near-duplicates
// (>95% cosine similarity on question text).
func MergeAssessments(existing, generated []Assessment) ([]Assessment, MergeReport) {
	report := MergeReport{}

	existingTexts := make([]string, len(existing))
	for i, a := range existing {
		existingTexts[i] = a.Text
	}

	result := make([]Assessment, len(existing))
	copy(result, existing)

	for _, candidate := range generated {
		if isDuplicate(candidate.Text, existingTexts) {
			report.Skipped++
			continue
		}
		result = append(result, candidate)
		existingTexts = append(existingTexts, candidate.Text)
		report.Added++
	}

	return result, report
}

// MergeExamples appends new examples to existing ones, skips near-duplicates, then
// re-sorts the full list by difficulty (easy → medium → hard).
func MergeExamples(existing, generated []Example) ([]Example, MergeReport) {
	report := MergeReport{}

	existingKeys := make([]string, len(existing))
	for i, e := range existing {
		existingKeys[i] = exampleKey(e)
	}

	result := make([]Example, len(existing))
	copy(result, existing)

	for _, candidate := range generated {
		key := exampleKey(candidate)
		if isDuplicate(key, existingKeys) {
			report.Skipped++
			continue
		}
		result = append(result, candidate)
		existingKeys = append(existingKeys, key)
		report.Added++
	}

	sort.SliceStable(result, func(i, j int) bool {
		return difficultyOrder(result[i].Difficulty) < difficultyOrder(result[j].Difficulty)
	})

	return result, report
}

// MergeTeachingNotes additively merges Markdown teaching notes: all existing sections
// are kept unchanged; sections from generated that are not already present are appended.
func MergeTeachingNotes(existing, generated TeachingNotes) (TeachingNotes, MergeReport) {
	if existing == "" {
		return generated, MergeReport{}
	}

	// Build set of H2 section names already in existing.
	existingNames := h2SectionNames(existing)

	// Extract ordered H2 sections from generated.
	genSections := splitH2Sections(generated)

	report := MergeReport{}
	var result strings.Builder
	result.WriteString(existing)

	for _, sec := range genSections {
		if existingNames[sec.name] {
			continue // Section already exists — keep existing, do not overwrite.
		}
		// Ensure separation from previous content.
		s := result.String()
		if len(s) > 0 && !strings.HasSuffix(s, "\n\n") {
			if strings.HasSuffix(s, "\n") {
				result.WriteString("\n")
			} else {
				result.WriteString("\n\n")
			}
		}
		result.WriteString(sec.raw)
		report.Added++
		report.Sections = append(report.Sections, sec.name)
	}

	return result.String(), report
}

// MergeAssessmentsYAML merges two YAML assessment documents, preserving all fields.
// Questions with >95% similar text are treated as duplicates and skipped.
func MergeAssessmentsYAML(existingYAML, generatedYAML string) (string, MergeReport, error) {
	type doc struct {
		TopicID    string                   `yaml:"topic_id,omitempty"`
		Provenance string                   `yaml:"provenance,omitempty"`
		Questions  []map[string]interface{} `yaml:"questions"`
	}

	var existing, generated doc
	if err := yaml.Unmarshal([]byte(existingYAML), &existing); err != nil {
		return generatedYAML, MergeReport{}, fmt.Errorf("parsing existing assessments: %w", err)
	}
	if err := yaml.Unmarshal([]byte(generatedYAML), &generated); err != nil {
		return generatedYAML, MergeReport{}, fmt.Errorf("parsing generated assessments: %w", err)
	}

	existingTexts := make([]string, 0, len(existing.Questions))
	for _, q := range existing.Questions {
		existingTexts = append(existingTexts, stringField(q, "text"))
	}

	report := MergeReport{}
	merged := make([]map[string]interface{}, len(existing.Questions))
	copy(merged, existing.Questions)

	for _, q := range generated.Questions {
		text := stringField(q, "text")
		if isDuplicate(text, existingTexts) {
			report.Skipped++
			continue
		}
		merged = append(merged, q)
		existingTexts = append(existingTexts, text)
		report.Added++
	}

	existing.Questions = merged
	data, err := yaml.Marshal(existing)
	if err != nil {
		return generatedYAML, report, fmt.Errorf("serializing merged assessments: %w", err)
	}
	return string(data), report, nil
}

// MergeExamplesYAML merges two YAML examples documents, preserving all fields.
// Examples with >95% similar topic+scenario text are treated as duplicates.
func MergeExamplesYAML(existingYAML, generatedYAML string) (string, MergeReport, error) {
	type doc struct {
		TopicID        string                   `yaml:"topic_id,omitempty"`
		Provenance     string                   `yaml:"provenance,omitempty"`
		Description    string                   `yaml:"description,omitempty"`
		WorkedExamples []map[string]interface{} `yaml:"worked_examples"`
	}

	var existing, generated doc
	if err := yaml.Unmarshal([]byte(existingYAML), &existing); err != nil {
		return generatedYAML, MergeReport{}, fmt.Errorf("parsing existing examples: %w", err)
	}
	if err := yaml.Unmarshal([]byte(generatedYAML), &generated); err != nil {
		return generatedYAML, MergeReport{}, fmt.Errorf("parsing generated examples: %w", err)
	}

	existingKeys := make([]string, 0, len(existing.WorkedExamples))
	for _, e := range existing.WorkedExamples {
		existingKeys = append(existingKeys, exampleMapKey(e))
	}

	report := MergeReport{}
	merged := make([]map[string]interface{}, len(existing.WorkedExamples))
	copy(merged, existing.WorkedExamples)

	for _, e := range generated.WorkedExamples {
		key := exampleMapKey(e)
		if isDuplicate(key, existingKeys) {
			report.Skipped++
			continue
		}
		merged = append(merged, e)
		existingKeys = append(existingKeys, key)
		report.Added++
	}

	// Re-sort by difficulty.
	sort.SliceStable(merged, func(i, j int) bool {
		return difficultyOrder(stringField(merged[i], "difficulty")) <
			difficultyOrder(stringField(merged[j], "difficulty"))
	})

	existing.WorkedExamples = merged
	data, err := yaml.Marshal(existing)
	if err != nil {
		return generatedYAML, report, fmt.Errorf("serializing merged examples: %w", err)
	}
	return string(data), report, nil
}

// --- helpers ---

// markdownSection represents a single H2 section from a Markdown document.
type markdownSection struct {
	name string
	raw  string
}

// h2SectionNames returns the set of H2 section names in a Markdown document.
func h2SectionNames(md string) map[string]bool {
	names := make(map[string]bool)
	for _, line := range strings.Split(md, "\n") {
		if strings.HasPrefix(line, "## ") {
			names[strings.TrimSpace(strings.TrimPrefix(line, "## "))] = true
		}
	}
	return names
}

// splitH2Sections splits a Markdown document into its H2 sections (in order).
// Text before the first H2 heading is ignored.
func splitH2Sections(md string) []markdownSection {
	var sections []markdownSection
	var current *markdownSection

	for _, line := range strings.Split(md, "\n") {
		if strings.HasPrefix(line, "## ") {
			if current != nil {
				sections = append(sections, *current)
			}
			name := strings.TrimSpace(strings.TrimPrefix(line, "## "))
			current = &markdownSection{name: name, raw: line + "\n"}
		} else if current != nil {
			current.raw += line + "\n"
		}
	}
	if current != nil {
		sections = append(sections, *current)
	}
	return sections
}

// isDuplicate returns true if text is ≥95% similar to any item in the corpus.
func isDuplicate(text string, corpus []string) bool {
	for _, existing := range corpus {
		if validator.CosineSimilarity(text, existing) >= skipThreshold {
			return true
		}
	}
	return false
}

// exampleKey creates a text key for Example similarity comparison.
func exampleKey(e Example) string {
	return e.Topic + " " + e.Scenario
}

// exampleMapKey creates a text key for map-based example similarity comparison.
func exampleMapKey(e map[string]interface{}) string {
	return stringField(e, "topic") + " " + stringField(e, "scenario")
}

// stringField safely extracts a string value from a map.
func stringField(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// difficultyOrder maps difficulty strings to sort order (easy=0, medium=1, hard=2).
func difficultyOrder(d string) int {
	switch strings.ToLower(d) {
	case "easy":
		return 0
	case "medium":
		return 1
	case "hard":
		return 2
	default:
		return 3
	}
}
