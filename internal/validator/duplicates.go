package validator

import (
	"fmt"
	"math"
	"strings"
	"unicode"
)

// DuplicatePair represents two similar content items.
type DuplicatePair struct {
	IndexA     int
	IndexB     int
	Similarity float64
}

// Tokenize splits text into lowercase tokens, removing punctuation.
func Tokenize(text string) []string {
	text = strings.ToLower(text)
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	return words
}

// CosineSimilarity computes cosine similarity between two texts using bag-of-words.
func CosineSimilarity(a, b string) float64 {
	tokensA := Tokenize(a)
	tokensB := Tokenize(b)

	if len(tokensA) == 0 || len(tokensB) == 0 {
		return 0
	}

	// Build term frequency vectors
	freqA := termFrequency(tokensA)
	freqB := termFrequency(tokensB)

	// Compute dot product and magnitudes
	var dotProduct, magA, magB float64

	allTerms := make(map[string]bool)
	for t := range freqA {
		allTerms[t] = true
	}
	for t := range freqB {
		allTerms[t] = true
	}

	for term := range allTerms {
		a := freqA[term]
		b := freqB[term]
		dotProduct += a * b
		magA += a * a
		magB += b * b
	}

	if magA == 0 || magB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(magA) * math.Sqrt(magB))
}

// termFrequency builds a term frequency map from tokens.
func termFrequency(tokens []string) map[string]float64 {
	freq := make(map[string]float64)
	for _, t := range tokens {
		freq[t]++
	}
	return freq
}

// FindDuplicates finds pairs of texts that exceed the similarity threshold.
func FindDuplicates(texts []string, threshold float64) []DuplicatePair {
	var pairs []DuplicatePair

	for i := 0; i < len(texts); i++ {
		for j := i + 1; j < len(texts); j++ {
			sim := CosineSimilarity(texts[i], texts[j])
			if sim >= threshold {
				pairs = append(pairs, DuplicatePair{
					IndexA:     i,
					IndexB:     j,
					Similarity: sim,
				})
			}
		}
	}

	return pairs
}

// FormatDuplicateReport creates a human-readable report of duplicate pairs.
func FormatDuplicateReport(pairs []DuplicatePair, texts []string) []string {
	var report []string
	for _, p := range pairs {
		report = append(report, fmt.Sprintf(
			"%.0f%% similar: [%d] %q ↔ [%d] %q",
			p.Similarity*100,
			p.IndexA, truncate(texts[p.IndexA], 50),
			p.IndexB, truncate(texts[p.IndexB], 50),
		))
	}
	return report
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
