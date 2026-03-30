package validator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/validator"
)

func TestAssessQuality(t *testing.T) {
	tests := []struct {
		name     string
		topic    validator.TopicInfo
		expected int
	}{
		{
			name: "level-0-minimal",
			topic: validator.TopicInfo{
				HasID:                 true,
				HasName:               true,
				HasLearningObjectives: true,
			},
			expected: 0,
		},
		{
			name: "level-1-basic",
			topic: validator.TopicInfo{
				HasID:                 true,
				HasName:               true,
				HasLearningObjectives: true,
				HasPrerequisites:      true,
				HasDifficulty:         true,
				HasBloomLevels:        true,
			},
			expected: 1,
		},
		{
			name: "level-2-structured",
			topic: validator.TopicInfo{
				HasID: true, HasName: true, HasLearningObjectives: true,
				HasPrerequisites: true, HasDifficulty: true, HasBloomLevels: true,
				HasTeachingSequence: true,
				HasMisconceptions:   true,
				HasEngagementHooks:  true,
			},
			expected: 2,
		},
		{
			name: "level-3-teachable",
			topic: validator.TopicInfo{
				HasID: true, HasName: true, HasLearningObjectives: true,
				HasPrerequisites: true, HasDifficulty: true, HasBloomLevels: true,
				HasTeachingSequence: true, HasMisconceptions: true, HasEngagementHooks: true,
				HasTeachingNotes: true,
				HasExamples:      true,
				HasAssessments:   true,
			},
			expected: 3,
		},
		{
			name: "level-4-complete",
			topic: validator.TopicInfo{
				HasID: true, HasName: true, HasLearningObjectives: true,
				HasPrerequisites: true, HasDifficulty: true, HasBloomLevels: true,
				HasTeachingSequence: true, HasMisconceptions: true, HasEngagementHooks: true,
				HasTeachingNotes: true, HasExamples: true, HasAssessments: true,
				HasTranslation:     true,
				HasCrossCurriculum: true,
			},
			expected: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.AssessQuality(tt.topic)
			if got != tt.expected {
				t.Errorf("AssessQuality() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestTopicInfoFromYAML_DetectsTranslationDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create topic file
	topicYAML := `id: MT3-09
name: Garis Lurus
name_en: Straight Lines
learning_objectives:
  - id: LO1
    text: "Test"
    bloom: apply
prerequisites: [MT3-08]
difficulty: intermediate
cross_curriculum:
  - subject: science
    topic: measurement
`
	topicFile := filepath.Join(tmpDir, "MT3-09.yaml")
	if err := os.WriteFile(topicFile, []byte(topicYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create companion files
	os.WriteFile(filepath.Join(tmpDir, "MT3-09.teaching.md"), []byte("# Notes"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "MT3-09.examples.yaml"), []byte("examples: []"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "MT3-09.assessments.yaml"), []byte("assessments: []"), 0o644)

	// Without translations dir — should NOT detect translation
	info := validator.TopicInfoFromYAML([]byte(topicYAML), topicFile, tmpDir)
	if info.HasTranslation {
		t.Error("should not detect translation when translations/ dir does not exist")
	}

	// Create translations/en/MT3-09.yaml
	transDir := filepath.Join(tmpDir, "translations", "en")
	if err := os.MkdirAll(transDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(transDir, "MT3-09.yaml"), []byte("name: Straight Lines\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// With translations dir — should detect translation
	info = validator.TopicInfoFromYAML([]byte(topicYAML), topicFile, tmpDir)
	if !info.HasTranslation {
		t.Error("should detect translation when translations/en/MT3-09.yaml exists")
	}
}
