package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type GeminiEmbeddingProvider struct {
	apiKey string
	model  string
	client *http.Client
}

func NewGeminiEmbeddingProvider(apiKey, model string) *GeminiEmbeddingProvider {
	if model == "" {
		model = "gemini-embedding-001"
	}
	return &GeminiEmbeddingProvider{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 20 * time.Second},
	}
}

func (p *GeminiEmbeddingProvider) Dimension() int {
	return 768
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiEmbedRequest struct {
	Model                string        `json:"model"`
	Content              geminiContent `json:"content"`
	OutputDimensionality int           `json:"outputDimensionality,omitempty"`
}

type geminiEmbedResponse struct {
	Embedding struct {
		Values []float32 `json:"values"`
	} `json:"embedding"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *GeminiEmbeddingProvider) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:embedContent?key=%s", p.model, p.apiKey)

	reqBody := geminiEmbedRequest{
		Model: "models/" + p.model,
		Content: geminiContent{
			Parts: []geminiPart{{Text: text}},
		},
		OutputDimensionality: 768,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal gemini embedding request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read gemini response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini api returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var res geminiEmbedResponse
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, fmt.Errorf("failed to unmarshal gemini response: %w", err)
	}

	if res.Error != nil {
		return nil, fmt.Errorf("gemini error: %s", res.Error.Message)
	}

	return res.Embedding.Values, nil
}

func (p *GeminiEmbeddingProvider) BatchGenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))
	for i, t := range texts {
		emb, err := p.GenerateEmbedding(ctx, t)
		if err != nil {
			return nil, err
		}
		results[i] = emb
	}
	return results, nil
}

