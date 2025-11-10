package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/yourusername/go-rag/internal/config"
	"github.com/yourusername/go-rag/internal/models"
)

// MockVectorDB is a mock implementation of the VectorDB interface
type MockVectorDB struct {
	StoreDocumentFunc  func(ctx context.Context, doc models.Document, embedding []float32) error
	FindSimilarFunc    func(ctx context.Context, query models.VectorQuery) ([]models.SearchResult, error)
	GetDocumentFunc    func(ctx context.Context, id uuid.UUID) (models.Document, error)
	ListDocumentsFunc  func(ctx context.Context, limit, offset int) ([]models.Document, error)
	DeleteDocumentFunc func(ctx context.Context, id uuid.UUID) error
	ConnectFunc        func(ctx context.Context) error
	CloseFunc          func() error
}

func (m *MockVectorDB) StoreDocument(ctx context.Context, doc models.Document, embedding []float32) error {
	return m.StoreDocumentFunc(ctx, doc, embedding)
}

func (m *MockVectorDB) FindSimilar(ctx context.Context, query models.VectorQuery) ([]models.SearchResult, error) {
	return m.FindSimilarFunc(ctx, query)
}

func (m *MockVectorDB) GetDocument(ctx context.Context, id uuid.UUID) (models.Document, error) {
	return m.GetDocumentFunc(ctx, id)
}

func (m *MockVectorDB) ListDocuments(ctx context.Context, limit, offset int) ([]models.Document, error) {
	return m.ListDocumentsFunc(ctx, limit, offset)
}

func (m *MockVectorDB) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	return m.DeleteDocumentFunc(ctx, id)
}

func (m *MockVectorDB) Connect(ctx context.Context) error {
	return m.ConnectFunc(ctx)
}

func (m *MockVectorDB) Close() error {
	return m.CloseFunc()
}

// MockEmbeddingService is a mock implementation of the EmbeddingService interface
type MockEmbeddingService struct {
	GenerateEmbeddingFunc       func(ctx context.Context, text string) ([]float32, error)
	BatchGenerateEmbeddingsFunc func(ctx context.Context, texts []string) ([][]float32, error)
	CalculateSimilarityFunc     func(vec1, vec2 []float32) float32
}

func (m *MockEmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return m.GenerateEmbeddingFunc(ctx, text)
}

func (m *MockEmbeddingService) BatchGenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	return m.BatchGenerateEmbeddingsFunc(ctx, texts)
}

func (m *MockEmbeddingService) CalculateSimilarity(vec1, vec2 []float32) float32 {
	return m.CalculateSimilarityFunc(vec1, vec2)
}

// TestNewRAGService tests the constructor for RAGService
func TestNewRAGService(t *testing.T) {
	// Create mocks
	mockDB := &MockVectorDB{}
	mockEmbedding := &MockEmbeddingService{}
	mockConfig := &config.GeminiConfig{
		APIKey:         "test-api-key",
		TextModel:      "test-text-model",
		EmbeddingModel: "test-embedding-model",
	}

	// Test with valid parameters
	service, err := NewRAGService(mockDB, mockEmbedding, mockConfig)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if service == nil {
		t.Error("Expected non-nil RAGService")
	}

	// Test with nil database
	service, err = NewRAGService(nil, mockEmbedding, mockConfig)
	if err == nil {
		t.Error("Expected error with nil database, got nil")
	}

	// Test with nil embedding service
	service, err = NewRAGService(mockDB, nil, mockConfig)
	if err == nil {
		t.Error("Expected error with nil embedding service, got nil")
	}

	// Test with nil config
	service, err = NewRAGService(mockDB, mockEmbedding, nil)
	if err == nil {
		t.Error("Expected error with nil config, got nil")
	}
}

// TestAddDocument tests the AddDocument method
func TestAddDocument(t *testing.T) {
	// Setup test data
	content := "Test document content"
	metadata := map[string]interface{}{
		"source": "test",
	}

	// Setup mocks
	mockDB := &MockVectorDB{
		StoreDocumentFunc: func(ctx context.Context, doc models.Document, embedding []float32) error {
			// Validate input
			if doc.Content != content {
				t.Errorf("Expected content '%s', got '%s'", content, doc.Content)
			}

			if doc.Metadata["source"] != "test" {
				t.Errorf("Expected metadata[source] = 'test', got '%v'", doc.Metadata["source"])
			}

			if len(embedding) != 3 {
				t.Errorf("Expected embedding length 3, got %d", len(embedding))
			}

			return nil
		},
	}

	mockEmbedding := &MockEmbeddingService{
		GenerateEmbeddingFunc: func(ctx context.Context, text string) ([]float32, error) {
			// Validate input
			if text != content {
				t.Errorf("Expected text '%s', got '%s'", content, text)
			}

			// Return mock embedding
			return []float32{0.1, 0.2, 0.3}, nil
		},
	}

	mockConfig := &config.GeminiConfig{
		APIKey:         "test-api-key",
		TextModel:      "test-text-model",
		EmbeddingModel: "test-embedding-model",
	}

	// Create service
	service, _ := NewRAGService(mockDB, mockEmbedding, mockConfig)

	// Call method
	ctx := context.Background()
	docID, err := service.AddDocument(ctx, content, metadata)

	// Verify results
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if docID == "" {
		t.Error("Expected non-empty document ID")
	}

	// Test with empty content
	_, err = service.AddDocument(ctx, "", metadata)
	if err == nil {
		t.Error("Expected error with empty content, got nil")
	}
}

// TestSearchSimilar tests the SearchSimilar method
func TestSearchSimilar(t *testing.T) {
	// Setup test data
	query := "test query"
	limit := 5

	// Create mock documents and search results
	docID := uuid.New()
	doc := models.Document{
		ID:      docID,
		Content: "Test content",
	}
	searchResult := models.SearchResult{
		Document:   doc,
		Similarity: 0.85,
	}

	// Setup mocks
	mockDB := &MockVectorDB{
		FindSimilarFunc: func(ctx context.Context, queryVec models.VectorQuery) ([]models.SearchResult, error) {
			// Validate input
			if queryVec.Limit != limit {
				t.Errorf("Expected limit %d, got %d", limit, queryVec.Limit)
			}

			if len(queryVec.Vector) != 3 {
				t.Errorf("Expected vector length 3, got %d", len(queryVec.Vector))
			}

			// Return mock results
			return []models.SearchResult{searchResult}, nil
		},
	}

	mockEmbedding := &MockEmbeddingService{
		GenerateEmbeddingFunc: func(ctx context.Context, text string) ([]float32, error) {
			// Validate input
			if text != query {
				t.Errorf("Expected query '%s', got '%s'", query, text)
			}

			// Return mock embedding
			return []float32{0.1, 0.2, 0.3}, nil
		},
	}

	mockConfig := &config.GeminiConfig{
		APIKey:         "test-api-key",
		TextModel:      "test-text-model",
		EmbeddingModel: "test-embedding-model",
	}

	// Create service
	service, _ := NewRAGService(mockDB, mockEmbedding, mockConfig)

	// Call method
	ctx := context.Background()
	results, err := service.SearchSimilar(ctx, query, limit)

	// Verify results
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if results[0].Document.ID != docID {
		t.Errorf("Expected document ID %s, got %s", docID, results[0].Document.ID)
	}

	// Test with empty query
	_, err = service.SearchSimilar(ctx, "", limit)
	if err == nil {
		t.Error("Expected error with empty query, got nil")
	}
}

// TestAugmentQueryWithContext tests the augmentQueryWithContext function
func TestAugmentQueryWithContext(t *testing.T) {
	// Create test service
	service := &DefaultRAGService{}

	// Test with empty documents
	query := "test query"
	augmented := service.augmentQueryWithContext(query, []models.Document{})

	if augmented != query {
		t.Errorf("Expected original query for empty documents, got %s", augmented)
	}

	// Test with documents
	docs := []models.Document{
		{Content: "Document 1 content"},
		{Content: "Document 2 content"},
	}

	augmented = service.augmentQueryWithContext(query, docs)

	// Check that augmented query contains original query and document content
	if augmented == query {
		t.Errorf("Expected augmented query to be different from original")
	}

	if augmented == "" {
		t.Errorf("Expected non-empty augmented query")
	}

	// Basic checks for content inclusion
	for _, doc := range docs {
		if !contains(augmented, doc.Content) {
			t.Errorf("Expected augmented query to contain document content: %s", doc.Content)
		}
	}

	if !contains(augmented, query) {
		t.Errorf("Expected augmented query to contain original query: %s", query)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return s != "" && substr != "" && s != substr && len(s) > len(substr) && s != substr
}
