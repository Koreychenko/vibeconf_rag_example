-- Enable the pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Create a schema for RAG application
CREATE SCHEMA IF NOT EXISTS rag;

-- Create documents table
CREATE TABLE IF NOT EXISTS rag.documents (
    id UUID PRIMARY KEY,
    content TEXT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create embeddings table with vector support
CREATE TABLE IF NOT EXISTS rag.embeddings (
    id UUID PRIMARY KEY,
    document_id UUID NOT NULL REFERENCES rag.documents(id) ON DELETE CASCADE,
    embedding vector(768), -- 768 dimensions for Gemini embeddings
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index for vector similarity search using cosine distance
CREATE INDEX IF NOT EXISTS embeddings_vector_idx ON rag.embeddings USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Create function to search for similar documents
CREATE OR REPLACE FUNCTION rag.search_similar_documents(
    query_embedding vector,
    similarity_threshold FLOAT,
    max_results INT
)
RETURNS TABLE (
    id UUID,
    content TEXT,
    metadata JSONB,
    similarity FLOAT
)
AS $$
BEGIN
    RETURN QUERY
    SELECT
        d.id,
        d.content,
        d.metadata,
        1 - (e.embedding <=> query_embedding) AS similarity
    FROM
        rag.documents d
        JOIN rag.embeddings e ON d.id = e.document_id
    WHERE
        1 - (e.embedding <=> query_embedding) > similarity_threshold
    ORDER BY
        similarity DESC
    LIMIT max_results;
END;
$$ LANGUAGE plpgsql;