package contextpack

import "strings"

// Chunker 文档分块器
type Chunker struct {
	MaxChunkSize int // 最大分块大小（字符数）
	OverlapSize  int // 重叠大小
	MinChunkSize int // 最小分块大小
}

// Chunk 分块结果
type Chunk struct {
	Content    string // 分块内容
	Position   int    // 在原文档中的位置
	DocumentID string // 源文档ID
	Index      int    // 分块索引
}

// Document 文档结构
type Document struct {
	ID      string
	Content string
	Path    string
}

// NewChunker 创建分块器
func NewChunker(maxSize, overlap, minSize int) *Chunker {
	if maxSize <= 0 {
		maxSize = 1000
	}
	if overlap < 0 {
		overlap = 0
	}
	if minSize <= 0 {
		minSize = 100
	}
	if overlap >= maxSize {
		overlap = maxSize / 2
	}
	return &Chunker{
		MaxChunkSize: maxSize,
		OverlapSize:  overlap,
		MinChunkSize: minSize,
	}
}

// Chunk 文档分块
func (c *Chunker) Chunk(document *Document) []*Chunk {
	if document == nil || len(document.Content) == 0 {
		return nil
	}

	content := document.Content

	if len(content) <= c.MaxChunkSize {
		return []*Chunk{
			{
				Content:    content,
				Position:   0,
				DocumentID: document.ID,
				Index:      0,
			},
		}
	}

	var chunks []*Chunk
	paragraphs := splitByParagraph(content)

	chunkIndex := 0
	currentChunk := ""
	currentPosition := 0

	for _, para := range paragraphs {
		if len(para) > c.MaxChunkSize {
			if len(currentChunk) >= c.MinChunkSize {
				chunks = append(chunks, &Chunk{
					Content:    currentChunk,
					Position:   currentPosition,
					DocumentID: document.ID,
					Index:      chunkIndex,
				})
				chunkIndex++
				currentPosition += len(currentChunk)
				currentChunk = ""
			}

			sentences := splitBySentence(para)
			for _, sent := range sentences {
				if len(currentChunk)+len(sent)+1 > c.MaxChunkSize {
					if len(currentChunk) >= c.MinChunkSize {
						chunks = append(chunks, &Chunk{
							Content:    currentChunk,
							Position:   currentPosition,
							DocumentID: document.ID,
							Index:      chunkIndex,
						})
						chunkIndex++
						currentPosition += len(currentChunk)
						currentChunk = ""
					}
				}
				currentChunk += sent + " "
			}
			continue
		}

		if len(currentChunk)+len(para)+2 > c.MaxChunkSize {
			if len(currentChunk) >= c.MinChunkSize {
				chunks = append(chunks, &Chunk{
					Content:    currentChunk,
					Position:   currentPosition,
					DocumentID: document.ID,
					Index:      chunkIndex,
				})
				chunkIndex++
				currentPosition += len(currentChunk)
				currentChunk = ""
			}
		}

		currentChunk += para + "\n\n"
	}

	if len(currentChunk) >= c.MinChunkSize {
		chunks = append(chunks, &Chunk{
			Content:    currentChunk,
			Position:   currentPosition,
			DocumentID: document.ID,
			Index:      chunkIndex,
		})
	}

	return chunks
}

// ChunkWithOverlap 带重叠的分块
func (c *Chunker) ChunkWithOverlap(document *Document) []*Chunk {
	chunks := c.Chunk(document)

	if c.OverlapSize <= 0 || len(chunks) <= 1 {
		return chunks
	}

	var result []*Chunk
	for i, chunk := range chunks {
		if i == 0 {
			result = append(result, chunk)
			continue
		}

		prevChunk := chunks[i-1]
		overlap := ""
		if len(prevChunk.Content) > c.OverlapSize {
			overlap = prevChunk.Content[len(prevChunk.Content)-c.OverlapSize:]
		} else {
			overlap = prevChunk.Content
		}

		result = append(result, &Chunk{
			Content:    overlap + chunk.Content,
			Position:   chunk.Position - len(overlap),
			DocumentID: chunk.DocumentID,
			Index:      chunk.Index,
		})
	}

	return result
}

// splitByParagraph 按段落分割
func splitByParagraph(text string) []string {
	parts := strings.Split(text, "\n\n")
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// splitBySentence 按句子分割
func splitBySentence(text string) []string {
	var sentences []string
	var current strings.Builder

	for _, r := range text {
		current.WriteRune(r)

		if r == '.' || r == '!' || r == '?' || r == '。' || r == '！' || r == '？' {
			if current.Len() > 0 {
				sentences = append(sentences, strings.TrimSpace(current.String()))
				current.Reset()
			}
		}
	}

	if current.Len() > 0 {
		sentences = append(sentences, strings.TrimSpace(current.String()))
	}

	return sentences
}
