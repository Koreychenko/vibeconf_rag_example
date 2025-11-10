package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/yourusername/go-rag/internal/config"
)

// GeminiEmbeddingRequest represents a request to the Gemini Embedding API
type GeminiEmbeddingRequest struct {
	Content struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"content"`
}

// GeminiEmbeddingResponse represents a response from the Gemini Embedding API
type GeminiEmbeddingResponse struct {
	Embedding struct {
		Values []float32 `json:"values"`
	} `json:"embedding"`
}

// EmbeddingService provides functionality for generating and working with embeddings
type EmbeddingService interface {
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
	BatchGenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error)
	CalculateSimilarity(vec1, vec2 []float32) float32
}

// GeminiEmbeddingService is an implementation of EmbeddingService using Google's Gemini API
type GeminiEmbeddingService struct {
	apiKey         string
	embeddingModel string
	httpClient     *http.Client
}

// NewGeminiEmbeddingService creates a new embedding service using Google's Gemini API
func NewGeminiEmbeddingService(cfg *config.GeminiConfig) (EmbeddingService, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("Gemini API key is required")
	}

	return &GeminiEmbeddingService{
		apiKey:         cfg.APIKey,
		embeddingModel: cfg.EmbeddingModel,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// GenerateEmbedding generates an embedding vector for the given text
func (s *GeminiEmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	// Clean and prepare text
	text = strings.TrimSpace(text)

	// Create request body
	reqBody := GeminiEmbeddingRequest{}
	reqBody.Content.Parts = []struct {
		Text string `json:"text"`
	}{
		{Text: text},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1/models/%s:embedContent?key=%s",
		s.embeddingModel, s.apiKey)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var embResponse GeminiEmbeddingResponse
	if err := json.Unmarshal(body, &embResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return embResponse.Embedding.Values, nil
}

// BatchGenerateEmbeddings generates embedding vectors for multiple texts
func (s *GeminiEmbeddingService) BatchGenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("texts cannot be empty")
	}

	// Process each text sequentially
	// This could be optimized with concurrent requests in a production system
	var embeddings [][]float32
	for _, text := range texts {
		embedding, err := s.GenerateEmbedding(ctx, text)
		if err != nil {
			return nil, err
		}
		embeddings = append(embeddings, embedding)
	}

	return embeddings, nil
}

// CalculateSimilarity calculates cosine similarity between two vectors
func (s *GeminiEmbeddingService) CalculateSimilarity(vec1, vec2 []float32) float32 {
	if len(vec1) != len(vec2) {
		return 0
	}

	var dotProduct float64
	var norm1 float64
	var norm2 float64

	for i := 0; i < len(vec1); i++ {
		dotProduct += float64(vec1[i] * vec2[i])
		norm1 += float64(vec1[i] * vec1[i])
		norm2 += float64(vec2[i] * vec2[i])
	}

	// Avoid division by zero
	if norm1 == 0 || norm2 == 0 {
		return 0
	}

	similarity := dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
	return float32(similarity)
}
