package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/go-rag/internal/models"
)

// MockRAGService is a mock implementation of the RAGService interface for testing
type MockRAGService struct {
	// AddDocument mocks
	AddDocumentFunc func(ctx context.Context, content string, metadata map[string]interface{}) (string, error)

	// SearchSimilar mocks
	SearchSimilarFunc func(ctx context.Context, query string, limit int) ([]models.SearchResult, error)

	// Query mocks
	QueryFunc func(ctx context.Context, query string, limit int) (*models.RAGResponse, error)
}

// AddDocument implements RAGService.AddDocument
func (m *MockRAGService) AddDocument(ctx context.Context, content string, metadata map[string]interface{}) (string, error) {
	return m.AddDocumentFunc(ctx, content, metadata)
}

// SearchSimilar implements RAGService.SearchSimilar
func (m *MockRAGService) SearchSimilar(ctx context.Context, query string, limit int) ([]models.SearchResult, error) {
	return m.SearchSimilarFunc(ctx, query, limit)
}

// Query implements RAGService.Query
func (m *MockRAGService) Query(ctx context.Context, query string, limit int) (*models.RAGResponse, error) {
	return m.QueryFunc(ctx, query, limit)
}

// setupTestRouter creates a test router with the given MockRAGService
func setupTestRouter(mockService *MockRAGService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	server := NewServer(mockService)
	return server.router
}

// TestHealthEndpoint tests the health check endpoint
func TestHealthEndpoint(t *testing.T) {
	router := setupTestRouter(&MockRAGService{})

	req := httptest.NewRequest("GET", "/health", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response["status"])
	}
}

// TestStoreDocumentHandler tests the document storage endpoint
func TestStoreDocumentHandler(t *testing.T) {
	// Create mock service
	mockService := &MockRAGService{
		AddDocumentFunc: func(ctx context.Context, content string, metadata map[string]interface{}) (string, error) {
			// Validate input
			if content == "" {
				t.Error("Empty content passed to AddDocument")
			}

			// Return a mock document ID
			return uuid.New().String(), nil
		},
	}

	router := setupTestRouter(mockService)

	// Create test request
	reqBody := DocumentRequest{
		Content:  "Test document content",
		Metadata: map[string]interface{}{"test": "metadata"},
	}

	jsonData, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/documents", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	// Check response
	if recorder.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, recorder.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if _, err := uuid.Parse(response["id"]); err != nil {
		t.Errorf("Response does not contain a valid UUID: %v", err)
	}
}

// TestSearchHandler tests the search endpoint
func TestSearchHandler(t *testing.T) {
	// Create mock service
	mockDoc := models.NewDocument("Test content", nil)
	mockResult := models.SearchResult{
		Document:   mockDoc,
		Similarity: 0.85,
	}

	mockService := &MockRAGService{
		SearchSimilarFunc: func(ctx context.Context, query string, limit int) ([]models.SearchResult, error) {
			// Validate input
			if query == "" {
				t.Error("Empty query passed to SearchSimilar")
			}

			if limit == 0 {
				// Default limit should be applied
				limit = 5
			}

			// Return mock results
			results := make([]models.SearchResult, limit)
			for i := 0; i < limit; i++ {
				results[i] = mockResult
			}
			return results, nil
		},
	}

	router := setupTestRouter(mockService)

	// Create test request
	reqBody := SearchRequest{
		Query: "test query",
		Limit: 3,
	}

	jsonData, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/search", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	// Check response
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	var results []models.SearchResult
	if err := json.Unmarshal(recorder.Body.Bytes(), &results); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	for _, result := range results {
		if result.Similarity != 0.85 {
			t.Errorf("Expected similarity 0.85, got %f", result.Similarity)
		}
	}
}

// TestQueryHandler tests the RAG query endpoint
func TestQueryHandler(t *testing.T) {
	// Create mock service
	mockDoc := models.NewDocument("Test content", nil)
	mockResponse := &models.RAGResponse{
		Answer:    "This is a test answer",
		Documents: []models.Document{mockDoc},
	}

	mockService := &MockRAGService{
		QueryFunc: func(ctx context.Context, query string, limit int) (*models.RAGResponse, error) {
			// Validate input
			if query == "" {
				t.Error("Empty query passed to Query")
			}

			return mockResponse, nil
		},
	}

	router := setupTestRouter(mockService)

	// Create test request
	reqBody := models.RAGQuery{
		Query: "test question?",
		Limit: 3,
	}

	jsonData, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/query", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	// Check response
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	var response models.RAGResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Answer != "This is a test answer" {
		t.Errorf("Expected answer 'This is a test answer', got '%s'", response.Answer)
	}

	if len(response.Documents) != 1 {
		t.Errorf("Expected 1 document, got %d", len(response.Documents))
	}
}

// TestGetDocumentHandler tests the document retrieval endpoint
func TestGetDocumentHandler(t *testing.T) {
	mockService := &MockRAGService{}
	router := setupTestRouter(mockService)

	// Create a valid UUID for testing
	validID := uuid.New().String()

	// Test with valid UUID
	req := httptest.NewRequest("GET", "/api/documents/"+validID, nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	// In the current implementation, GetDocumentHandler always returns a not implemented message
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	// Test with invalid UUID
	req = httptest.NewRequest("GET", "/api/documents/invalid-uuid", nil)
	recorder = httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d for invalid UUID, got %d", http.StatusBadRequest, recorder.Code)
	}
}

// TestListDocumentsHandler tests the document listing endpoint
func TestListDocumentsHandler(t *testing.T) {
	mockService := &MockRAGService{}
	router := setupTestRouter(mockService)

	// Test with default parameters
	req := httptest.NewRequest("GET", "/api/documents", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	// In the current implementation, ListDocumentsHandler always returns a not implemented message
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	// Test with custom parameters
	req = httptest.NewRequest("GET", "/api/documents?limit=20&offset=10", nil)
	recorder = httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
	}
}

// TestDeleteDocumentHandler tests the document deletion endpoint
func TestDeleteDocumentHandler(t *testing.T) {
	mockService := &MockRAGService{}
	router := setupTestRouter(mockService)

	// Create a valid UUID for testing
	validID := uuid.New().String()

	// Test with valid UUID
	req := httptest.NewRequest("DELETE", "/api/documents/"+validID, nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	// In the current implementation, DeleteDocumentHandler always returns a not implemented message
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	// Test with invalid UUID
	req = httptest.NewRequest("DELETE", "/api/documents/invalid-uuid", nil)
	recorder = httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d for invalid UUID, got %d", http.StatusBadRequest, recorder.Code)
	}
}

// TestInvalidJSONInput tests error handling for invalid JSON input
func TestInvalidJSONInput(t *testing.T) {
	mockService := &MockRAGService{}
	router := setupTestRouter(mockService)

	// Test invalid JSON for document storage
	req := httptest.NewRequest("POST", "/api/documents", bytes.NewBuffer([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d for invalid JSON, got %d", http.StatusBadRequest, recorder.Code)
	}

	// Test invalid JSON for search
	req = httptest.NewRequest("POST", "/api/search", bytes.NewBuffer([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	recorder = httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d for invalid JSON, got %d", http.StatusBadRequest, recorder.Code)
	}

	// Test invalid JSON for query
	req = httptest.NewRequest("POST", "/api/query", bytes.NewBuffer([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	recorder = httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d for invalid JSON, got %d", http.StatusBadRequest, recorder.Code)
	}
}
