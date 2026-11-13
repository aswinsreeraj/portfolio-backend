package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/aswin/portfolio-rag/internal/domain"
)

type GeminiLLMProvider struct {
	apiKey   string
	model    string
	client   *http.Client
	fallback *FallbackLLMProvider
}

func NewGeminiLLMProvider(apiKey, model string) *GeminiLLMProvider {
	if model == "" {
		model = "gemini-3.5-flash-lite"
	}
	return &GeminiLLMProvider{
		apiKey:   apiKey,
		model:    model,
		client:   &http.Client{Timeout: 30 * time.Second},
		fallback: NewFallbackLLMProvider(),
	}
}

func (p *GeminiLLMProvider) Name() string {
	return "gemini"
}

type geminiPartItem struct {
	Text string `json:"text"`
}

type geminiTurn struct {
	Role  string           `json:"role"`
	Parts []geminiPartItem `json:"parts"`
}

type geminiGenRequest struct {
	SystemInstruction *struct {
		Parts []geminiPartItem `json:"parts"`
	} `json:"system_instruction,omitempty"`
	Contents         []geminiTurn `json:"contents"`
	GenerationConfig *struct {
		Temperature     float32 `json:"temperature"`
		MaxOutputTokens int     `json:"maxOutputTokens"`
	} `json:"generationConfig,omitempty"`
}

type geminiGenResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *GeminiLLMProvider) GenerateResponse(ctx context.Context, systemPrompt string, contextText string, messages []domain.ChatMessage) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", p.model, p.apiKey)

	fullSystemPrompt := systemPrompt
	if contextText != "" {
		fullSystemPrompt += fmt.Sprintf("\n\n<portfolio_context>\n%s\n</portfolio_context>\n", contextText)
	}

	reqBody := geminiGenRequest{
		SystemInstruction: &struct {
			Parts []geminiPartItem `json:"parts"`
		}{
			Parts: []geminiPartItem{{Text: fullSystemPrompt}},
		},
		GenerationConfig: &struct {
			Temperature     float32 `json:"temperature"`
			MaxOutputTokens int     `json:"maxOutputTokens"`
		}{
			Temperature:     0.2, // low temperature for grounded facts
			MaxOutputTokens: 250, // Capped to guarantee answers remain under 100 words
		},
	}

	for _, msg := range messages {
		role := "user"
		if msg.Role == "assistant" {
			role = "model"
		}
		reqBody.Contents = append(reqBody.Contents, geminiTurn{
			Role:  role,
			Parts: []geminiPartItem{{Text: msg.Content}},
		})
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal gemini request: %w", err)
	}

	// Retry loop for transient 503 / 429
	var bodyBytes []byte
	var statusCode int

	for attempt := 0; attempt < 2; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
		if err != nil {
			return "", fmt.Errorf("failed to create gemini request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := p.client.Do(req)
		if err != nil {
			slog.Warn("Gemini API call failed, retrying...", "attempt", attempt+1, "error", err)
			time.Sleep(1 * time.Second)
			continue
		}

		bodyBytes, err = io.ReadAll(resp.Body)
		statusCode = resp.StatusCode
		resp.Body.Close()

		if statusCode == http.StatusOK {
			break
		}

		slog.Warn("Gemini returned non-200 status", "status", statusCode, "attempt", attempt+1)
		time.Sleep(1 * time.Second)
	}

	if statusCode != http.StatusOK {
		slog.Warn("Gemini API temporarily unavailable, falling back to grounded responder", "status", statusCode)
		return p.fallback.GenerateResponse(ctx, systemPrompt, contextText, messages)
	}

	var res geminiGenResponse
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return "", fmt.Errorf("failed to unmarshal gemini response: %w", err)
	}

	if res.Error != nil {
		slog.Warn("Gemini API returned error body, using fallback", "error", res.Error.Message)
		return p.fallback.GenerateResponse(ctx, systemPrompt, contextText, messages)
	}

	if len(res.Candidates) == 0 || len(res.Candidates[0].Content.Parts) == 0 {
		return p.fallback.GenerateResponse(ctx, systemPrompt, contextText, messages)
	}

	return res.Candidates[0].Content.Parts[0].Text, nil
}
