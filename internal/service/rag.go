package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/yourusername/go-rag/internal/config"
	"github.com/yourusername/go-rag/internal/database"
	"github.com/yourusername/go-rag/internal/embeddings"
	"github.com/yourusername/go-rag/internal/models"
)

// GeminiGenerationRequest represents a request to the Gemini API for text generation
type GeminiGenerationRequest struct {
	Contents []GeminiContent `json:"contents"`
}

// GeminiContent represents the content part of a Gemini request
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart represents a part of the content in a Gemini request
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiGenerationResponse represents a response from the Gemini API for text generation
type GeminiGenerationResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// RAGService provides Retrieval Augmented Generation functionality
type RAGService interface {
	AddDocument(ctx context.Context, content string, metadata map[string]interface{}) (string, error)
	SearchSimilar(ctx context.Context, query string, limit int) ([]models.SearchResult, error)
	Query(ctx context.Context, query string, limit int) (*models.RAGResponse, error)
}

// DefaultRAGService is the default implementation of the RAGService
type DefaultRAGService struct {
	db               database.VectorDB
	embeddingService embeddings.EmbeddingService
	geminiConfig     *config.GeminiConfig
	httpClient       *http.Client
}

// NewRAGService creates a new RAG service
func NewRAGService(
	db database.VectorDB,
	embeddingService embeddings.EmbeddingService,
	geminiConfig *config.GeminiConfig,
) (RAGService, error) {
	if db == nil {
		return nil, fmt.Errorf("database is required")
	}
	if embeddingService == nil {
		return nil, fmt.Errorf("embedding service is required")
	}
	if geminiConfig == nil {
		return nil, fmt.Errorf("Gemini config is required")
	}

	return &DefaultRAGService{
		db:               db,
		embeddingService: embeddingService,
		geminiConfig:     geminiConfig,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

// AddDocument adds a document to the RAG system
func (s *DefaultRAGService) AddDocument(
	ctx context.Context,
	content string,
	metadata map[string]interface{},
) (string, error) {
	if content == "" {
		return "", fmt.Errorf("document content cannot be empty")
	}

	// Create a new document
	doc := models.NewDocument(content, metadata)

	// Generate embedding for the document
	embedding, err := s.embeddingService.GenerateEmbedding(ctx, content)
	if err != nil {
		return "", fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Store document and embedding in the database
	if err := s.db.StoreDocument(ctx, doc, embedding); err != nil {
		return "", fmt.Errorf("failed to store document: %w", err)
	}

	return doc.ID.String(), nil
}

// SearchSimilar searches for documents similar to the query
func (s *DefaultRAGService) SearchSimilar(
	ctx context.Context,
	query string,
	limit int,
) ([]models.SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	if limit <= 0 {
		limit = 5 // Default limit
	}

	// Generate embedding for the query
	queryEmbedding, err := s.embeddingService.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Create vector query
	vectorQuery := models.VectorQuery{
		Vector:    queryEmbedding,
		Limit:     limit,
		Threshold: 0.0, // No threshold for now
	}

	// Search for similar documents
	results, err := s.db.FindSimilar(ctx, vectorQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to find similar documents: %w", err)
	}

	return results, nil
}

// Query performs a RAG query, retrieving relevant documents and generating a response
func (s *DefaultRAGService) Query(
	ctx context.Context,
	query string,
	limit int,
) (*models.RAGResponse, error) {
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	// Retrieve relevant documents
	results, err := s.SearchSimilar(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve documents: %w", err)
	}

	// Extract documents for the response
	var documents []models.Document
	for _, result := range results {
		documents = append(documents, result.Document)
	}

	// Augment query with document context
	augmentedQuery := s.augmentQueryWithContext(query, documents)

	// Generate response using Gemini
	answer, err := s.generateResponseWithGemini(ctx, augmentedQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	// Create RAG response
	response := &models.RAGResponse{
		Answer:    answer,
		Documents: documents,
	}

	return response, nil
}

// retrieveRelevantDocuments fetches documents relevant to the query
func (s *DefaultRAGService) retrieveRelevantDocuments(ctx context.Context, query string, limit int) ([]models.Document, error) {
	// This is a wrapper around SearchSimilar that extracts just the documents
	results, err := s.SearchSimilar(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	// Extract documents
	var documents []models.Document
	for _, result := range results {
		documents = append(documents, result.Document)
	}

	return documents, nil
}

// augmentQueryWithContext adds context from retrieved documents to the query
func (s *DefaultRAGService) augmentQueryWithContext(query string, documents []models.Document) string {
	if len(documents) == 0 {
		return query
	}

	var sb strings.Builder

	// Add context from documents
	sb.WriteString("Context information is below.\n")
	sb.WriteString("---------------------\n")

	for i, doc := range documents {
		sb.WriteString(fmt.Sprintf("Document %d:\n%s\n\n", i+1, doc.Content))
	}

	sb.WriteString("---------------------\n")
	sb.WriteString("Given the context information and not prior knowledge, answer the following query:\n")
	sb.WriteString(query)

	return sb.String()
}

// generateResponseWithGemini generates a response using Google's Gemini model
func (s *DefaultRAGService) generateResponseWithGemini(ctx context.Context, query string) (string, error) {
	// Create request body
	reqBody := GeminiGenerationRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{
						Text: query,
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1/models/%s:generateContent?key=%s",
		s.geminiConfig.TextModel, s.geminiConfig.APIKey)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		log.Printf("Gemini API error response: %s", string(body))
		return "", fmt.Errorf("API error (status %d)", resp.StatusCode)
	}

	// Parse response
	var genResponse GeminiGenerationResponse
	if err := json.Unmarshal(body, &genResponse); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Extract text from response
	if len(genResponse.Candidates) == 0 || len(genResponse.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return genResponse.Candidates[0].Content.Parts[0].Text, nil
}
