package repository

import (
	"context"
	"database/sql"
	"fmt"

	pgvector "github.com/pgvector/pgvector-go"
	"github.com/aswin/portfolio-rag/internal/domain"
)

type VectorRepository struct {
	db *sql.DB
}

func NewVectorRepository(db *sql.DB) *VectorRepository {
	return &VectorRepository{db: db}
}

// ReplaceChunksForDocument atomically deletes old chunks and inserts new ones with vector embeddings
func (r *VectorRepository) ReplaceChunksForDocument(ctx context.Context, docID string, chunks []domain.Chunk) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Delete existing chunks for this document
	if _, err := tx.ExecContext(ctx, "DELETE FROM knowledge_chunks WHERE document_id = $1", docID); err != nil {
		return fmt.Errorf("failed to delete old chunks for %s: %w", docID, err)
	}

	// 2. Insert new chunks
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO knowledge_chunks (id, document_id, chunk_index, heading, content, embedding)
		VALUES ($1, $2, $3, $4, $5, $6)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	for _, chunk := range chunks {
		_, err := stmt.ExecContext(
			ctx,
			chunk.ID,
			chunk.DocumentID,
			chunk.ChunkIndex,
			chunk.Heading,
			chunk.Content,
			chunk.Embedding,
		)
		if err != nil {
			return fmt.Errorf("failed to insert chunk: %w", err)
		}
	}

	return tx.Commit()
}

// SearchSimilarChunks retrieves the top-K chunks closest to query embedding via cosine similarity
func (r *VectorRepository) SearchSimilarChunks(ctx context.Context, queryEmbedding []float32, limit int) ([]domain.Chunk, []domain.SourceCitation, error) {
	if limit <= 0 {
		limit = 5
	}

	vec := pgvector.NewVector(queryEmbedding)

	query := `
		SELECT c.id, c.document_id, c.chunk_index, c.heading, c.content,
		       d.title, d.source_type, COALESCE(d.url, ''),
		       1 - (c.embedding <=> $1) AS score
		FROM knowledge_chunks c
		JOIN knowledge_documents d ON c.document_id = d.id
		ORDER BY c.embedding <=> $1 ASC
		LIMIT $2;
	`

	rows, err := r.db.QueryContext(ctx, query, vec, limit)
	if err != nil {
		return nil, nil, fmt.Errorf("similarity search failed: %w", err)
	}
	defer rows.Close()

	var chunks []domain.Chunk
	var citations []domain.SourceCitation

	for rows.Next() {
		var chunk domain.Chunk
		var title, sourceType, url string
		var score float32

		if err := rows.Scan(
			&chunk.ID,
			&chunk.DocumentID,
			&chunk.ChunkIndex,
			&chunk.Heading,
			&chunk.Content,
			&title,
			&sourceType,
			&url,
			&score,
		); err != nil {
			return nil, nil, fmt.Errorf("failed to scan chunk row: %w", err)
		}

		chunks = append(chunks, chunk)
		citations = append(citations, domain.SourceCitation{
			Title:      title,
			SourceType: sourceType,
			Heading:    chunk.Heading,
			URL:        url,
			Score:      score,
		})
	}

	return chunks, citations, nil
}
