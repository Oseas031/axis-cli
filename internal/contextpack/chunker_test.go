package contextpack

import (
	"strings"
	"testing"
)

func TestChunker_Chunk(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		maxChunk int
		minSize  int
		expected int
	}{
		{"small document", "hello world", 1000, 100, 1},
		{"large document", strings.Repeat("paragraph\n\n", 100), 500, 100, 3},
		{"empty document", "", 1000, 100, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunker := NewChunker(tt.maxChunk, 0, tt.minSize)
			doc := &Document{ID: "test", Content: tt.content}
			chunks := chunker.Chunk(doc)

			if len(chunks) != tt.expected {
				t.Errorf("Chunk() = %d chunks, want %d", len(chunks), tt.expected)
			}
		})
	}
}

func TestSplitByParagraph(t *testing.T) {
	text := "paragraph 1\n\nparagraph 2\n\nparagraph 3"
	paragraphs := splitByParagraph(text)

	if len(paragraphs) != 3 {
		t.Errorf("splitByParagraph() = %d paragraphs, want 3", len(paragraphs))
	}
}

func TestSplitBySentence(t *testing.T) {
	text := "sentence 1. sentence 2! sentence 3?"
	sentences := splitBySentence(text)

	if len(sentences) != 3 {
		t.Errorf("splitBySentence() = %d sentences, want 3", len(sentences))
	}
}

func TestChunker_NilDocument(t *testing.T) {
	chunker := NewChunker(1000, 0, 100)
	chunks := chunker.Chunk(nil)

	if chunks != nil {
		t.Errorf("Chunk(nil) = %v, want nil", chunks)
	}
}

func TestChunker_MinChunkSize(t *testing.T) {
	chunker := NewChunker(100, 0, 50)
	doc := &Document{ID: "test", Content: "short"}
	chunks := chunker.Chunk(doc)

	if len(chunks) != 1 {
		t.Errorf("Chunk() with content < MinChunkSize = %d chunks, want 1", len(chunks))
	}
}

func TestChunker_Overlap(t *testing.T) {
	chunker := NewChunker(100, 20, 10)
	content := strings.Repeat("abcdefghij\n\n", 20)
	doc := &Document{ID: "test", Content: content}
	chunks := chunker.ChunkWithOverlap(doc)

	if len(chunks) < 2 {
		t.Errorf("ChunkWithOverlap() = %d chunks, want >= 2", len(chunks))
	}

	for i := 1; i < len(chunks); i++ {
		if chunks[i].Position < chunks[i-1].Position {
			t.Errorf("Chunk[%d].Position (%d) < Chunk[%d].Position (%d)", i, chunks[i].Position, i-1, chunks[i-1].Position)
		}
	}
}
