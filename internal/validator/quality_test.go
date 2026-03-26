package validator_test

import (
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
