package generator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"gopkg.in/yaml.v3"
)

// TopicEnrichment holds the structured Level 2 fields extracted by AI.
type TopicEnrichment struct {
	Teaching        TeachingInfo `yaml:"teaching"`
	EngagementHooks []string    `yaml:"engagement_hooks"`
}

// GenerateTopicEnrichment uses AI to produce structured Level 2 fields
// (teaching.sequence, teaching.common_misconceptions, engagement_hooks)
// for a topic. It reads any existing teaching notes markdown to inform
// the extraction.
func GenerateTopicEnrichment(ctx context.Context, provider ai.Provider, genCtx *GenerationContext) (*GenerationResult, error) {
	prompt := BuildTopicEnrichmentPrompt(genCtx)

	resp, err := provider.Complete(ctx, ai.CompletionRequest{
		Messages: []ai.Message{
			{Role: "system", Content: "You are an expert educator. Output valid YAML only — no markdown fences, no explanatory text."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   2048,
		Temperature: 0.5,
	})
	if err != nil {
		return nil, fmt.Errorf("AI enrichment failed: %w", err)
	}

	return &GenerationResult{
		Content:      resp.Content,
		Model:        resp.Model,
		InputTokens:  resp.InputTokens,
		OutputTokens: resp.OutputTokens,
	}, nil
}

// BuildTopicEnrichmentPrompt constructs the prompt for topic YAML enrichment.
func BuildTopicEnrichmentPrompt(genCtx *GenerationContext) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Generate structured teaching metadata for the topic: %s (%s)\n", genCtx.Topic.Name, genCtx.Topic.ID))
	sb.WriteString(fmt.Sprintf("Subject: %s\n", genCtx.Topic.SubjectID))
	sb.WriteString(fmt.Sprintf("Syllabus: %s\n", genCtx.Topic.SyllabusID))
	sb.WriteString(fmt.Sprintf("Difficulty: %s\n\n", genCtx.Topic.Difficulty))

	sb.WriteString("## Learning Objectives\n")
	for _, lo := range genCtx.Topic.LearningObjectives {
		sb.WriteString(fmt.Sprintf("- %s (%s): %s\n", lo.ID, lo.Bloom, lo.Text))
	}
	sb.WriteString("\n")

	// Include existing teaching notes markdown if available for context
	teachingNotes := genCtx.ExistingNotes
	if teachingNotes == "" {
		// Try to load from the companion file
		notesFile := filepath.Join(genCtx.TopicDir, genCtx.Topic.ID+".teaching.md")
		if data, err := os.ReadFile(notesFile); err == nil {
			teachingNotes = string(data)
		}
	}
	if teachingNotes != "" {
		sb.WriteString("## Existing Teaching Notes (use these to inform your output)\n")
		sb.WriteString(teachingNotes)
		sb.WriteString("\n\n")
	}

	if genCtx.SchemaRules != "" {
		sb.WriteString("## JSON Schema (your output MUST conform to this schema)\n")
		sb.WriteString("```json\n")
		sb.WriteString(genCtx.SchemaRules)
		sb.WriteString("\n```\n\n")
	}

	sb.WriteString(`## Output Format
Output ONLY valid YAML with exactly this structure:

teaching:
  sequence:
    - "Step 1: Brief description of teaching step (XX min)"
    - "Step 2: Brief description of teaching step (XX min)"
    - "Step 3: Brief description of teaching step (XX min)"
  common_misconceptions:
    - misconception: "What students wrongly believe"
      remediation: "How to correct this misunderstanding"
    - misconception: "Another common error"
      remediation: "How to fix it"
engagement_hooks:
  - "Real-world connection or question to engage students"
  - "Another engaging scenario or question"

Rules:
- teaching.sequence: 3-6 steps, each a short string with approximate duration
- teaching.common_misconceptions: 2-4 entries, each with misconception and remediation
- engagement_hooks: 2-4 culturally relevant real-world connections
- Output ONLY the YAML — no markdown fences, no extra text
`)

	return sb.String()
}

// EnrichTopicYAML reads the existing topic YAML file, merges in the Level 2
// enrichment fields, and returns the updated YAML content. It preserves all
// existing fields and only adds/overwrites teaching and engagement_hooks.
func EnrichTopicYAML(topicFilePath string, enrichmentYAML string) (string, error) {
	// Parse the enrichment output
	var enrichment TopicEnrichment
	if err := yaml.Unmarshal([]byte(enrichmentYAML), &enrichment); err != nil {
		return "", fmt.Errorf("parsing enrichment YAML: %w", err)
	}

	// Read and parse the existing topic file as a generic map to preserve all fields
	data, err := os.ReadFile(topicFilePath)
	if err != nil {
		return "", fmt.Errorf("reading topic file: %w", err)
	}

	var raw yaml.Node
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return "", fmt.Errorf("parsing topic YAML: %w", err)
	}

	// Work with the mapping node (raw.Content[0] is the document's mapping)
	if raw.Kind != yaml.DocumentNode || len(raw.Content) == 0 {
		return "", fmt.Errorf("unexpected YAML structure")
	}
	mapping := raw.Content[0]
	if mapping.Kind != yaml.MappingNode {
		return "", fmt.Errorf("expected mapping node, got %d", mapping.Kind)
	}

	// Marshal enrichment fields to yaml.Node for insertion
	teachingNode, err := yamlValueNode(enrichment.Teaching)
	if err != nil {
		return "", fmt.Errorf("marshaling teaching: %w", err)
	}
	hooksNode, err := yamlValueNode(enrichment.EngagementHooks)
	if err != nil {
		return "", fmt.Errorf("marshaling engagement_hooks: %w", err)
	}

	// Update or insert "teaching" and "engagement_hooks" in the mapping
	setMappingKey(mapping, "teaching", teachingNode)
	setMappingKey(mapping, "engagement_hooks", hooksNode)

	// Ensure mastery, ai_teaching_notes, and assessments_file are present.
	// These may have been dropped by the AI during import.
	topicID := getMappingValue(mapping, "id")
	if topicID != "" {
		if !hasMappingKey(mapping, "mastery") {
			masteryYAML := "minimum_score: 0.75\nassessment_count: 3\nspaced_repetition:\n  initial_interval_days: 3\n  multiplier: 2.5"
			var masteryNode yaml.Node
			if err := yaml.Unmarshal([]byte(masteryYAML), &masteryNode); err == nil && masteryNode.Kind == yaml.DocumentNode && len(masteryNode.Content) > 0 {
				setMappingKey(mapping, "mastery", masteryNode.Content[0])
			}
		}
		if !hasMappingKey(mapping, "ai_teaching_notes") {
			setMappingKey(mapping, "ai_teaching_notes", &yaml.Node{Kind: yaml.ScalarNode, Value: topicID + ".teaching.md", Tag: "!!str"})
		}
		if !hasMappingKey(mapping, "assessments_file") {
			setMappingKey(mapping, "assessments_file", &yaml.Node{Kind: yaml.ScalarNode, Value: topicID + ".assessments.yaml", Tag: "!!str"})
		}
	}

	// Marshal back
	out, err := yaml.Marshal(&raw)
	if err != nil {
		return "", fmt.Errorf("marshaling updated YAML: %w", err)
	}

	return string(out), nil
}

// setMappingKey sets or replaces a key in a yaml.MappingNode.
func setMappingKey(mapping *yaml.Node, key string, value *yaml.Node) {
	for i := 0; i < len(mapping.Content)-1; i += 2 {
		if mapping.Content[i].Value == key {
			mapping.Content[i+1] = value
			return
		}
	}
	// Key not found — append
	mapping.Content = append(mapping.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key, Tag: "!!str"},
		value,
	)
}

// hasMappingKey checks whether a yaml.MappingNode contains a given key.
func hasMappingKey(mapping *yaml.Node, key string) bool {
	for i := 0; i < len(mapping.Content)-1; i += 2 {
		if mapping.Content[i].Value == key {
			return true
		}
	}
	return false
}

// getMappingValue returns the scalar value for a key in a yaml.MappingNode,
// or empty string if not found or not a scalar.
func getMappingValue(mapping *yaml.Node, key string) string {
	for i := 0; i < len(mapping.Content)-1; i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1].Value
		}
	}
	return ""
}

// yamlValueNode marshals a Go value into a yaml.Node suitable for insertion.
func yamlValueNode(v interface{}) (*yaml.Node, error) {
	data, err := yaml.Marshal(v)
	if err != nil {
		return nil, err
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
		return doc.Content[0], nil
	}
	return &doc, nil
}
