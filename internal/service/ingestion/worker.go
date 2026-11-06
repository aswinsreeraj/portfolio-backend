package ingestion

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	pgvector "github.com/pgvector/pgvector-go"
	"github.com/aswin/portfolio-rag/internal/domain"
	"github.com/aswin/portfolio-rag/internal/repository"
)

type IngestionWorker struct {
	knowledgeDir string
	docRepo      *repository.DocumentRepository
	vectorRepo   *repository.VectorRepository
	chunker      *DocumentChunker
	embedder     domain.EmbeddingProvider
}

func NewIngestionWorker(
	knowledgeDir string,
	docRepo *repository.DocumentRepository,
	vectorRepo *repository.VectorRepository,
	embedder domain.EmbeddingProvider,
) *IngestionWorker {
	return &IngestionWorker{
		knowledgeDir: knowledgeDir,
		docRepo:      docRepo,
		vectorRepo:   vectorRepo,
		chunker:      NewDocumentChunker(1200),
		embedder:     embedder,
	}
}

// SyncKnowledge scans knowledgeDir, detects new or modified documents via SHA256,
// generates embeddings, and upserts them into Postgres.
func (w *IngestionWorker) SyncKnowledge(ctx context.Context) error {
	slog.Info("Starting knowledge base synchronization...", "dir", w.knowledgeDir)

	existingHashes, err := w.docRepo.GetAllHashes(ctx)
	if err != nil {
		return fmt.Errorf("failed to load existing hashes: %w", err)
	}

	indexedCount := 0
	skippedCount := 0

	err = filepath.WalkDir(w.knowledgeDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		relPath, err := filepath.Rel(w.knowledgeDir, path)
		if err != nil {
			relPath = path
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		sum := fmt.Sprintf("%x", sha256.Sum256(data))
		if existingHash, ok := existingHashes[relPath]; ok && existingHash == sum {
			skippedCount++
			return nil // Up-to-date, skip re-indexing
		}

		content := string(data)
		title, sourceType, url := extractDocMetadata(relPath, content)

		chunks := w.chunker.ChunkDocument(relPath, title, content)
		if len(chunks) == 0 {
			return nil
		}

		// Extract chunk texts for batch embedding
		texts := make([]string, len(chunks))
		for i, c := range chunks {
			texts[i] = fmt.Sprintf("%s\n%s", c.Heading, c.Content)
		}

		embeddings, err := w.embedder.BatchGenerateEmbeddings(ctx, texts)
		if err != nil {
			return fmt.Errorf("failed to generate embeddings for %s: %w", relPath, err)
		}

		for i := range chunks {
			chunks[i].Embedding = pgvector.NewVector(embeddings[i])
		}

		doc := domain.Document{
			ID:         relPath,
			Title:      title,
			SourceType: sourceType,
			URL:        url,
			SHA256:     sum,
		}

		if err := w.docRepo.UpsertDocument(ctx, doc); err != nil {
			return fmt.Errorf("failed to upsert document %s: %w", relPath, err)
		}

		if err := w.vectorRepo.ReplaceChunksForDocument(ctx, relPath, chunks); err != nil {
			return fmt.Errorf("failed to replace chunks for %s: %w", relPath, err)
		}

		indexedCount++
		slog.Info("Indexed knowledge document", "id", relPath, "chunks", len(chunks))
		return nil
	})

	if err != nil {
		return fmt.Errorf("knowledge sync error: %w", err)
	}

	slog.Info("Knowledge sync completed", "indexed", indexedCount, "skipped_unchanged", skippedCount)
	return nil
}

func extractDocMetadata(relPath, content string) (title, sourceType, url string) {
	lines := strings.Split(content, "\n")
	title = relPath
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, "# ") {
			title = strings.TrimPrefix(trimmed, "# ")
			break
		}
	}

	// Classify source type based on folder
	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) > 1 {
		sourceType = parts[0]
	} else {
		sourceType = "general"
	}

	// Try extracting URL
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.Contains(trimmed, "http://") || strings.Contains(trimmed, "https://") {
			tokens := strings.Fields(trimmed)
			for _, tok := range tokens {
				cleanTok := strings.Trim(tok, "()<>*[]`")
				if strings.HasPrefix(cleanTok, "https://") || strings.HasPrefix(cleanTok, "http://") {
					url = cleanTok
					break
				}
			}
			if url != "" {
				break
			}
		}
	}

	return title, sourceType, url
}
