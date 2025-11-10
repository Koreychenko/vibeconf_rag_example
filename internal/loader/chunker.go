package loader

import (
	"strings"
	"unicode"
)

// ChunkingStrategy defines different approaches for chunking text
type ChunkingStrategy string

const (
	// ByParagraph chunks text by paragraphs
	ByParagraph ChunkingStrategy = "paragraph"
	// BySentence chunks text by sentences
	BySentence ChunkingStrategy = "sentence"
	// ByFixedSize chunks text by a fixed number of characters
	ByFixedSize ChunkingStrategy = "fixed_size"
)

// ChunkingOptions defines options for text chunking
type ChunkingOptions struct {
	// Strategy determines how to chunk the text
	Strategy ChunkingStrategy
	// MaxChunkSize is the maximum size of a chunk in characters/tokens
	MaxChunkSize int
	// ChunkOverlap is the number of characters/tokens that overlap between chunks
	ChunkOverlap int
}

// DefaultChunkingOptions returns the default chunking options
func DefaultChunkingOptions() ChunkingOptions {
	return ChunkingOptions{
		Strategy:     ByParagraph,
		MaxChunkSize: 1000,
		ChunkOverlap: 100,
	}
}

// ChunkText splits text into chunks based on the specified strategy
func ChunkText(text string, options ChunkingOptions) []string {
	switch options.Strategy {
	case ByParagraph:
		return chunkByParagraph(text, options.MaxChunkSize, options.ChunkOverlap)
	case BySentence:
		return chunkBySentence(text, options.MaxChunkSize, options.ChunkOverlap)
	case ByFixedSize:
		return chunkByFixedSize(text, options.MaxChunkSize, options.ChunkOverlap)
	default:
		return chunkByParagraph(text, options.MaxChunkSize, options.ChunkOverlap)
	}
}

// chunkByParagraph splits text into chunks by paragraphs
func chunkByParagraph(text string, maxSize, overlap int) []string {
	// Split text by double newlines which typically indicate paragraphs
	paragraphs := strings.Split(text, "\n\n")

	// Clean paragraphs (remove empty ones, trim whitespace)
	var cleanParagraphs []string
	for _, p := range paragraphs {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			cleanParagraphs = append(cleanParagraphs, trimmed)
		}
	}

	// Handle case with no paragraphs
	if len(cleanParagraphs) == 0 {
		return []string{}
	}

	// If each paragraph is smaller than maxSize, return individual paragraphs
	if allParagraphsUnderMaxSize(cleanParagraphs, maxSize) {
		return cleanParagraphs
	}

	// Otherwise, use fixed size chunking with paragraph boundaries preserved where possible
	return chunkPreservingParagraphs(cleanParagraphs, maxSize, overlap)
}

// chunkBySentence splits text into chunks by sentences
func chunkBySentence(text string, maxSize, overlap int) []string {
	// Simple sentence splitting by common punctuation followed by space
	sentences := splitIntoSentences(text)

	// Clean sentences
	var cleanSentences []string
	for _, s := range sentences {
		trimmed := strings.TrimSpace(s)
		if trimmed != "" {
			cleanSentences = append(cleanSentences, trimmed)
		}
	}

	// Handle case with no sentences
	if len(cleanSentences) == 0 {
		return []string{}
	}

	return combineItemsIntoChunks(cleanSentences, maxSize, overlap)
}

// chunkByFixedSize splits text into chunks of fixed size
func chunkByFixedSize(text string, maxSize, overlap int) []string {
	var chunks []string
	text = strings.TrimSpace(text)

	if text == "" {
		return chunks
	}

	if len(text) <= maxSize {
		return []string{text}
	}

	// Sliding window with overlap
	for i := 0; i < len(text); i += maxSize - overlap {
		end := i + maxSize
		if end > len(text) {
			end = len(text)
		}

		// Don't create tiny chunks at the end
		if end-i < maxSize/4 && len(chunks) > 0 {
			// Extend the previous chunk
			lastIdx := len(chunks) - 1
			chunks[lastIdx] = chunks[lastIdx] + " " + text[i:end]
			break
		}

		chunk := text[i:end]
		chunks = append(chunks, chunk)

		if end == len(text) {
			break
		}
	}

	return chunks
}

// chunkPreservingParagraphs combines paragraphs into chunks while preserving paragraph boundaries
func chunkPreservingParagraphs(paragraphs []string, maxSize, overlap int) []string {
	return combineItemsIntoChunks(paragraphs, maxSize, overlap)
}

// combineItemsIntoChunks combines text items (paragraphs, sentences) into chunks
func combineItemsIntoChunks(items []string, maxSize, overlap int) []string {
	var chunks []string
	var currentChunk string

	for _, item := range items {
		// If adding this item exceeds maxSize and we already have content
		if len(currentChunk) > 0 && len(currentChunk)+1+len(item) > maxSize {
			// Store current chunk
			chunks = append(chunks, strings.TrimSpace(currentChunk))

			// Start new chunk with overlap if possible
			if overlap > 0 && len(currentChunk) > overlap {
				// Try to find a clean break point for overlap
				words := strings.Fields(currentChunk)
				if len(words) > 3 { // Need at least a few words for sensible overlap
					// Take approximately the last 'overlap' characters worth of words
					overlapText := getOverlapText(words, overlap)
					currentChunk = overlapText + " " + item
				} else {
					currentChunk = item
				}
			} else {
				currentChunk = item
			}
		} else {
			// Add separator if not the first item in chunk
			if len(currentChunk) > 0 {
				currentChunk += " "
			}
			currentChunk += item
		}

		// Handle case where a single item is larger than maxSize
		if len(currentChunk) > maxSize {
			// Chunk it by fixed size
			fixedChunks := chunkByFixedSize(currentChunk, maxSize, overlap)

			// Add all but the last fixed chunk
			if len(fixedChunks) > 1 {
				chunks = append(chunks, fixedChunks[:len(fixedChunks)-1]...)
			}

			// Keep the last one as our current chunk
			if len(fixedChunks) > 0 {
				currentChunk = fixedChunks[len(fixedChunks)-1]
			} else {
				currentChunk = ""
			}
		}
	}

	// Add the final chunk if not empty
	if strings.TrimSpace(currentChunk) != "" {
		chunks = append(chunks, strings.TrimSpace(currentChunk))
	}

	return chunks
}

// getOverlapText returns approximately 'overlap' characters from the end of the word list
func getOverlapText(words []string, overlapChars int) string {
	total := 0
	startIdx := len(words)

	// Count backwards from the end to find where to start the overlap
	for i := len(words) - 1; i >= 0; i-- {
		total += len(words[i]) + 1 // +1 for space
		if total >= overlapChars {
			startIdx = i
			break
		}
	}

	// Ensure we don't take too many words
	if startIdx < 0 {
		startIdx = 0
	}

	return strings.Join(words[startIdx:], " ")
}

// splitIntoSentences splits text into sentences using punctuation
func splitIntoSentences(text string) []string {
	var sentences []string
	var currentSentence strings.Builder

	// Simple state to avoid splitting on periods in abbreviations, numbers, etc.
	inAbbreviation := false

	for i, r := range text {
		currentSentence.WriteRune(r)

		// Check for potential end of sentence
		if r == '.' || r == '!' || r == '?' {
			// Look ahead to see if this is truly the end of a sentence
			if i+1 < len(text) {
				next := rune(text[i+1])

				// If followed by space and uppercase letter, likely a sentence boundary
				if unicode.IsSpace(next) && i+2 < len(text) && unicode.IsUpper(rune(text[i+2])) {
					sentences = append(sentences, currentSentence.String())
					currentSentence.Reset()
					inAbbreviation = false
					continue
				}

				// If this is a period followed by space, it's likely a sentence boundary
				if r == '.' && unicode.IsSpace(next) && !inAbbreviation {
					sentences = append(sentences, currentSentence.String())
					currentSentence.Reset()
					continue
				}
			} else if i+1 == len(text) {
				// End of text, add final sentence
				sentences = append(sentences, currentSentence.String())
				break
			}
		}

		// Track potential abbreviations
		if unicode.IsLetter(r) && i+1 < len(text) && text[i+1] == '.' {
			inAbbreviation = true
		} else if unicode.IsSpace(r) {
			inAbbreviation = false
		}
	}

	// Add any remaining text as a sentence
	if currentSentence.Len() > 0 {
		sentences = append(sentences, currentSentence.String())
	}

	return sentences
}

// allParagraphsUnderMaxSize checks if all paragraphs are smaller than maxSize
func allParagraphsUnderMaxSize(paragraphs []string, maxSize int) bool {
	for _, p := range paragraphs {
		if len(p) > maxSize {
			return false
		}
	}
	return true
}
