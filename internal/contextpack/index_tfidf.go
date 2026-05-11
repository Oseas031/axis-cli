package contextpack

import (
	"math"
	"sort"
	"strings"
	"unicode"
)

// ScoredChunk pairs a document chunk with its query relevance score.
type ScoredChunk struct {
	Chunk DocumentChunk
	Score float64
}

// TFIDFIndex holds an in-memory TF-IDF index over document chunks.
type TFIDFIndex struct {
	IDF     map[string]float64   `json:"idf"`
	Vectors []map[string]float64 `json:"vectors"`
	Chunks  []DocumentChunk      `json:"chunks"`
}

var stopWords = map[string]bool{
	"the": true, "a": true, "an": true, "and": true, "or": true,
	"but": true, "in": true, "on": true, "at": true, "to": true,
	"for": true, "of": true, "with": true, "by": true, "from": true,
	"up": true, "about": true, "into": true, "through": true,
	"during": true, "before": true, "after": true, "above": true,
	"below": true, "between": true, "among": true, "is": true,
	"are": true, "was": true, "were": true, "be": true, "been": true,
	"being": true, "have": true, "has": true, "had": true, "do": true,
	"does": true, "did": true, "will": true, "would": true,
	"shall": true, "should": true, "may": true, "might": true,
	"can": true, "could": true, "must": true, "this": true,
	"that": true, "these": true, "those": true, "i": true,
	"you": true, "he": true, "she": true, "it": true, "we": true,
	"they": true, "me": true, "him": true, "her": true, "us": true,
	"them": true, "my": true, "your": true, "his": true,
	"its": true, "our": true, "their": true, "what": true,
	"which": true, "who": true, "whom": true, "whose": true,
	"where": true, "when": true, "why": true, "how": true,
	"all": true, "each": true, "few": true, "more": true,
	"most": true, "other": true, "some": true, "such": true,
	"no": true, "nor": true, "not": true, "only": true,
	"own": true, "same": true, "so": true, "than": true,
	"too": true, "very": true, "just": true, "now": true,
	"then": true, "here": true, "there": true, "both": true,
	"either": true, "neither": true, "one": true, "two": true,
	"first": true, "last": true, "next": true, "many": true,
	"much": true, "another": true,
}

func tokenize(text string) []string {
	var tokens []string
	text = strings.ToLower(text)
	start := -1
	for i, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if start == -1 {
				start = i
			}
		} else {
			if start != -1 {
				token := text[start:i]
				if !stopWords[token] && len(token) > 1 {
					tokens = append(tokens, token)
				}
				start = -1
			}
		}
	}
	if start != -1 {
		token := text[start:]
		if !stopWords[token] && len(token) > 1 {
			tokens = append(tokens, token)
		}
	}
	return tokens
}

func termFreq(tokens []string) map[string]float64 {
	tf := make(map[string]float64)
	for _, t := range tokens {
		tf[t]++
	}
	maxFreq := 0.0
	for _, c := range tf {
		if c > maxFreq {
			maxFreq = c
		}
	}
	if maxFreq > 0 {
		for t := range tf {
			tf[t] /= maxFreq
		}
	}
	return tf
}

// Build computes TF-IDF vectors for all chunks. Replaces any existing index state.
func (idx *TFIDFIndex) Build(chunks []DocumentChunk) {
	idx.Chunks = chunks
	n := len(chunks)
	if n == 0 {
		idx.IDF = make(map[string]float64)
		idx.Vectors = nil
		return
	}

	allTokens := make([][]string, n)
	for i, chunk := range chunks {
		allTokens[i] = tokenize(chunk.Content)
	}

	df := make(map[string]int)
	for _, tokens := range allTokens {
		seen := make(map[string]bool)
		for _, t := range tokens {
			if !seen[t] {
				seen[t] = true
				df[t]++
			}
		}
	}

	idx.IDF = make(map[string]float64)
	for t, count := range df {
		idx.IDF[t] = math.Log(float64(n)/float64(count)) + 1.0
	}

	idx.Vectors = make([]map[string]float64, n)
	for i, tokens := range allTokens {
		tf := termFreq(tokens)
		vec := make(map[string]float64)
		var norm float64
		for t, freq := range tf {
			if idf, ok := idx.IDF[t]; ok {
				weight := freq * idf
				vec[t] = weight
				norm += weight * weight
			}
		}
		if norm > 0 {
			norm = math.Sqrt(norm)
			for t := range vec {
				vec[t] /= norm
			}
		}
		idx.Vectors[i] = vec
	}
}

// Query ranks document chunks by cosine similarity to the query text.
// Returns up to topK results sorted by descending score.
func (idx *TFIDFIndex) Query(text string, topK int) []ScoredChunk {
	if len(idx.Chunks) == 0 || len(idx.Vectors) == 0 {
		return nil
	}

	tokens := tokenize(text)
	if len(tokens) == 0 {
		return nil
	}

	tf := termFreq(tokens)
	qvec := make(map[string]float64)
	var norm float64
	for t, freq := range tf {
		if idf, ok := idx.IDF[t]; ok {
			weight := freq * idf
			qvec[t] = weight
			norm += weight * weight
		}
	}
	if norm > 0 {
		norm = math.Sqrt(norm)
		for t := range qvec {
			qvec[t] /= norm
		}
	}

	scored := make([]ScoredChunk, 0, len(idx.Chunks))
	for i, dvec := range idx.Vectors {
		var score float64
		for t, qw := range qvec {
			if dw, ok := dvec[t]; ok {
				score += qw * dw
			}
		}
		scored = append(scored, ScoredChunk{Chunk: idx.Chunks[i], Score: score})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	if topK > 0 && topK < len(scored) {
		scored = scored[:topK]
	}
	return scored
}
