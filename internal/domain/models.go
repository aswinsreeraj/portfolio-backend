package domain

import (
	"time"

	"github.com/google/uuid"
	pgvector "github.com/pgvector/pgvector-go"
)

// Document represents a top-level knowledge file indexed in the vector database
type Document struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	SourceType  string    `json:"source_type"`
	URL         string    `json:"url"`
	SHA256      string    `json:"sha256"`
	LastUpdated time.Time `json:"last_updated"`
}

// Chunk represents a semantically segmented portion of a document with its vector embedding
type Chunk struct {
	ID         uuid.UUID       `json:"id"`
	DocumentID string          `json:"document_id"`
	ChunkIndex int             `json:"chunk_index"`
	Heading    string          `json:"heading"`
	Content    string          `json:"content"`
	Embedding  pgvector.Vector `json:"-"`
	CreatedAt  time.Time       `json:"created_at"`
}

// SourceCitation represents metadata returned to the frontend for citation badges
type SourceCitation struct {
	Title      string  `json:"title"`
	SourceType string  `json:"source_type"`
	Heading    string  `json:"heading"`
	URL        string  `json:"url,omitempty"`
	Score      float32 `json:"score"`
}

// ChatMessage represents a single turn in a chat interaction
type ChatMessage struct {
	Role    string `json:"role"` // "user", "assistant", "system"
	Content string `json:"content"`
}

// ChatRequest is the payload sent from the frontend chat interface
type ChatRequest struct {
	Messages       []ChatMessage `json:"messages" binding:"required,min=1"`
	Mode           string        `json:"mode,omitempty"` // ask-about-me, explore-projects, technical-deep-dive, interview-mode, recruiter-mode
	ConversationID string        `json:"conversation_id,omitempty"`
}

// ChatResponse is the payload returned to the frontend with answers and citations
type ChatResponse struct {
	Response       string           `json:"response"`
	Sources        []SourceCitation `json:"sources"`
	ConversationID string           `json:"conversation_id"`
	Mode           string           `json:"mode"`
}

