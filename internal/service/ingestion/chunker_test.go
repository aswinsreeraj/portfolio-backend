package ingestion

import (
	"testing"
)

func TestChunkDocumentByHeadings(t *testing.T) {
	chunker := NewDocumentChunker(200)

	content := `# Document Title

This is introductory text.

## Section 1: Overview
Here is the first section with details.

## Section 2: Architecture
Here is the second section discussing Go and distributed systems.`

	chunks := chunker.ChunkDocument("test/doc.md", "Document Title", content)

	if len(chunks) < 3 {
		t.Fatalf("expected at least 3 chunks, got %d", len(chunks))
	}

	if chunks[0].DocumentID != "test/doc.md" {
		t.Errorf("expected docID test/doc.md, got %s", chunks[0].DocumentID)
	}

	foundSection2 := false
	for _, c := range chunks {
		if c.Heading == "Section 2: Architecture" {
			foundSection2 = true
		}
	}

	if !foundSection2 {
		t.Errorf("expected to find chunk with heading 'Section 2: Architecture'")
	}
}

