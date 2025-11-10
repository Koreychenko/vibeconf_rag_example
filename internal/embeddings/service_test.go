package embeddings

import (
	"context"
	"net/http"
	"testing"

	"github.com/yourusername/go-rag/internal/config"
)

// TestCalculateSimilarity tests the vector similarity calculation
func TestCalculateSimilarity(t *testing.T) {
	// Create a service instance - we don't need API connectivity for this test
	service := &GeminiEmbeddingService{
		apiKey:         "test-key",
		embeddingModel: "test-model",
		httpClient:     &http.Client{},
	}

	testCases := []struct {
		name     string
		vec1     []float32
		vec2     []float32
		expected float32
	}{
		{
			name:     "identical vectors",
			vec1:     []float32{1.0, 2.0, 3.0},
			vec2:     []float32{1.0, 2.0, 3.0},
			expected: 1.0, // Identical vectors should have similarity 1.0
		},
		{
			name:     "orthogonal vectors",
			vec1:     []float32{1.0, 0.0, 0.0},
			vec2:     []float32{0.0, 1.0, 0.0},
			expected: 0.0, // Orthogonal vectors should have similarity 0.0
		},
		{
			name:     "similar vectors",
			vec1:     []float32{1.0, 2.0, 3.0},
			vec2:     []float32{2.0, 3.0, 4.0},
			expected: 0.99, // Approximate value, using a wider range for comparison
		},
		{
			name:     "different length vectors",
			vec1:     []float32{1.0, 2.0, 3.0},
			vec2:     []float32{1.0, 2.0},
			expected: 0.0, // Different lengths should return 0
		},
		{
			name:     "zero vector",
			vec1:     []float32{0.0, 0.0, 0.0},
			vec2:     []float32{1.0, 2.0, 3.0},
			expected: 0.0, // Zero vector should return 0
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			similarity := service.CalculateSimilarity(tc.vec1, tc.vec2)

			// Using a wider epsilon for floating point comparison
			const epsilon = 0.01
			if similarity < tc.expected-epsilon || similarity > tc.expected+epsilon {
				t.Errorf("Expected similarity around %f, got %f", tc.expected, similarity)
			}
		})
	}
}

// TestNewGeminiEmbeddingService tests the service constructor
func TestNewGeminiEmbeddingService(t *testing.T) {
	testCases := []struct {
		name        string
		cfg         *config.GeminiConfig
		expectError bool
	}{
		{
			name: "valid config",
			cfg: &config.GeminiConfig{
				APIKey:         "test-key",
				TextModel:      "text-model",
				EmbeddingModel: "embedding-model",
			},
			expectError: false,
		},
		{
			name: "empty API key",
			cfg: &config.GeminiConfig{
				APIKey:         "",
				TextModel:      "text-model",
				EmbeddingModel: "embedding-model",
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service, err := NewGeminiEmbeddingService(tc.cfg)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error, but got nil")
				}
				if service != nil {
					t.Errorf("Expected nil service, but got non-nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
				if service == nil {
					t.Errorf("Expected non-nil service, but got nil")
				}
			}
		})
	}
}

// TestGenerateEmbeddingEmptyText tests handling of empty text input
func TestGenerateEmbeddingEmptyText(t *testing.T) {
	service := &GeminiEmbeddingService{
		apiKey:         "test-key",
		embeddingModel: "test-model",
		httpClient:     &http.Client{},
	}

	ctx := context.Background()
	_, err := service.GenerateEmbedding(ctx, "")

	if err == nil {
		t.Errorf("Expected error for empty text, got nil")
	}
}
