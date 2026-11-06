package embedding

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"math"
	"strings"
)

// FallbackEmbeddingProvider generates deterministic normalized 768-dimensional vectors
// using token n-gram feature hashing. This guarantees vector similarity works out-of-the-box
// in offline and local testing environments.
type FallbackEmbeddingProvider struct {
	dim int
}

func NewFallbackEmbeddingProvider() *FallbackEmbeddingProvider {
	return &FallbackEmbeddingProvider{dim: 768}
}

func (p *FallbackEmbeddingProvider) Dimension() int {
	return p.dim
}

func (p *FallbackEmbeddingProvider) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	vec := make([]float32, p.dim)
	normalized := strings.ToLower(strings.TrimSpace(text))
	tokens := strings.Fields(normalized)

	if len(tokens) == 0 {
		return vec, nil
	}

	// 1. Unigrams and bigrams hashing
	for i, token := range tokens {
		hashToken(token, 1.0, vec, p.dim)
		if i+1 < len(tokens) {
			bigram := token + " " + tokens[i+1]
			hashToken(bigram, 1.5, vec, p.dim)
		}
	}

	// 2. L2 normalize vector so cosine distance works properly
	var sumSq float64
	for _, v := range vec {
		sumSq += float64(v * v)
	}
	if sumSq > 0 {
		norm := float32(1.0 / math.Sqrt(sumSq))
		for i := range vec {
			vec[i] *= norm
		}
	}

	return vec, nil
}

func (p *FallbackEmbeddingProvider) BatchGenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	result := make([][]float32, len(texts))
	for i, t := range texts {
		emb, err := p.GenerateEmbedding(ctx, t)
		if err != nil {
			return nil, err
		}
		result[i] = emb
	}
	return result, nil
}

func hashToken(token string, weight float32, vec []float32, dim int) {
	h := sha256.Sum256([]byte(token))
	idx1 := binary.BigEndian.Uint32(h[0:4]) % uint32(dim)
	idx2 := binary.BigEndian.Uint32(h[4:8]) % uint32(dim)
	sign1 := float32(1)
	if h[8]%2 == 0 {
		sign1 = -1
	}
	sign2 := float32(1)
	if h[9]%2 == 0 {
		sign2 = -1
	}

	vec[idx1] += sign1 * weight
	vec[idx2] += sign2 * (weight * 0.5)
}

