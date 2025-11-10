package database

import (
	"testing"
)

// TestNewPostgresVectorDB tests the constructor for PostgresVectorDB
func TestNewPostgresVectorDB(t *testing.T) {
	// Test with valid parameters
	connectionString := "postgres://user:password@localhost:5432/testdb?sslmode=disable"
	dimensions := 768

	db, err := NewPostgresVectorDB(connectionString, dimensions)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if db == nil {
		t.Error("Expected non-nil VectorDB instance")
	}

	// Test type casting (implementation detail)
	_, ok := db.(*PostgresVectorDB)
	if !ok {
		t.Error("Expected *PostgresVectorDB type")
	}
}
