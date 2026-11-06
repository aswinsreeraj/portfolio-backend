package ingestion

import (
	"strings"

	"github.com/google/uuid"
	"github.com/aswin/portfolio-rag/internal/domain"
)

// DocumentChunker decomposes markdown documents into semantically coherent segments
type DocumentChunker struct {
	maxChunkLength int
}

func NewDocumentChunker(maxChunkLength int) *DocumentChunker {
	if maxChunkLength <= 0 {
		maxChunkLength = 1200 // characters per chunk (~250-300 tokens)
	}
	return &DocumentChunker{maxChunkLength: maxChunkLength}
}

// ChunkDocument parses a markdown file and returns structured chunks tagged with headings
func (c *DocumentChunker) ChunkDocument(docID, title string, content string) []domain.Chunk {
	lines := strings.Split(content, "\n")
	var chunks []domain.Chunk

	currentHeading := title
	var currentLines []string
	chunkIdx := 0

	flushChunk := func() {
		text := strings.TrimSpace(strings.Join(currentLines, "\n"))
		if len(text) > 0 {
			chunks = append(chunks, domain.Chunk{
				ID:         uuid.New(),
				DocumentID: docID,
				ChunkIndex: chunkIdx,
				Heading:    currentHeading,
				Content:    text,
			})
			chunkIdx++
		}
		currentLines = nil
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect markdown headings: #, ##, ###
		if strings.HasPrefix(trimmed, "# ") || strings.HasPrefix(trimmed, "## ") || strings.HasPrefix(trimmed, "### ") {
			if len(currentLines) > 0 {
				flushChunk()
			}
			currentHeading = strings.TrimLeft(trimmed, "# ")
			continue
		}

		currentLines = append(currentLines, line)

		// Check if chunk has accumulated enough characters and hits a paragraph boundary
		curLen := 0
		for _, l := range currentLines {
			curLen += len(l) + 1
		}

		if curLen >= c.maxChunkLength && trimmed == "" {
			flushChunk()
		}
	}

	flushChunk()
	return chunks
}
