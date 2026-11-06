package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/aswin/portfolio-rag/internal/domain"
)

// FallbackLLMProvider synthesizes clear, direct answers from retrieved context chunks.
// Used when external LLM credentials are not configured or as an instant offline fallback.
type FallbackLLMProvider struct{}

func NewFallbackLLMProvider() *FallbackLLMProvider {
	return &FallbackLLMProvider{}
}

func (p *FallbackLLMProvider) Name() string {
	return "deterministic-fallback"
}

func (p *FallbackLLMProvider) GenerateResponse(ctx context.Context, systemPrompt string, contextText string, messages []domain.ChatMessage) (string, error) {
	if len(messages) == 0 {
		return "Hi, I'm Aswin! Feel free to ask me anything about my projects (ARDA, EVNTX, Tenome), my Go backend experience, or my engineering background.", nil
	}

	lastUserMsg := ""
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			lastUserMsg = strings.TrimSpace(messages[i].Content)
			break
		}
	}
	_ = lastUserMsg

	if contextText == "" {
		return "I haven't documented that topic in my current portfolio projects. Feel free to ask about my work on ARDA, EVNTX, Tenome, distributed systems, or Go concurrency!", nil
	}

	// Clean up markdown headers and build an informative synthesized summary
	lines := strings.Split(contextText, "\n")
	var keyPoints []string
	currentTopic := ""

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "### Chunk from:") {
			currentTopic = strings.TrimPrefix(trimmed, "### Chunk from: ")
			continue
		}
		if strings.HasPrefix(trimmed, "# ") || strings.HasPrefix(trimmed, "## ") || strings.HasPrefix(trimmed, "### ") {
			continue
		}
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			keyPoints = append(keyPoints, strings.TrimPrefix(strings.TrimPrefix(trimmed, "- "), "* "))
		} else if len(trimmed) > 40 && len(keyPoints) < 6 {
			keyPoints = append(keyPoints, trimmed)
		}
	}

	if len(keyPoints) == 0 {
		return fmt.Sprintf("Here is what I've built and documented regarding %s:\n\n%s", currentTopic, contextText), nil
	}

	// Limit to top 5 most informative sentences / bullets
	limit := 5
	if len(keyPoints) < limit {
		limit = len(keyPoints)
	}

	var sb strings.Builder
	sb.WriteString("Here is what I've engineered and documented on this:\n\n")
	for i := 0; i < limit; i++ {
		sb.WriteString("• " + keyPoints[i] + "\n\n")
	}

	sb.WriteString(fmt.Sprintf("Feel free to ask for more details on any specific architectural decision, benchmarks, or codebase patterns!"))
	return strings.TrimSpace(sb.String()), nil
}
