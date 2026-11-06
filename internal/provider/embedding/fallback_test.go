package embedding

import (
	"context"
	"math"
	"testing"
)

func TestFallbackEmbeddingDimensionAndNormalization(t *testing.T) {
	p := NewFallbackEmbeddingProvider()
	ctx := context.Background()

	vec, err := p.GenerateEmbedding(ctx, "Distributed systems in Go with Apache Kafka and PostgreSQL")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(vec) != 768 {
		t.Fatalf("expected dimension 768, got %d", len(vec))
	}

	// Calculate L2 norm
	var sumSq float64
	for _, v := range vec {
		sumSq += float64(v * v)
	}
	norm := math.Sqrt(sumSq)

	if math.Abs(norm-1.0) > 1e-4 {
		t.Fatalf("expected vector to be unit-normalized (~1.0), got %f", norm)
	}
}

func TestFallbackEmbeddingBatch(t *testing.T) {
	p := NewFallbackEmbeddingProvider()
	ctx := context.Background()

	texts := []string{"First chunk", "Second chunk", "Third chunk"}
	vectors, err := p.BatchGenerateEmbeddings(ctx, texts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(vectors) != 3 {
		t.Fatalf("expected 3 vectors, got %d", len(vectors))
	}
	for i, v := range vectors {
		if len(v) != 768 {
			t.Errorf("vector %d has length %d, expected 768", i, len(v))
		}
	}
}

