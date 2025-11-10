package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourusername/go-rag/internal/api"
	"github.com/yourusername/go-rag/internal/config"
	"github.com/yourusername/go-rag/internal/database"
	"github.com/yourusername/go-rag/internal/embeddings"
	"github.com/yourusername/go-rag/internal/service"
)

func main() {
	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %v, shutting down gracefully", sig)
		cancel()
	}()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, err := database.NewPostgresVectorDB(cfg.Database.ConnectionString(), cfg.Embeddings.Dimensions)
	if err != nil {
		log.Fatalf("Failed to create database connection: %v", err)
	}

	// Connect to the database
	if err := db.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize embedding service
	embeddingService, err := embeddings.NewGeminiEmbeddingService(&cfg.Gemini)
	if err != nil {
		log.Fatalf("Failed to initialize embedding service: %v", err)
	}

	// Initialize RAG service
	ragService, err := service.NewRAGService(db, embeddingService, &cfg.Gemini)
	if err != nil {
		log.Fatalf("Failed to initialize RAG service: %v", err)
	}

	// Initialize API server
	server := api.NewServer(ragService)

	// Start the HTTP server in a goroutine
	serverPort := fmt.Sprintf("%d", cfg.Server.Port)
	log.Printf("Starting server on port %s", serverPort)

	go func() {
		if err := server.Start(serverPort); err != nil {
			log.Printf("Server stopped: %v", err)
			cancel()
		}
	}()

	// Wait for cancellation
	<-ctx.Done()
	log.Println("Server shutdown complete")
}
