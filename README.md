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
- Comprehensive testing suite
- Flexible deployment options (Docker or standalone)

## Requirements

- Docker and Docker Compose (for containerized deployment)
- Go 1.24 or later (for local development)
- Google Gemini API key
- PostgreSQL with pgvector extension (for local development without Docker)

## Getting Started

### Running with Docker (Recommended)

The easiest way to get started is using Docker, which handles all dependencies including PostgreSQL with pgvector:

1. Clone the repository
   ```bash
   git clone <repository-url>
   cd go-rag
   ```

2. Create a `.env` file in the project root (see [Environment Variables](#environment-variables) section)
   ```bash
   cp .env.dist .env
   # Edit the .env file to add your GEMINI_API_KEY
   ```

3. Start the containers:
   ```bash
   make docker-up
   ```
   This will start both the PostgreSQL database and the RAG service.

4. Check if the containers are running:
   ```bash
   make docker-ps
   ```

5. Load sample data (optional):
   ```bash
   make docker-load-samples
   ```

6. Test the API using curl or any API client:
   ```bash
   curl http://localhost:8080/health
   ```

7. To stop the containers:
   ```bash
   make docker-down
   ```

### Running Without Docker (Local Development)

For local development without Docker:

1. Clone the repository
   ```bash
   git clone <repository-url>
   cd go-rag
   ```

2. Install PostgreSQL with pgvector extension
   - For Ubuntu/Debian: `apt-get install postgresql postgresql-contrib`
   - For macOS: `brew install postgresql`
   - Then install pgvector (see [pgvector installation](https://github.com/pgvector/pgvector#installation))

3. Initialize the database
   - Create a database: `createdb ragdb`
   - Run the initialization script: `psql -d ragdb -f deployments/init-scripts/01-init-pgvector.sql`

4. Create a `.env` file and update database connection details
   ```bash
   cp .env.dist .env
   # Update DB_HOST, DB_USER, DB_PASSWORD and add your GEMINI_API_KEY
   ```

5. Build and run the application
   ```bash
   make build
   make run
   ```

6. Build the data loader and load sample data (optional)
   ```bash
   make build-loader
   make load-samples
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
GEMINI_TEXT_MODEL=gemini-2.5-flash  # Current recommended model for text generation
GEMINI_EMBEDDING_MODEL=embedding-001  # Model for generating vector embeddings

# Vector dimensions for embeddings
EMBEDDING_DIMENSIONS=768
```

## Makefile Commands

The project includes a Makefile with various commands to simplify development. Use `make help` to see all available commands.

### Local Development Commands
```bash
# Build and run
make build         # Build the application
make run           # Run the application locally (requires local PostgreSQL setup)
make clean         # Clean build artifacts
make fmt           # Format Go code
make test          # Run tests
make build-loader  # Build the data loader
make load-samples  # Load sample data locally (requires running database)
```

### Docker Commands
```bash
# Docker operations
make docker-up     # Start all containers in the background
make docker-down   # Stop all containers
make docker-logs   # Start all containers with logs in foreground (useful for debugging)
make docker-rebuild # Rebuild and restart only the app container
make docker-ps     # Show Docker container status
make docker-load-samples # Load sample data in Docker container

# Quick development setup
make dev-setup     # Set up the development environment (builds and starts containers)
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