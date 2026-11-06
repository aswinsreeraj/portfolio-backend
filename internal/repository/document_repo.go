package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aswin/portfolio-rag/internal/domain"
)

type DocumentRepository struct {
	db *sql.DB
}

func NewDocumentRepository(db *sql.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

// GetAllHashes returns map of docID -> SHA256 to determine delta changes
func (r *DocumentRepository) GetAllHashes(ctx context.Context) (map[string]string, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, sha256 FROM knowledge_documents")
	if err != nil {
		return nil, fmt.Errorf("failed to query document hashes: %w", err)
	}
	defer rows.Close()

	hashes := make(map[string]string)
	for rows.Next() {
		var id, sha string
		if err := rows.Scan(&id, &sha); err != nil {
			return nil, err
		}
		hashes[id] = sha
	}
	return hashes, nil
}

// UpsertDocument stores or updates document metadata
func (r *DocumentRepository) UpsertDocument(ctx context.Context, doc domain.Document) error {
	query := `
		INSERT INTO knowledge_documents (id, title, source_type, url, sha256, last_updated)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			source_type = EXCLUDED.source_type,
			url = EXCLUDED.url,
			sha256 = EXCLUDED.sha256,
			last_updated = EXCLUDED.last_updated
	`
	_, err := r.db.ExecContext(ctx, query, doc.ID, doc.Title, doc.SourceType, doc.URL, doc.SHA256, time.Now())
	if err != nil {
		return fmt.Errorf("failed to upsert document %s: %w", doc.ID, err)
	}
	return nil
}
