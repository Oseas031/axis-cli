package contextpack

import (
	"unicode/utf8"
)

// semanticBoundary represents a break point within content.
type semanticBoundary struct {
	pos   int // byte position in original string
	score int // higher = coarser / safer boundary
}

// boundaryScore assigns priority to break types.
// Higher score = coarser boundary = preferred truncation point.
const (
	scoreParagraph = 50 // double newline
	scoreCodeBlock = 40 // code block fences or closing braces on their own line
	scoreSentence  = 30 // sentence-ending punctuation
	scoreLine      = 20 // single newline
	scoreHard      = 10 // any other whitespace
	scoreFallback  = 0  // hard cut at maxBytes (last resort)
)

// truncateAtSemanticBoundary returns the largest prefix of content that does
// not exceed maxBytes and ends at a semantic boundary.
// The returned int is the truncation position in the original string (0 if empty).
func truncateAtSemanticBoundary(content string, maxBytes int) (string, int) {
	if maxBytes <= 0 {
		return "", 0
	}
	if len(content) <= maxBytes {
		return content, len(content)
	}

	best := semanticBoundary{pos: 0, score: -1}
	searchEnd := maxBytes

	// Walk forward looking for break points up to maxBytes.
	for i := 0; i < searchEnd; i++ {
		b := content[i]

		// Paragraph boundary: \n\n, \r\n\r\n, \r\r
		if b == '\n' || b == '\r' {
			if i+1 < searchEnd {
				next := content[i+1]
				if next == '\n' || next == '\r' {
					// Skip to the end of the blank line so we don't eat stray whitespace.
					j := i
					for j < searchEnd && (content[j] == '\n' || content[j] == '\r') {
						j++
					}
					if j > best.pos {
						best = semanticBoundary{pos: j, score: scoreParagraph}
					}
					i = j - 1
					continue
				}
			}
			// Single newline
			if i > best.pos {
				best = semanticBoundary{pos: i, score: scoreLine}
			}
			continue
		}

		// Code-block fence or closing brace on its own line.
		// Check for ```
		if b == '`' && i+2 < searchEnd && content[i+1] == '`' && content[i+2] == '`' {
			// Back up to start of the line if possible.
			start := i
			for start > 0 && content[start-1] != '\n' && content[start-1] != '\r' {
				start--
			}
			if start > best.pos {
				best = semanticBoundary{pos: start, score: scoreCodeBlock}
			}
			i += 2
			continue
		}
		// Check for "}\n" or "}\r" — common Go function end.
		if b == '}' {
			// If the next character after } is newline, the boundary is after the newline.
			if i+1 < searchEnd && (content[i+1] == '\n' || content[i+1] == '\r') {
				j := i + 1
				for j < searchEnd && (content[j] == '\n' || content[j] == '\r') {
					j++
				}
				if j > best.pos {
					best = semanticBoundary{pos: j, score: scoreCodeBlock}
				}
				i = j - 1
				continue
			}
		}

		// Sentence boundaries: . ? ! followed by space or newline.
		if (b == '.' || b == '?' || b == '!') && i+1 < searchEnd {
			next := content[i+1]
			if next == ' ' || next == '\n' || next == '\r' || next == '\t' {
				j := i + 1
				for j < searchEnd && (content[j] == ' ' || content[j] == '\t' || content[j] == '\n' || content[j] == '\r') {
					j++
				}
				if j > best.pos {
					best = semanticBoundary{pos: j, score: scoreSentence}
				}
				i = j - 1
				continue
			}
		}

		// Fallback whitespace boundary.
		if (b == ' ' || b == '\t') && i > best.pos && best.score < scoreHard {
			best = semanticBoundary{pos: i, score: scoreHard}
		}
	}

	// If no semantic boundary found, do a hard cut but land on a valid UTF-8 rune boundary.
	if best.pos == 0 && best.score < 0 {
		best.pos = maxBytes
		for best.pos > 0 && !utf8.RuneStart(content[best.pos]) {
			best.pos--
		}
		if best.pos == 0 {
			// Cannot safely cut at all; return empty.
			return "", 0
		}
		best.score = scoreFallback
	}

	return content[:best.pos], best.pos
}
