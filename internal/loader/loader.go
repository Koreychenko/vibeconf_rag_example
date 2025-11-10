package loader

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yourusername/go-rag/internal/database"
	"github.com/yourusername/go-rag/internal/embeddings"
	"github.com/yourusername/go-rag/internal/models"
)

// SourceType represents the type of document source
type SourceType string

const (
	// TextFile represents a text file source
	TextFile SourceType = "text_file"
	// JSONFile represents a JSON file source
	JSONFile SourceType = "json_file"
	// Directory represents a directory containing document files
	Directory SourceType = "directory"
)

// DocumentSource represents a source of documents to load
type DocumentSource struct {
	Type     SourceType
	Path     string
	Metadata map[string]interface{}
}

// DocumentLoader handles loading documents into the RAG system
type DocumentLoader struct {
	db               database.VectorDB
	embeddingService embeddings.EmbeddingService
	chunkingOptions  ChunkingOptions
}

// NewDocumentLoader creates a new document loader
func NewDocumentLoader(
	db database.VectorDB,
	embeddingService embeddings.EmbeddingService,
	chunkingOptions ChunkingOptions,
) *DocumentLoader {
	// Use default chunking options if not provided
	if chunkingOptions.MaxChunkSize == 0 {
		chunkingOptions = DefaultChunkingOptions()
	}

	return &DocumentLoader{
		db:               db,
		embeddingService: embeddingService,
		chunkingOptions:  chunkingOptions,
	}
}

// LoadFromFile loads documents from a file
func (l *DocumentLoader) LoadFromFile(ctx context.Context, path string, metadata map[string]interface{}) error {
	// Check file exists
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to access file: %w", err)
	}

	if info.IsDir() {
		return l.loadFromDirectory(ctx, path, metadata)
	}

	// Determine file type based on extension
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".txt", ".md":
		return l.loadFromTextFile(ctx, path, metadata)
	case ".json":
		return fmt.Errorf("JSON file loading not implemented yet")
	default:
		// Attempt to load as text file
		return l.loadFromTextFile(ctx, path, metadata)
	}
}

// loadFromTextFile loads a document from a text file
func (l *DocumentLoader) loadFromTextFile(ctx context.Context, path string, metadata map[string]interface{}) error {
	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Create combined metadata
	meta := l.createFileMetadata(path, metadata)

	// Process the document
	return l.ProcessDocument(ctx, string(content), meta)
}

// loadFromDirectory loads all text files in a directory
func (l *DocumentLoader) loadFromDirectory(ctx context.Context, dirPath string, metadata map[string]interface{}) error {
	// Walk through directory
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file is a text file
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".txt" || ext == ".md" {
			// Create metadata for this file
			fileMeta := l.createFileMetadata(path, metadata)

			// Load the file
			if err := l.loadFromTextFile(ctx, path, fileMeta); err != nil {
				log.Printf("Warning: failed to load file %s: %v", path, err)
				// Continue processing other files
				return nil
			}
		}

		return nil
	})
}

// ProcessDocument processes a document text, chunks it, generates embeddings, and stores in the database
func (l *DocumentLoader) ProcessDocument(ctx context.Context, content string, metadata map[string]interface{}) error {
	// Skip empty documents
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("empty document content")
	}

	// Chunk the document
	chunks := ChunkText(content, l.chunkingOptions)

	// Log chunking result
	log.Printf("Document chunked into %d parts", len(chunks))

	// Process each chunk
	for i, chunk := range chunks {
		// Create chunk-specific metadata
		chunkMeta := l.createChunkMetadata(i, len(chunks), metadata)

		// Generate embedding for the chunk
		embedding, err := l.embeddingService.GenerateEmbedding(ctx, chunk)
		if err != nil {
			return fmt.Errorf("failed to generate embedding for chunk %d: %w", i, err)
		}

		// Create document model
		doc := models.NewDocument(chunk, chunkMeta)

		// Store document and embedding
		if err := l.db.StoreDocument(ctx, doc, embedding); err != nil {
			return fmt.Errorf("failed to store chunk %d: %w", i, err)
		}

		log.Printf("Stored chunk %d/%d", i+1, len(chunks))
	}

	return nil
}

// createFileMetadata creates metadata for a file
func (l *DocumentLoader) createFileMetadata(filePath string, baseMetadata map[string]interface{}) map[string]interface{} {
	// Start with a copy of the base metadata
	meta := make(map[string]interface{})
	for k, v := range baseMetadata {
		meta[k] = v
	}

	// Add file-specific metadata
	meta["source"] = "file"
	meta["file_path"] = filePath
	meta["file_name"] = filepath.Base(filePath)
	meta["file_ext"] = filepath.Ext(filePath)
	meta["loaded_at"] = time.Now()

	return meta
}

// createChunkMetadata creates metadata for a chunk
func (l *DocumentLoader) createChunkMetadata(chunkIndex, totalChunks int, baseMetadata map[string]interface{}) map[string]interface{} {
	// Start with a copy of the base metadata
	meta := make(map[string]interface{})
	for k, v := range baseMetadata {
		meta[k] = v
	}

	// Add chunk-specific metadata
	meta["chunk_index"] = chunkIndex
	meta["chunk_count"] = totalChunks

	return meta
}
