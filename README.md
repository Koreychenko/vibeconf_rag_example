# Go RAG System with Gemini and PostgreSQL/pgvector

A Retrieval Augmented Generation (RAG) system built with Go, PostgreSQL with pgvector extension, and Google Gemini AI.

## Architecture

The system implements a RAG (Retrieval Augmented Generation) pattern:

1. **Retrieval**: The system retrieves relevant documents from a vector database based on similarity to the query.
2. **Augmentation**: The retrieved information augments the context for generation.
3. **Generation**: The system generates a response using Google Gemini.

## Features

- Document storage and retrieval with vector embeddings
- Semantic search using vector similarity
- RAG-based query answering with Google Gemini
- Document chunking with multiple strategies (paragraph, sentence, fixed-size)
- Containerized deployment with Docker
- RESTful API for all operations
- Makefile for common development tasks
- Bulk data loading from files and directories

## Requirements

- Docker and Docker Compose
- Google Gemini API key
- Go 1.21 or later (for local development)

## Getting Started

1. Clone the repository
2. Configure the `.env` file (see Environment Variables section)
3. Build and run the containers:

```bash
# Using docker-compose directly
docker-compose -f deployments/docker-compose.yml up -d

# Or using the Makefile
make docker-up
```

4. Load sample data (optional):

```bash
# Using the Makefile
make docker-load-samples
```

## Environment Variables

Create a `.env` file in the project root with the following variables:

```
# Server configuration
SERVER_PORT=8080

# Database configuration
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=ragdb
DB_SSL_MODE=disable

# Google Gemini API configuration
GEMINI_API_KEY=your-gemini-api-key  # Replace with your actual API key
GEMINI_TEXT_MODEL=gemini-2.5-flash
GEMINI_EMBEDDING_MODEL=embedding-001

# Vector dimensions for embeddings
EMBEDDING_DIMENSIONS=768
```

## Makefile Commands

The project includes a Makefile with various commands to simplify development:

```
# Build and run
make build         # Build the application
make run           # Run the application locally
make clean         # Clean build artifacts
make fmt           # Format Go code
make test          # Run tests

# Docker operations
make docker-up     # Start all containers in the background
make docker-down   # Stop all containers
make docker-logs   # Start all containers with logs in foreground
make docker-rebuild # Rebuild and restart only the app container
make docker-ps     # Show Docker container status

# Data loading
make build-loader  # Build the data loader
make load-samples  # Load sample data locally
make docker-load-samples # Load sample data in Docker container

# Development setup
make dev-setup     # Set up the development environment
make help          # Show help message
```

## Data Loading

The system includes a data loader tool that can:

- Load documents from individual files or directories
- Automatically chunk documents using different strategies
- Generate embeddings for each chunk
- Store documents and embeddings in the database

### Chunking Strategies

- **Paragraph**: Chunks text by paragraphs (default)
- **Sentence**: Chunks text by sentences
- **Fixed Size**: Chunks text by a fixed number of characters

### Loading Data

```bash
# Load from a directory
./dataloader -dir ./path/to/documents

# Load a single file
./dataloader -file ./path/to/document.txt

# Customize chunking
./dataloader -dir ./data/samples -strategy sentence -chunk-size 500 -chunk-overlap 50
```

## API Endpoints

- `GET /health` - Health check endpoint
- `POST /api/documents` - Store a document
- `GET /api/documents/{id}` - Retrieve a document by ID
- `GET /api/documents` - List documents
- `DELETE /api/documents/{id}` - Delete a document
- `POST /api/search` - Search for similar documents
- `POST /api/query` - Query with RAG

## Example Usage

### Store a Document

```bash
curl -X POST http://localhost:8080/api/documents \
  -H "Content-Type: application/json" \
  -d '{"content":"Go is an open source programming language that makes it simple to build secure, scalable systems."}'
```

### Query with RAG

```bash
curl -X POST http://localhost:8080/api/query \
  -H "Content-Type: application/json" \
  -d '{"query":"What makes Go good for scalable systems?"}'
```

## Project Structure

- `cmd/api`: Main application entry point
- `cmd/dataloader`: Data loading tool
- `data/samples`: Sample documents for testing
- `internal`: Internal packages
  - `api`: API handlers and server
  - `config`: Application configuration
  - `database`: Database interactions
  - `embeddings`: Embedding generation service
  - `loader`: Document loading and chunking
  - `models`: Data models
  - `service`: RAG service implementation
- `deployments`: Deployment configurations
  - `docker-compose.yml`: Docker Compose configuration
  - `Dockerfile`: Docker build configuration
  - `init-scripts`: Database initialization scripts

## License

This project is open source and available under the MIT License.