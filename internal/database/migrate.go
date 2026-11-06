package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

const initSchemaSQL = `
CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS knowledge_documents (
    id VARCHAR(128) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    source_type VARCHAR(64) NOT NULL,
    url VARCHAR(512),
    sha256 VARCHAR(64) NOT NULL,
    last_updated TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS knowledge_chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id VARCHAR(128) NOT NULL REFERENCES knowledge_documents(id) ON DELETE CASCADE,
    chunk_index INT NOT NULL,
    heading VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    embedding vector(768),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_knowledge_chunks_doc_id ON knowledge_chunks(document_id);

CREATE INDEX IF NOT EXISTS idx_knowledge_chunks_embedding 
    ON knowledge_chunks USING hnsw (embedding vector_cosine_ops);
`

// RunMigrations executes baseline schema migrations
func RunMigrations(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	slog.Info("Running database migrations...")

	_, err := db.ExecContext(ctx, initSchemaSQL)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	slog.Info("Database migrations applied successfully")
	return nil
}
