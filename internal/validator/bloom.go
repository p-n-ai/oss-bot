package validator

import "strings"

// LearningObjective represents a topic's learning objective for bloom validation.
type LearningObjective struct {
	ID    string
	Text  string
	Bloom string
}

// AssessmentQuestion represents an assessment question for bloom validation.
type AssessmentQuestion struct {
	ID                string
	Text              string
	Difficulty        string
	LearningObjective string
}

// bloomVerbs maps verbs to their Bloom's taxonomy level.
var bloomVerbs = map[string]string{
	// Remember
	"list": "remember", "define": "remember", "recall": "remember",
	"identify": "remember", "name": "remember", "state": "remember",
	"label": "remember", "recognise": "remember", "recognize": "remember",
	// Understand
	"explain": "understand", "describe": "understand", "summarise": "understand",
	"summarize": "understand", "interpret": "understand", "classify": "understand",
	"discuss": "understand", "distinguish": "understand", "paraphrase": "understand",
	// Apply
	"solve": "apply", "calculate": "apply", "use": "apply",
	"apply": "apply", "demonstrate": "apply", "compute": "apply",
	"determine": "apply", "construct": "apply", "show": "apply",
	// Analyze
	"compare": "analyze", "contrast": "analyze", "differentiate": "analyze",
	"analyse": "analyze", "analyze": "analyze", "examine": "analyze",
	"investigate": "analyze", "categorise": "analyze", "categorize": "analyze",
	// Evaluate
	"justify": "evaluate", "evaluate": "evaluate", "assess": "evaluate",
	"critique": "evaluate", "judge": "evaluate", "argue": "evaluate",
	"defend": "evaluate", "recommend": "evaluate",
	// Create
	"design": "create", "create": "create", "formulate": "create",
	"compose": "create", "develop": "create", "invent": "create",
	"plan": "create", "produce": "create", "propose": "create",
}

// bloomOrder defines the hierarchy of Bloom's levels (index = rank).
var bloomOrder = []string{"remember", "understand", "apply", "analyze", "evaluate", "create"}

// BloomLevelForVerb returns the Bloom's taxonomy level for a given verb.
// Returns empty string if the verb is not recognized.
func BloomLevelForVerb(verb string) string {
	return bloomVerbs[strings.ToLower(verb)]
}

// bloomRank returns the numeric rank of a Bloom's level (0=remember, 5=create).
// Returns -1 if the level is not recognized.
func bloomRank(level string) int {
	for i, l := range bloomOrder {
		if l == level {
			return i
		}
	}
	return -1
}

// ValidateBloomConsistency checks that assessment questions don't exceed
// the Bloom's level of their referenced learning objective.
func ValidateBloomConsistency(objectives []LearningObjective, questions []AssessmentQuestion) []string {
	objMap := make(map[string]string) // LO ID -> bloom level
	for _, o := range objectives {
		objMap[o.ID] = o.Bloom
	}

	var errors []string
	for _, q := range questions {
		loBloom, ok := objMap[q.LearningObjective]
		if !ok {
			errors = append(errors, "question "+q.ID+" references unknown learning objective "+q.LearningObjective)
			continue
		}

		// Extract first verb from question text
		questionBloom := detectBloomFromText(q.Text)
		if questionBloom == "" {
			continue // Can't detect, skip
		}

		loRank := bloomRank(loBloom)
		qRank := bloomRank(questionBloom)

		if qRank > loRank {
			errors = append(errors,
				"question "+q.ID+" uses "+questionBloom+"-level verb but learning objective "+
					q.LearningObjective+" is at "+loBloom+" level")
		}
	}

	return errors
}

// detectBloomFromText extracts the highest Bloom's level verb from text.
func detectBloomFromText(text string) string {
	words := strings.Fields(strings.ToLower(text))
	highestRank := -1
	highestLevel := ""

	for _, word := range words {
		// Strip punctuation
		word = strings.Trim(word, ".,;:!?()\"'")
		if level, ok := bloomVerbs[word]; ok {
			rank := bloomRank(level)
			if rank > highestRank {
				highestRank = rank
				highestLevel = level
			}
		}
	}

	return highestLevel
}
