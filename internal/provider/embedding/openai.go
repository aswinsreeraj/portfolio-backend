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

type OpenAIEmbeddingProvider struct {
	apiKey string
	model  string
	client *http.Client
}

func NewOpenAIEmbeddingProvider(apiKey, model string) *OpenAIEmbeddingProvider {
	if model == "" {
		model = "text-embedding-3-small"
	}
	return &OpenAIEmbeddingProvider{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (p *OpenAIEmbeddingProvider) Dimension() int {
	return 768
}

type openAIEmbedRequest struct {
	Input      any    `json:"input"`
	Model      string `json:"model"`
	Dimensions int    `json:"dimensions"`
}

type openAIEmbedResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *OpenAIEmbeddingProvider) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	reqBody := openAIEmbedRequest{
		Input:      text,
		Model:      p.model,
		Dimensions: 768,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal openai embed request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create openai request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai embed request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read openai response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai api returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var res openAIEmbedResponse
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, fmt.Errorf("failed to unmarshal openai response: %w", err)
	}

	if res.Error != nil {
		return nil, fmt.Errorf("openai error: %s", res.Error.Message)
	}

	if len(res.Data) == 0 {
		return nil, fmt.Errorf("openai returned empty embedding data")
	}

	return res.Data[0].Embedding, nil
}

func (p *OpenAIEmbeddingProvider) BatchGenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	reqBody := openAIEmbedRequest{
		Input:      texts,
		Model:      p.model,
		Dimensions: 768,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal openai batch embed request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create openai batch request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai batch embed request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read openai batch response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai batch api returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var res openAIEmbedResponse
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, fmt.Errorf("failed to unmarshal openai response: %w", err)
	}

	if res.Error != nil {
		return nil, fmt.Errorf("openai error: %s", res.Error.Message)
	}

	results := make([][]float32, len(texts))
	for _, item := range res.Data {
		if item.Index < len(results) {
			results[item.Index] = item.Embedding
		}
	}

	return results, nil
}

