package config

/*
This file defines the configuration structure for the RAG system.

Key responsibilities:
- Load configuration from environment variables or files
- Validate configuration parameters
- Provide access to configuration values
*/

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config represents the application configuration
type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	Gemini     GeminiConfig
	Embeddings EmbeddingsConfig
}

// ServerConfig contains server-related configuration
type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// DatabaseConfig contains database-related configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// GeminiConfig contains Google Gemini API configuration
type GeminiConfig struct {
	APIKey         string
	TextModel      string
	EmbeddingModel string
}

// EmbeddingsConfig contains embedding-related configuration
type EmbeddingsConfig struct {
	Dimensions int
}

// LoadConfig loads the application configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Server configuration
	serverPort, err := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid server port: %w", err)
	}

	// Database configuration
	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid database port: %w", err)
	}

	// Embedding dimensions
	dimensions, err := strconv.Atoi(getEnv("EMBEDDING_DIMENSIONS", "768"))
	if err != nil {
		return nil, fmt.Errorf("invalid embedding dimensions: %w", err)
	}

	// API key validation
	geminiAPIKey := getEnv("GEMINI_API_KEY", "")
	if geminiAPIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required")
	}

	return &Config{
		Server: ServerConfig{
			Port:         serverPort,
			ReadTimeout:  time.Second * 15,
			WriteTimeout: time.Second * 15,
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "ragdb"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Gemini: GeminiConfig{
			APIKey:         geminiAPIKey,
			TextModel:      getEnv("GEMINI_TEXT_MODEL", "gemini-1.5-pro"),
			EmbeddingModel: getEnv("GEMINI_EMBEDDING_MODEL", "embedding-001"),
		},
		Embeddings: EmbeddingsConfig{
			Dimensions: dimensions,
		},
	}, nil
}

// ConnectionString returns the PostgreSQL connection string based on the configuration
func (c *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}

// Helper function to get environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
