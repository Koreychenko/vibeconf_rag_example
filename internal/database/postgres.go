package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/pgvector/pgvector-go"

	"github.com/yourusername/go-rag/internal/models"
)

// VectorDB defines the interface for vector database operations
type VectorDB interface {
	Connect(ctx context.Context) error
	Close() error
	StoreDocument(ctx context.Context, doc models.Document, embedding []float32) error
	FindSimilar(ctx context.Context, query models.VectorQuery) ([]models.SearchResult, error)
	GetDocument(ctx context.Context, id uuid.UUID) (models.Document, error)
	ListDocuments(ctx context.Context, limit, offset int) ([]models.Document, error)
	DeleteDocument(ctx context.Context, id uuid.UUID) error
}

// PostgresVectorDB is a PostgreSQL implementation of VectorDB with pgvector extension
type PostgresVectorDB struct {
	connStr    string
	db         *sql.DB
	dimensions int
}

// NewPostgresVectorDB creates a new PostgreSQL vector database connection
func NewPostgresVectorDB(connectionString string, dimensions int) (VectorDB, error) {
	return &PostgresVectorDB{
		connStr:    connectionString,
		dimensions: dimensions,
	}, nil
}

// Connect establishes a connection to the database
func (p *PostgresVectorDB) Connect(ctx context.Context) error {
	db, err := sql.Open("postgres", p.connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	p.db = db
	log.Println("Successfully connected to the database")

	return nil
}

// Close closes the database connection
func (p *PostgresVectorDB) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// StoreDocument stores a document and its embedding in the database
func (p *PostgresVectorDB) StoreDocument(ctx context.Context, doc models.Document, embedding []float32) error {
	if p.db == nil {
		return fmt.Errorf("database not connected")
	}

	// Begin transaction
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Convert metadata to JSON
	metadataJSON, err := json.Marshal(doc.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Insert document
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO rag.documents (id, content, metadata, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)",
		doc.ID, doc.Content, metadataJSON, doc.CreatedAt, doc.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert document: %w", err)
	}

	// Convert embedding to pgvector
	vector := pgvector.NewVector(embedding)

	// Insert embedding
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO rag.embeddings (id, document_id, embedding, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)",
		uuid.New(), doc.ID, vector, time.Now(), time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to insert embedding: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// FindSimilar finds documents similar to the query vector
func (p *PostgresVectorDB) FindSimilar(ctx context.Context, query models.VectorQuery) ([]models.SearchResult, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	// Convert query vector to pgvector
	queryVector := pgvector.NewVector(query.Vector)

	// Use the similarity search function
	rows, err := p.db.QueryContext(
		ctx,
		`SELECT d.id, d.content, d.metadata, similarity
		 FROM rag.search_similar_documents($1, $2, $3) as sr
		 JOIN rag.documents d ON sr.id = d.id`,
		queryVector, 0.0, query.Limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute similarity search: %w", err)
	}
	defer rows.Close()

	var results []models.SearchResult
	for rows.Next() {
		var doc models.Document
		var metadataJSON []byte
		var similarity float32

		if err := rows.Scan(&doc.ID, &doc.Content, &metadataJSON, &similarity); err != nil {
			return nil, fmt.Errorf("failed to scan result row: %w", err)
		}

		// Parse metadata
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &doc.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		results = append(results, models.SearchResult{
			Document:   doc,
			Similarity: similarity,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating result rows: %w", err)
	}

	return results, nil
}

// GetDocument retrieves a document by ID
func (p *PostgresVectorDB) GetDocument(ctx context.Context, id uuid.UUID) (models.Document, error) {
	if p.db == nil {
		return models.Document{}, fmt.Errorf("database not connected")
	}

	var doc models.Document
	var metadataJSON []byte

	err := p.db.QueryRowContext(
		ctx,
		"SELECT id, content, metadata, created_at, updated_at FROM rag.documents WHERE id = $1",
		id,
	).Scan(&doc.ID, &doc.Content, &metadataJSON, &doc.CreatedAt, &doc.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.Document{}, fmt.Errorf("document not found")
		}
		return models.Document{}, fmt.Errorf("failed to get document: %w", err)
	}

	// Parse metadata
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &doc.Metadata); err != nil {
			return models.Document{}, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return doc, nil
}

// ListDocuments retrieves a list of documents with pagination
func (p *PostgresVectorDB) ListDocuments(ctx context.Context, limit, offset int) ([]models.Document, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	// Set default values if needed
	if limit <= 0 {
		limit = 10
	}

	rows, err := p.db.QueryContext(
		ctx,
		"SELECT id, content, metadata, created_at, updated_at FROM rag.documents ORDER BY created_at DESC LIMIT $1 OFFSET $2",
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}
	defer rows.Close()

	var documents []models.Document
	for rows.Next() {
		var doc models.Document
		var metadataJSON []byte

		if err := rows.Scan(&doc.ID, &doc.Content, &metadataJSON, &doc.CreatedAt, &doc.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan document row: %w", err)
		}

		// Parse metadata
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &doc.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		documents = append(documents, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating document rows: %w", err)
	}

	return documents, nil
}

// DeleteDocument deletes a document and its embedding by ID
func (p *PostgresVectorDB) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	if p.db == nil {
		return fmt.Errorf("database not connected")
	}

	// Start a transaction
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Delete document (cascade will delete embeddings)
	result, err := tx.ExecContext(
		ctx,
		"DELETE FROM rag.documents WHERE id = $1",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("document not found")
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
