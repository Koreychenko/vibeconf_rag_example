package models

import (
	"time"

	"github.com/google/uuid"
)

// Document represents a document stored in the RAG system
type Document struct {
	ID        uuid.UUID              `json:"id"`
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// NewDocument creates a new document with the given content and metadata
func NewDocument(content string, metadata map[string]interface{}) Document {
	now := time.Now()
	return Document{
		ID:        uuid.New(),
		Content:   content,
		Metadata:  metadata,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Embedding represents a vector embedding of a document
type Embedding struct {
	ID         uuid.UUID `json:"id"`
	DocumentID uuid.UUID `json:"document_id"`
	Vector     []float32 `json:"vector"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// NewEmbedding creates a new embedding for a document
func NewEmbedding(documentID uuid.UUID, vector []float32) Embedding {
	now := time.Now()
	return Embedding{
		ID:         uuid.New(),
		DocumentID: documentID,
		Vector:     vector,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// VectorQuery represents a vector similarity search query
type VectorQuery struct {
	Vector    []float32 `json:"vector"`
	Limit     int       `json:"limit"`
	Threshold float32   `json:"threshold"`
}

// SearchResult represents the result of a vector similarity search
type SearchResult struct {
	Document   Document `json:"document"`
	Similarity float32  `json:"similarity"`
}

// RAGQuery represents a query for the RAG system
type RAGQuery struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

// RAGResponse represents the response from the RAG system
type RAGResponse struct {
	Answer    string      `json:"answer"`
	Documents []Document  `json:"documents,omitempty"`
	Metadata  interface{} `json:"metadata,omitempty"`
}
