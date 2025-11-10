package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/go-rag/internal/config"
	"github.com/yourusername/go-rag/internal/database"
	"github.com/yourusername/go-rag/internal/embeddings"
	"github.com/yourusername/go-rag/internal/loader"
)

// CLI flags
var (
	dataDir       string
	filePath      string
	chunkStrategy string
	chunkSize     int
	chunkOverlap  int
)

func init() {
	// Define command line flags
	flag.StringVar(&dataDir, "dir", "", "Directory containing document files to load")
	flag.StringVar(&filePath, "file", "", "Single document file to load")
	flag.StringVar(&chunkStrategy, "strategy", "paragraph", "Chunking strategy (paragraph, sentence, fixed_size)")
	flag.IntVar(&chunkSize, "chunk-size", 1000, "Maximum size of chunks in characters")
	flag.IntVar(&chunkOverlap, "chunk-overlap", 100, "Overlap between chunks in characters")
}

func main() {
	// Parse command-line flags
	flag.Parse()

	// Validate input
	if dataDir == "" && filePath == "" {
		log.Fatal("Either -dir or -file must be specified")
	}

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

	// Convert chunking strategy string to the appropriate enum
	var chunkingStrategy loader.ChunkingStrategy
	switch chunkStrategy {
	case "paragraph":
		chunkingStrategy = loader.ByParagraph
	case "sentence":
		chunkingStrategy = loader.BySentence
	case "fixed_size":
		chunkingStrategy = loader.ByFixedSize
	default:
		log.Fatalf("Unknown chunking strategy: %s", chunkStrategy)
	}

	// Initialize chunking options
	chunkingOptions := loader.ChunkingOptions{
		Strategy:     chunkingStrategy,
		MaxChunkSize: chunkSize,
		ChunkOverlap: chunkOverlap,
	}

	// Initialize document loader
	documentLoader := loader.NewDocumentLoader(db, embeddingService, chunkingOptions)

	// Start the loading process
	startTime := time.Now()
	log.Println("Starting document loading process...")

	// Process either a directory or a single file
	if dataDir != "" {
		// Create base metadata for this loading session
		metadata := map[string]interface{}{
			"loaded_by": "data_loader",
			"batch_id":  time.Now().Format("20060102-150405"),
		}

		log.Printf("Loading documents from directory: %s", dataDir)
		if err := documentLoader.LoadFromFile(ctx, dataDir, metadata); err != nil {
			log.Fatalf("Failed to load documents from directory: %v", err)
		}
	} else if filePath != "" {
		// Create metadata for this file
		metadata := map[string]interface{}{
			"loaded_by": "data_loader",
			"batch_id":  time.Now().Format("20060102-150405"),
		}

		log.Printf("Loading document from file: %s", filePath)
		if err := documentLoader.LoadFromFile(ctx, filePath, metadata); err != nil {
			log.Fatalf("Failed to load document from file: %v", err)
		}
	}

	// Report completion
	elapsed := time.Since(startTime)
	log.Printf("Document loading completed in %v", elapsed)
}
