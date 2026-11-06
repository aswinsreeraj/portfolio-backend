package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aswin/portfolio-rag/internal/domain"
)

type OpenAILLMProvider struct {
	apiKey string
	model  string
	client *http.Client
}

func NewOpenAILLMProvider(apiKey, model string) *OpenAILLMProvider {
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAILLMProvider{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *OpenAILLMProvider) Name() string {
	return "openai"
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatRequest struct {
	Model       string              `json:"model"`
	Messages    []openAIChatMessage `json:"messages"`
	Temperature float32             `json:"temperature"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message openAIChatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *OpenAILLMProvider) GenerateResponse(ctx context.Context, systemPrompt string, contextText string, messages []domain.ChatMessage) (string, error) {
	fullSystemPrompt := systemPrompt
	if contextText != "" {
		fullSystemPrompt += fmt.Sprintf("\n\n<portfolio_context>\n%s\n</portfolio_context>\n", contextText)
	}

	apiMessages := make([]openAIChatMessage, 0, len(messages)+1)
	apiMessages = append(apiMessages, openAIChatMessage{
		Role:    "system",
		Content: fullSystemPrompt,
	})

	for _, msg := range messages {
		apiMessages = append(apiMessages, openAIChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	reqBody := openAIChatRequest{
		Model:       p.model,
		Messages:    apiMessages,
		Temperature: 0.2,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal openai chat request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create openai request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai chat request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read openai response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai api returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var res openAIChatResponse
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return "", fmt.Errorf("failed to unmarshal openai response: %w", err)
	}

	if res.Error != nil {
		return "", fmt.Errorf("openai api error: %s", res.Error.Message)
	}

	if len(res.Choices) == 0 {
		return "", fmt.Errorf("openai returned no response choices")
	}

	return res.Choices[0].Message.Content, nil
}
