package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewDocument(t *testing.T) {
	// Test data
	content := "Test document content"
	metadata := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}

	// Create new document
	doc := NewDocument(content, metadata)

	// Validate the document
	if doc.Content != content {
		t.Errorf("Expected content %s, got %s", content, doc.Content)
	}

	if doc.ID == uuid.Nil {
		t.Error("Expected non-nil UUID")
	}

	if len(doc.Metadata) != 2 {
		t.Errorf("Expected metadata to have 2 keys, got %d", len(doc.Metadata))
	}

	if doc.Metadata["key1"] != "value1" {
		t.Errorf("Expected metadata[key1] = value1, got %v", doc.Metadata["key1"])
	}

	if doc.Metadata["key2"] != 42 {
		t.Errorf("Expected metadata[key2] = 42, got %v", doc.Metadata["key2"])
	}

	// Check creation and update times are set
	now := time.Now()
	if doc.CreatedAt.After(now) {
		t.Errorf("CreatedAt should not be in the future")
	}

	if doc.UpdatedAt.After(now) {
		t.Errorf("UpdatedAt should not be in the future")
	}

	// Created and updated should be the same initially
	if !doc.CreatedAt.Equal(doc.UpdatedAt) {
		t.Errorf("CreatedAt and UpdatedAt should be equal for new document")
	}
}

func TestNewEmbedding(t *testing.T) {
	// Test data
	docID := uuid.New()
	vector := []float32{0.1, 0.2, 0.3, 0.4, 0.5}

	// Create new embedding
	embedding := NewEmbedding(docID, vector)

	// Validate the embedding
	if embedding.DocumentID != docID {
		t.Errorf("Expected document ID %s, got %s", docID, embedding.DocumentID)
	}

	if embedding.ID == uuid.Nil {
		t.Error("Expected non-nil UUID for embedding ID")
	}

	if len(embedding.Vector) != len(vector) {
		t.Errorf("Expected vector length %d, got %d", len(vector), len(embedding.Vector))
	}

	for i, val := range embedding.Vector {
		if val != vector[i] {
			t.Errorf("Vector mismatch at index %d: expected %f, got %f", i, vector[i], val)
		}
	}

	// Check creation and update times are set
	now := time.Now()
	if embedding.CreatedAt.After(now) {
		t.Errorf("CreatedAt should not be in the future")
	}

	if embedding.UpdatedAt.After(now) {
		t.Errorf("UpdatedAt should not be in the future")
	}

	// Created and updated should be the same initially
	if !embedding.CreatedAt.Equal(embedding.UpdatedAt) {
		t.Errorf("CreatedAt and UpdatedAt should be equal for new embedding")
	}
}

func TestVectorQueryStructure(t *testing.T) {
	// Create a VectorQuery
	query := VectorQuery{
		Vector:    []float32{0.1, 0.2, 0.3},
		Limit:     10,
		Threshold: 0.7,
	}

	// Basic validation
	if len(query.Vector) != 3 {
		t.Errorf("Expected vector length 3, got %d", len(query.Vector))
	}

	if query.Limit != 10 {
		t.Errorf("Expected limit 10, got %d", query.Limit)
	}

	if query.Threshold != 0.7 {
		t.Errorf("Expected threshold 0.7, got %f", query.Threshold)
	}
}

func TestSearchResultStructure(t *testing.T) {
	// Create a Document
	doc := NewDocument("test content", nil)

	// Create a SearchResult
	result := SearchResult{
		Document:   doc,
		Similarity: 0.85,
	}

	// Basic validation
	if result.Document.ID != doc.ID {
		t.Errorf("Expected document ID %s, got %s", doc.ID, result.Document.ID)
	}

	if result.Document.Content != "test content" {
		t.Errorf("Expected content 'test content', got '%s'", result.Document.Content)
	}

	if result.Similarity != 0.85 {
		t.Errorf("Expected similarity 0.85, got %f", result.Similarity)
	}
}

func TestRAGQueryStructure(t *testing.T) {
	// Create a RAGQuery
	query := RAGQuery{
		Query: "What is RAG?",
		Limit: 5,
	}

	// Basic validation
	if query.Query != "What is RAG?" {
		t.Errorf("Expected query 'What is RAG?', got '%s'", query.Query)
	}

	if query.Limit != 5 {
		t.Errorf("Expected limit 5, got %d", query.Limit)
	}
}

func TestRAGResponseStructure(t *testing.T) {
	// Create test documents
	doc1 := NewDocument("Document 1", nil)
	doc2 := NewDocument("Document 2", nil)
	docs := []Document{doc1, doc2}

	// Create metadata
	metadata := map[string]interface{}{
		"response_time": 0.25,
	}

	// Create a RAGResponse
	response := RAGResponse{
		Answer:    "This is the answer",
		Documents: docs,
		Metadata:  metadata,
	}

	// Basic validation
	if response.Answer != "This is the answer" {
		t.Errorf("Expected answer 'This is the answer', got '%s'", response.Answer)
	}

	if len(response.Documents) != 2 {
		t.Errorf("Expected 2 documents, got %d", len(response.Documents))
	}

	if response.Documents[0].Content != "Document 1" {
		t.Errorf("Expected content 'Document 1', got '%s'", response.Documents[0].Content)
	}

	if response.Documents[1].Content != "Document 2" {
		t.Errorf("Expected content 'Document 2', got '%s'", response.Documents[1].Content)
	}

	// Check metadata
	meta, ok := response.Metadata.(map[string]interface{})
	if !ok {
		t.Errorf("Expected metadata to be a map[string]interface{}")
	} else {
		if val, ok := meta["response_time"]; !ok || val != 0.25 {
			t.Errorf("Expected metadata[response_time] = 0.25, got %v", val)
		}
	}
}
