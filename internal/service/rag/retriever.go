package rag

import (
	"context"
	"fmt"
	"strings"

	"github.com/aswin/portfolio-rag/internal/domain"
	"github.com/aswin/portfolio-rag/internal/repository"
)

type RetrieverService struct {
	vectorRepo *repository.VectorRepository
	embedder   domain.EmbeddingProvider
}

func NewRetrieverService(vectorRepo *repository.VectorRepository, embedder domain.EmbeddingProvider) *RetrieverService {
	return &RetrieverService{
		vectorRepo: vectorRepo,
		embedder:   embedder,
	}
}

type RetrievalResult struct {
	ContextText string
	Citations   []domain.SourceCitation
	Chunks      []domain.Chunk
}

// RetrieveRelevantContext embeds the query and fetches the top matching portfolio chunks
func (s *RetrieverService) RetrieveRelevantContext(ctx context.Context, query string, topK int) (*RetrievalResult, error) {
	if topK <= 0 {
		topK = 5
	}

	queryEmbedding, err := s.embedder.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed search query: %w", err)
	}

	chunks, citations, err := s.vectorRepo.SearchSimilarChunks(ctx, queryEmbedding, topK)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	if len(chunks) == 0 {
		return &RetrievalResult{
			ContextText: "",
			Citations:   nil,
			Chunks:      nil,
		}, nil
	}

	var sb strings.Builder
	for i, chunk := range chunks {
		sb.WriteString(fmt.Sprintf("### Chunk from: %s (%s)\n", citations[i].Title, citations[i].SourceType))
		sb.WriteString(fmt.Sprintf("Heading: %s\n", chunk.Heading))
		sb.WriteString(chunk.Content)
		sb.WriteString("\n\n---\n\n")
	}

	return &RetrievalResult{
		ContextText: strings.TrimSpace(sb.String()),
		Citations:   citations,
		Chunks:      chunks,
	}, nil
}
