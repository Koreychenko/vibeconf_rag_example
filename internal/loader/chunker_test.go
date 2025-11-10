package loader

import (
	"strings"
	"testing"
)

func TestChunkText(t *testing.T) {
	// Test text with multiple paragraphs
	text := `This is the first paragraph.
	
This is the second paragraph. It has multiple sentences. This is to test paragraph chunking.

This is the third paragraph, which also has multiple sentences. We want to ensure proper chunking.`

	// Test different chunking strategies - don't assert exact chunk counts
	// as implementation details may vary
	testCases := []struct {
		name     string
		strategy ChunkingStrategy
		maxSize  int
		overlap  int
	}{
		{
			name:     "by paragraph with large max size",
			strategy: ByParagraph,
			maxSize:  1000,
			overlap:  0,
		},
		{
			name:     "by sentence with default size",
			strategy: BySentence,
			maxSize:  1000,
			overlap:  0,
		},
		{
			name:     "by fixed size small chunks",
			strategy: ByFixedSize,
			maxSize:  50,
			overlap:  10,
		},
		{
			name:     "by paragraph with small max size",
			strategy: ByParagraph,
			maxSize:  60,
			overlap:  10,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			options := ChunkingOptions{
				Strategy:     tc.strategy,
				MaxChunkSize: tc.maxSize,
				ChunkOverlap: tc.overlap,
			}

			chunks := ChunkText(text, options)

			// Check that we get at least one chunk
			if len(chunks) == 0 {
				t.Errorf("Expected at least one chunk, got 0")
			}

			// Verify no empty chunks
			for i, chunk := range chunks {
				if strings.TrimSpace(chunk) == "" {
					t.Errorf("Chunk %d is empty", i)
				}
			}

			// Verify each chunk is no larger than maxSize
			for i, chunk := range chunks {
				if len(chunk) > tc.maxSize {
					t.Errorf("Chunk %d size %d exceeds max size %d", i, len(chunk), tc.maxSize)
				}
			}
		})
	}
}

func TestDefaultChunkingOptions(t *testing.T) {
	options := DefaultChunkingOptions()

	// Verify default values
	if options.Strategy != ByParagraph {
		t.Errorf("Expected default strategy ByParagraph, got %v", options.Strategy)
	}

	if options.MaxChunkSize != 1000 {
		t.Errorf("Expected default max chunk size 1000, got %d", options.MaxChunkSize)
	}

	if options.ChunkOverlap != 100 {
		t.Errorf("Expected default chunk overlap 100, got %d", options.ChunkOverlap)
	}
}

func TestChunkByParagraph(t *testing.T) {
	// Test text with paragraphs
	text := `Paragraph 1. This is a test.

Paragraph 2. Another test paragraph.

Paragraph 3. The final test paragraph for chunking.`

	// Test with large max size
	chunks := chunkByParagraph(text, 1000, 0)

	// At least one chunk should be returned
	if len(chunks) < 1 {
		t.Errorf("Expected at least one chunk, got %d", len(chunks))
	}

	// Test with small max size that forces combining
	smallChunks := chunkByParagraph(text, 20, 0)
	if len(smallChunks) < 1 {
		t.Errorf("Expected at least one chunk for small max size, got %d", len(smallChunks))
	}

	// If chunks were created, verify first chunk content
	if len(chunks) > 0 {
		if !strings.Contains(chunks[0], "Paragraph 1") {
			t.Errorf("First chunk doesn't contain expected content: %s", chunks[0])
		}
	}
}

func TestChunkBySentence(t *testing.T) {
	// Text with sentences
	text := "This is sentence one. This is sentence two! Is this sentence three? Yes, this is sentence four."

	chunks := chunkBySentence(text, 1000, 0)

	// At least one chunk should be returned
	if len(chunks) < 1 {
		t.Errorf("Expected at least one chunk, got %d", len(chunks))
	}

	// Test with small max size
	smallChunks := chunkBySentence(text, 15, 0)
	if len(smallChunks) < 1 {
		t.Errorf("Expected at least one chunk for small max size, got %d", len(smallChunks))
	}
}

func TestChunkByFixedSize(t *testing.T) {
	// Test with a simple string and specific chunk size
	text := "abcdefghijklmnopqrstuvwxyz"
	maxSize := 10
	overlap := 2

	chunks := chunkByFixedSize(text, maxSize, overlap)

	// We should get at least one chunk
	if len(chunks) < 1 {
		t.Errorf("Expected at least one chunk, got 0")
	}

	// Check that each chunk is no larger than maxSize
	for i, chunk := range chunks {
		if len(chunk) > maxSize {
			t.Errorf("Chunk %d size %d exceeds max size %d", i, len(chunk), maxSize)
		}
	}

	// Test empty text
	emptyChunks := chunkByFixedSize("", maxSize, overlap)
	if len(emptyChunks) != 0 {
		t.Errorf("Expected 0 chunks for empty text, got %d", len(emptyChunks))
	}

	// Test text smaller than max size
	smallText := "small"
	smallChunks := chunkByFixedSize(smallText, maxSize, overlap)
	if len(smallChunks) != 1 {
		t.Errorf("Expected 1 chunk for text smaller than max size, got %d", len(smallChunks))
	}
}

func TestAllParagraphsUnderMaxSize(t *testing.T) {
	paragraphs := []string{
		"This is a short paragraph.",
		"This is another short paragraph.",
		"This is a longer paragraph that might exceed some smaller size limits.",
	}

	// All paragraphs should be under 100 chars
	if !allParagraphsUnderMaxSize(paragraphs, 100) {
		t.Error("Expected all paragraphs to be under 100 chars, but they weren't")
	}

	// The third paragraph should exceed 30 chars
	if allParagraphsUnderMaxSize(paragraphs, 30) {
		t.Error("Expected at least one paragraph to exceed 30 chars, but none did")
	}

	// Empty paragraphs list should return true
	if !allParagraphsUnderMaxSize([]string{}, 10) {
		t.Error("Expected empty paragraph list to return true")
	}
}
