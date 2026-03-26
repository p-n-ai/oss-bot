package parser

import (
	"strings"
)

// ChunkOptions controls how documents are split into chunks.
type ChunkOptions struct {
	// MaxChunkSize is the maximum tokens per chunk (default: 8000).
	// Token count is estimated as len(text)/4.
	MaxChunkSize int
	// OverlapTokens is the number of tokens to overlap between consecutive chunks
	// for context continuity (default: 200).
	OverlapTokens int
	// SplitOn is a list of heading patterns to split on (e.g., "# ", "## ", "Chapter ").
	// If empty, defaults to ["# ", "## ", "### ", "Chapter "].
	SplitOn []string
}

// Chunk represents a portion of a large document.
type Chunk struct {
	Index   int    // 0-based chunk index
	Total   int    // Total number of chunks (set after all chunks are created)
	Content string // The chunk text
	Heading string // The heading that starts this chunk (if any)
}

// defaultSplitPatterns are the heading patterns used when SplitOn is not specified.
var defaultSplitPatterns = []string{"# ", "## ", "### ", "Chapter "}

// ChunkDocument splits a large document into semantically meaningful chunks.
// It first tries to split at heading boundaries; if any resulting chunk still exceeds
// MaxChunkSize, it further splits that chunk by token count.
func ChunkDocument(text string, opts ChunkOptions) []Chunk {
	if opts.MaxChunkSize == 0 {
		opts.MaxChunkSize = 8000
	}
	if opts.OverlapTokens == 0 {
		opts.OverlapTokens = 200
	}
	splitOn := opts.SplitOn
	if len(splitOn) == 0 {
		splitOn = defaultSplitPatterns
	}

	// Split into lines and group by heading boundaries.
	sections := splitByHeadings(text, splitOn)

	// For each section, further split if it exceeds MaxChunkSize.
	var rawChunks []Chunk
	for _, sec := range sections {
		subChunks := splitBySize(sec.Content, sec.Heading, opts)
		rawChunks = append(rawChunks, subChunks...)
	}

	// Apply overlap between consecutive chunks.
	chunks := applyOverlap(rawChunks, opts.OverlapTokens)

	// Set Index and Total on all chunks.
	total := len(chunks)
	for i := range chunks {
		chunks[i].Index = i
		chunks[i].Total = total
	}

	return chunks
}

// section is an intermediate representation before chunking.
type section struct {
	Heading string
	Content string
}

// splitByHeadings groups lines into sections based on heading patterns.
func splitByHeadings(text string, patterns []string) []section {
	lines := strings.Split(text, "\n")
	var sections []section
	var current section
	hasContent := false

	for _, line := range lines {
		if isHeading(line, patterns) {
			if hasContent || current.Heading != "" {
				sections = append(sections, current)
			}
			current = section{
				Heading: extractHeadingText(line),
				Content: line + "\n",
			}
			hasContent = true
		} else {
			current.Content += line + "\n"
			if strings.TrimSpace(line) != "" {
				hasContent = true
			}
		}
	}

	if hasContent || current.Heading != "" {
		sections = append(sections, current)
	}

	if len(sections) == 0 {
		sections = []section{{Content: text}}
	}

	// Trim trailing newlines from each section's content.
	for i := range sections {
		sections[i].Content = strings.TrimRight(sections[i].Content, "\n")
	}

	return sections
}

// isHeading reports whether a line starts with one of the given heading patterns.
func isHeading(line string, patterns []string) bool {
	trimmed := strings.TrimLeft(line, " \t")
	for _, p := range patterns {
		if strings.HasPrefix(trimmed, p) {
			return true
		}
	}
	return false
}

// extractHeadingText returns the heading text stripped of markdown prefix.
func extractHeadingText(line string) string {
	line = strings.TrimLeft(line, " \t")
	// Strip leading '#' characters and spaces
	stripped := strings.TrimLeft(line, "#")
	return strings.TrimSpace(stripped)
}

// estimateTokens returns a rough token count (1 token ≈ 4 characters).
func estimateTokens(text string) int {
	return len(text) / 4
}

// splitBySize further splits a section's content if it exceeds MaxChunkSize.
func splitBySize(content, heading string, opts ChunkOptions) []Chunk {
	if estimateTokens(content) <= opts.MaxChunkSize {
		return []Chunk{{Heading: heading, Content: content}}
	}

	// Split by words, accumulating up to MaxChunkSize tokens per chunk.
	maxChars := opts.MaxChunkSize * 4
	var chunks []Chunk
	isFirst := true

	for len(content) > 0 {
		if len(content) <= maxChars {
			h := ""
			if isFirst {
				h = heading
			}
			chunks = append(chunks, Chunk{Heading: h, Content: content})
			break
		}

		// Find a good split point near maxChars (prefer newline boundary).
		splitAt := maxChars
		if splitAt > len(content) {
			splitAt = len(content)
		}

		// Walk back to find a newline for a cleaner split.
		for splitAt > maxChars/2 && content[splitAt-1] != '\n' {
			splitAt--
		}
		if splitAt == 0 {
			splitAt = maxChars
		}

		h := ""
		if isFirst {
			h = heading
			isFirst = false
		}
		chunks = append(chunks, Chunk{Heading: h, Content: content[:splitAt]})
		content = content[splitAt:]
	}

	return chunks
}

// applyOverlap adds context overlap between consecutive chunks.
// Each chunk (except the first) prepends the last OverlapTokens tokens from the previous chunk.
func applyOverlap(chunks []Chunk, overlapTokens int) []Chunk {
	if overlapTokens == 0 || len(chunks) <= 1 {
		return chunks
	}

	overlapChars := overlapTokens * 4
	result := make([]Chunk, len(chunks))
	result[0] = chunks[0]

	for i := 1; i < len(chunks); i++ {
		prev := chunks[i-1].Content
		overlap := ""
		if len(prev) > overlapChars {
			overlap = prev[len(prev)-overlapChars:]
		} else {
			overlap = prev
		}
		result[i] = Chunk{
			Heading: chunks[i].Heading,
			Content: overlap + chunks[i].Content,
		}
	}

	return result
}
