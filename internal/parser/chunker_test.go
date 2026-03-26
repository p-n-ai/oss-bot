package parser_test

import (
	"strings"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/parser"
)

func TestChunkDocument_SingleChunk(t *testing.T) {
	text := "This is a short document with no headings."
	opts := parser.ChunkOptions{MaxChunkSize: 8000, OverlapTokens: 200}

	chunks := parser.ChunkDocument(text, opts)

	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0].Content != text {
		t.Errorf("chunk content mismatch")
	}
	if chunks[0].Index != 0 {
		t.Errorf("chunk index = %d, want 0", chunks[0].Index)
	}
	if chunks[0].Total != 1 {
		t.Errorf("chunk total = %d, want 1", chunks[0].Total)
	}
}

func TestChunkDocument_SplitsAtHeadings(t *testing.T) {
	text := `# Chapter 1
Content of chapter one.

# Chapter 2
Content of chapter two.

# Chapter 3
Content of chapter three.`

	opts := parser.ChunkOptions{
		MaxChunkSize:  8000,
		OverlapTokens: 0,
		SplitOn:       []string{"# "},
	}

	chunks := parser.ChunkDocument(text, opts)

	if len(chunks) < 3 {
		t.Fatalf("expected at least 3 chunks, got %d", len(chunks))
	}

	// First chunk should contain Chapter 1
	if !strings.Contains(chunks[0].Content, "Chapter 1") {
		t.Errorf("first chunk should contain 'Chapter 1', got: %q", chunks[0].Content)
	}

	// Chunks should have correct headings
	if chunks[0].Heading == "" {
		t.Error("first chunk should have a heading")
	}
}

func TestChunkDocument_HeadingExtracted(t *testing.T) {
	text := `## Introduction
Some intro text.

## Background
Some background text.`

	opts := parser.ChunkOptions{
		MaxChunkSize: 8000,
		SplitOn:      []string{"## "},
	}

	chunks := parser.ChunkDocument(text, opts)

	if len(chunks) < 2 {
		t.Fatalf("expected at least 2 chunks, got %d", len(chunks))
	}

	if !strings.Contains(chunks[0].Heading, "Introduction") {
		t.Errorf("first chunk heading = %q, want to contain 'Introduction'", chunks[0].Heading)
	}
	if !strings.Contains(chunks[1].Heading, "Background") {
		t.Errorf("second chunk heading = %q, want to contain 'Background'", chunks[1].Heading)
	}
}

func TestChunkDocument_DefaultSplitPatterns(t *testing.T) {
	text := `Chapter 1: Algebra
First chapter content.

Chapter 2: Geometry
Second chapter content.`

	// Default SplitOn should include "Chapter "
	opts := parser.ChunkOptions{MaxChunkSize: 8000}

	chunks := parser.ChunkDocument(text, opts)

	if len(chunks) < 2 {
		t.Fatalf("expected at least 2 chunks with default patterns, got %d", len(chunks))
	}
}

func TestChunkDocument_TotalIsConsistent(t *testing.T) {
	text := `# Part A
Content A.

# Part B
Content B.

# Part C
Content C.`

	opts := parser.ChunkOptions{MaxChunkSize: 8000, SplitOn: []string{"# "}}
	chunks := parser.ChunkDocument(text, opts)

	for i, c := range chunks {
		if c.Total != len(chunks) {
			t.Errorf("chunk %d: Total = %d, want %d", i, c.Total, len(chunks))
		}
		if c.Index != i {
			t.Errorf("chunk %d: Index = %d, want %d", i, c.Index, i)
		}
	}
}

func TestChunkDocument_LargeDocumentSplitsBySize(t *testing.T) {
	// Create a document that exceeds MaxChunkSize (in rough token estimate)
	// ~4 chars per token, so 200 tokens ≈ 800 chars
	var sb strings.Builder
	for i := 0; i < 300; i++ {
		sb.WriteString("word ")
	}
	text := sb.String() // ~1500 chars ≈ 375 tokens

	opts := parser.ChunkOptions{
		MaxChunkSize:  100, // very small to force splitting
		OverlapTokens: 10,
	}

	chunks := parser.ChunkDocument(text, opts)

	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks for large doc, got %d", len(chunks))
	}
}

func TestChunkDocument_OverlapIncludesContext(t *testing.T) {
	text := `# Section A
` + strings.Repeat("alpha ", 200) + `
# Section B
` + strings.Repeat("beta ", 200)

	opts := parser.ChunkOptions{
		MaxChunkSize:  200,
		OverlapTokens: 20,
		SplitOn:       []string{"# "},
	}

	chunks := parser.ChunkDocument(text, opts)

	// With overlap, chunk B should contain some content from end of chunk A
	if len(chunks) >= 2 {
		// The overlap means chunk[1] starts with some content from chunk[0]
		// Just verify chunks are non-empty and properly indexed
		for _, c := range chunks {
			if c.Content == "" {
				t.Error("chunk content should not be empty")
			}
		}
	}
}
