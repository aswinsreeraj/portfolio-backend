package chat

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/aswin/portfolio-rag/internal/domain"
	"github.com/aswin/portfolio-rag/internal/service/rag"
)

type ChatService struct {
	retriever   *rag.RetrieverService
	llmProvider domain.LLMProvider
}

func NewChatService(retriever *rag.RetrieverService, llmProvider domain.LLMProvider) *ChatService {
	return &ChatService{
		retriever:   retriever,
		llmProvider: llmProvider,
	}
}

func (s *ChatService) ProcessChat(ctx context.Context, req domain.ChatRequest) (*domain.ChatResponse, error) {
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("chat messages slice cannot be empty")
	}

	convID := req.ConversationID
	if convID == "" {
		convID = uuid.New().String()
	}

	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode == "" {
		mode = "ask-about-me"
	}

	// 1. Identify last user question to run RAG query
	lastUserQuestion := ""
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			lastUserQuestion = strings.TrimSpace(req.Messages[i].Content)
			break
		}
	}

	if lastUserQuestion == "" {
		return &domain.ChatResponse{
			Response:       "Please ask a question regarding Aswin's experience, projects, or skills.",
			Sources:        nil,
			ConversationID: convID,
			Mode:           mode,
		}, nil
	}

	// 2. Retrieve relevant portfolio chunks
	retrievalResult, err := s.retriever.RetrieveRelevantContext(ctx, lastUserQuestion, 4)
	if err != nil {
		return nil, fmt.Errorf("knowledge retrieval failed: %w", err)
	}

	// 3. Select mode system prompt
	systemPrompt := GetSystemPromptForMode(mode)

	// 4. Generate answer with configured LLM provider
	answer, err := s.llmProvider.GenerateResponse(ctx, systemPrompt, retrievalResult.ContextText, req.Messages)
	if err != nil {
		return nil, fmt.Errorf("response generation failed: %w", err)
	}

	// Guarantee response is strictly 100 words or fewer
	answer = trimToWordLimit(answer, 100)

	// 5. Build unique citation list
	seenDocs := make(map[string]bool)
	var filteredCitations []domain.SourceCitation
	for _, c := range retrievalResult.Citations {
		key := fmt.Sprintf("%s-%s", c.SourceType, c.Title)
		if !seenDocs[key] && len(filteredCitations) < 3 {
			seenDocs[key] = true
			filteredCitations = append(filteredCitations, c)
		}
	}

	return &domain.ChatResponse{
		Response:       answer,
		Sources:        filteredCitations,
		ConversationID: convID,
		Mode:           mode,
	}, nil
}

// trimToWordLimit ensures that response answers strictly obey the 100-word ceiling
// by trimming cleanly at the nearest natural sentence boundary.
func trimToWordLimit(text string, maxWords int) string {
	words := strings.Fields(text)
	if len(words) <= maxWords {
		return text
	}

	truncated := strings.Join(words[:maxWords], " ")
	lastPunct := -1
	for i := len(truncated) - 1; i >= int(float64(len(truncated))*0.7); i-- {
		ch := truncated[i]
		if (ch == '.' || ch == '!' || ch == '?') && (i == len(truncated)-1 || truncated[i+1] == ' ' || truncated[i+1] == '\n') {
			lastPunct = i
			break
		}
	}

	if lastPunct != -1 {
		return truncated[:lastPunct+1]
	}
	return truncated + "..."
}

