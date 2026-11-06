package domain

import (
	"context"
)

// EmbeddingProvider generates high-dimensional vector embeddings for text chunks
type EmbeddingProvider interface {
	// GenerateEmbedding creates a 768-dimensional float32 vector for a single text input
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
	// BatchGenerateEmbeddings creates vectors for multiple chunks in parallel/batch
	BatchGenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error)
	// Dimension returns the vector dimension (768)
	Dimension() int
}

// LLMProvider generates text completions based on system instructions and portfolio context
type LLMProvider interface {
	// GenerateResponse produces an answer given system prompt, context documents, and message history
	GenerateResponse(ctx context.Context, systemPrompt string, contextText string, messages []ChatMessage) (string, error)
	// Name returns provider identifier (gemini, openai, fallback)
	Name() string
}

