package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// TopicInfo holds the presence of fields for quality assessment.
type TopicInfo struct {
	ID   string
	Name string

	// Level 0: basic identity
	HasID                 bool
	HasName               bool
	HasLearningObjectives bool

	// Level 1: structured metadata
	HasPrerequisites bool
	HasDifficulty    bool
	HasBloomLevels   bool

	// Level 2: teaching content
	HasTeachingSequence bool
	HasMisconceptions   bool
	HasEngagementHooks  bool

	// Level 3: full content files
	HasTeachingNotes bool
	HasExamples      bool
	HasAssessments   bool

	// Level 4: translations + cross-curriculum
	HasTranslation     bool
	HasCrossCurriculum bool

	// Level 5: authority validated
	HasAuthorityValidation bool

	// Claimed level (from YAML)
	ClaimedLevel int
}

// QualityReport holds the quality assessment results.
type QualityReport struct {
	Topics      []TopicQuality
	LevelCounts map[int]int
}

// TopicQuality holds quality info for a single topic.
type TopicQuality struct {
	ID           string
	Name         string
	ActualLevel  int
	ClaimedLevel int
	Overclaimed  bool
}

// AssessQuality determines the actual quality level of a topic based on present fields.
func AssessQuality(info TopicInfo) int {
	// Level 0: has id, name, learning_objectives
	if !info.HasID || !info.HasName || !info.HasLearningObjectives {
		return 0
	}

	// Level 1: + prerequisites, difficulty, bloom_levels
	if !info.HasPrerequisites || !info.HasDifficulty || !info.HasBloomLevels {
		return 0
	}

	// Level 2: + teaching.sequence, teaching.common_misconceptions, engagement_hooks
	if !info.HasTeachingSequence || !info.HasMisconceptions || !info.HasEngagementHooks {
		return 1
	}

	// Level 3: + teaching_notes file, examples file, assessments file
	if !info.HasTeachingNotes || !info.HasExamples || !info.HasAssessments {
		return 2
	}

	// Level 4: + translation, cross_curriculum
	if !info.HasTranslation || !info.HasCrossCurriculum {
		return 3
	}

	// Level 5: authority validated
	if !info.HasAuthorityValidation {
		return 4
	}

	return 5
}

// TopicInfoFromYAML parses a YAML topic file and inspects the directory
// for companion files to build a TopicInfo for quality assessment.
func TopicInfoFromYAML(data []byte, filePath, baseDir string) TopicInfo {
	var raw map[string]interface{}
	_ = yaml.Unmarshal(data, &raw)

	info := TopicInfo{}

	if id, ok := raw["id"].(string); ok && id != "" {
		info.HasID = true
		info.ID = id
	}
	if name, ok := raw["name"].(string); ok && name != "" {
		info.HasName = true
		info.Name = name
	}
	if los, ok := raw["learning_objectives"].([]interface{}); ok && len(los) > 0 {
		info.HasLearningObjectives = true
		// Check if bloom levels are present in learning objectives
		for _, lo := range los {
			if m, ok := lo.(map[string]interface{}); ok {
				if _, ok := m["bloom"]; ok {
					info.HasBloomLevels = true
					break
				}
			}
		}
	}
	if _, ok := raw["prerequisites"]; ok {
		info.HasPrerequisites = true
	}
	if _, ok := raw["difficulty"]; ok {
		info.HasDifficulty = true
	}

	// Check teaching sub-fields
	if teaching, ok := raw["teaching"].(map[string]interface{}); ok {
		if seq, ok := teaching["sequence"].([]interface{}); ok && len(seq) > 0 {
			info.HasTeachingSequence = true
		}
		if misc, ok := teaching["common_misconceptions"].([]interface{}); ok && len(misc) > 0 {
			info.HasMisconceptions = true
		}
	}
	if hooks, ok := raw["engagement_hooks"].([]interface{}); ok && len(hooks) > 0 {
		info.HasEngagementHooks = true
	}

	// Check companion files
	dir := filepath.Dir(filePath)
	base := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))

	if _, err := os.Stat(filepath.Join(dir, base+".teaching.md")); err == nil {
		info.HasTeachingNotes = true
	}
	if _, err := os.Stat(filepath.Join(dir, base+".examples.yaml")); err == nil {
		info.HasExamples = true
	}
	if _, err := os.Stat(filepath.Join(dir, base+".assessments.yaml")); err == nil {
		info.HasAssessments = true
	}

	// Check translations (translations/ directory with at least one lang subfolder containing a file for this topic)
	translationsDir := filepath.Join(dir, "translations")
	if entries, err := os.ReadDir(translationsDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				// Check if this lang folder has a translation file for this topic
				langDir := filepath.Join(translationsDir, entry.Name())
				candidates := []string{
					filepath.Join(langDir, base+".yaml"),
					filepath.Join(langDir, base+".teaching.md"),
					filepath.Join(langDir, base+".assessments.yaml"),
				}
				for _, candidate := range candidates {
					if _, err := os.Stat(candidate); err == nil {
						info.HasTranslation = true
						break
					}
				}
				if info.HasTranslation {
					break
				}
			}
		}
	}
	if _, ok := raw["cross_curriculum"]; ok {
		info.HasCrossCurriculum = true
	}
	if _, ok := raw["authority_validation"]; ok {
		info.HasAuthorityValidation = true
	}

	if ql, ok := raw["quality_level"].(int); ok {
		info.ClaimedLevel = ql
	}

	return info
}

// FormatQualityReport generates a human-readable quality report.
func FormatQualityReport(report QualityReport) string {
	result := "=== Quality Level Report ===\n"
	levelNames := map[int]string{
		0: "Stub", 1: "Basic", 2: "Structured",
		3: "Teachable", 4: "Complete", 5: "Gold",
	}

	for level := 5; level >= 0; level-- {
		count := report.LevelCounts[level]
		result += fmt.Sprintf("Level %d (%s): %d topics\n", level, levelNames[level], count)
	}

	// Flag overclaimed
	var overclaimed []TopicQuality
	for _, t := range report.Topics {
		if t.Overclaimed {
			overclaimed = append(overclaimed, t)
		}
	}
	if len(overclaimed) > 0 {
		result += "\n⚠️  Overclaimed quality levels:\n"
		for _, t := range overclaimed {
			result += fmt.Sprintf("  %s: claims Level %d, actual Level %d\n", t.ID, t.ClaimedLevel, t.ActualLevel)
		}
	}

	return result
}
