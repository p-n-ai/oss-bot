// Package generator implements the AI content generation pipeline.
package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Topic represents a parsed topic YAML file.
type Topic struct {
	ID                 string              `yaml:"id"`
	Name               string              `yaml:"name"`
	SubjectID          string              `yaml:"subject_id"`
	SyllabusID         string              `yaml:"syllabus_id"`
	Difficulty         string              `yaml:"difficulty"`
	Tier               string              `yaml:"tier,omitempty"`
	LearningObjectives []LearningObjective `yaml:"learning_objectives"`
	Prerequisites      PrerequisiteList    `yaml:"prerequisites"`
	Teaching           *TeachingInfo       `yaml:"teaching,omitempty"`
	BloomLevels        []string            `yaml:"bloom_levels,omitempty"`
	Mastery            *MasteryInfo        `yaml:"mastery,omitempty"`
	QualityLevel       int                 `yaml:"quality_level"`
	Provenance         string              `yaml:"provenance"`
	TeachingNotesFile  string              `yaml:"ai_teaching_notes,omitempty"`
	ExamplesFile       string              `yaml:"examples_file,omitempty"`
	AssessmentsFile    string              `yaml:"assessments_file,omitempty"`
}

// MasteryInfo holds mastery/spaced-repetition configuration.
type MasteryInfo struct {
	MinimumScore    float64          `yaml:"minimum_score"`
	AssessmentCount int              `yaml:"assessment_count"`
	SpacedRepetition *SpacedRepetition `yaml:"spaced_repetition,omitempty"`
}

// SpacedRepetition holds spaced repetition scheduling parameters.
type SpacedRepetition struct {
	InitialIntervalDays int     `yaml:"initial_interval_days"`
	Multiplier          float64 `yaml:"multiplier"`
}

// LearningObjective represents a single learning objective.
type LearningObjective struct {
	ID    string `yaml:"id"`
	Text  string `yaml:"text"`
	Bloom string `yaml:"bloom"`
}

// PrerequisiteList holds required and recommended prerequisites.
type PrerequisiteList struct {
	Required    []string `yaml:"required"`
	Recommended []string `yaml:"recommended"`
}

// TeachingInfo holds teaching-related content.
type TeachingInfo struct {
	Sequence             []string        `yaml:"sequence,omitempty"`
	CommonMisconceptions []Misconception `yaml:"common_misconceptions,omitempty"`
	EngagementHooks      []string        `yaml:"engagement_hooks,omitempty"`
}

// Misconception represents a common student misconception.
type Misconception struct {
	Misconception string `yaml:"misconception"`
	Remediation   string `yaml:"remediation"`
}

// GenerationContext holds all context needed for AI content generation.
type GenerationContext struct {
	Topic              Topic
	TopicDir           string  // Absolute path of the directory containing the topic file.
	Prerequisites      []Topic
	Siblings           []Topic
	ExistingNotes      string
	SchemaRules        string
	ValidationFeedback []string // Populated on retry after validation failure
}

// BuildContext assembles the generation context for a given topic ID.
func BuildContext(repoDir, topicID string) (*GenerationContext, error) {
	// Find the topic file
	topicFile, err := FindTopicFile(repoDir, topicID)
	if err != nil {
		return nil, fmt.Errorf("finding topic %s: %w", topicID, err)
	}

	topic, err := loadTopic(topicFile)
	if err != nil {
		return nil, fmt.Errorf("loading topic %s: %w", topicID, err)
	}

	ctx := &GenerationContext{
		Topic:    *topic,
		TopicDir: filepath.Dir(topicFile),
	}

	// Load prerequisites
	for _, prereqID := range topic.Prerequisites.Required {
		prereqFile, err := FindTopicFile(repoDir, prereqID)
		if err != nil {
			continue // Prerequisite might not exist yet
		}
		prereq, err := loadTopic(prereqFile)
		if err != nil {
			continue
		}
		ctx.Prerequisites = append(ctx.Prerequisites, *prereq)
	}

	// Load siblings (other topics in the same directory)
	topicDir := filepath.Dir(topicFile)
	siblings, err := loadSiblingTopics(topicDir, topicID)
	if err == nil {
		ctx.Siblings = siblings
	}

	// Load existing teaching notes if they exist
	if topic.TeachingNotesFile != "" {
		notesPath := filepath.Join(topicDir, topic.TeachingNotesFile)
		if data, err := os.ReadFile(notesPath); err == nil {
			ctx.ExistingNotes = string(data)
		}
	}

	return ctx, nil
}

// FindTopicFile searches the repo for a topic file with the given ID.
func FindTopicFile(repoDir, topicID string) (string, error) {
	var found string

	err := filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		if strings.HasSuffix(path, ".assessments.yaml") || strings.HasSuffix(path, ".examples.yaml") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		var partial struct {
			ID string `yaml:"id"`
		}
		if err := yaml.Unmarshal(data, &partial); err != nil {
			return nil
		}

		if partial.ID == topicID {
			found = path
			return filepath.SkipAll
		}
		return nil
	})

	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("topic %s not found in %s", topicID, repoDir)
	}
	return found, nil
}

// loadTopic parses a topic YAML file.
func loadTopic(path string) (*Topic, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var topic Topic
	if err := yaml.Unmarshal(data, &topic); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	return &topic, nil
}

// loadSiblingTopics loads all topics in the same directory, excluding the given ID.
func loadSiblingTopics(dir, excludeID string) ([]Topic, error) {
	var siblings []Topic

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".assessments.yaml") ||
			strings.HasSuffix(entry.Name(), ".examples.yaml") {
			continue
		}

		topic, err := loadTopic(filepath.Join(dir, entry.Name()))
		if err != nil || topic.ID == excludeID {
			continue
		}
		siblings = append(siblings, *topic)
	}

	return siblings, nil
}
