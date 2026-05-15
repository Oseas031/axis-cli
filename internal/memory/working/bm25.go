package working

import (
	"math"
	"sort"
	"strings"
	"unicode"
)

// bm25 implements a simple BM25 scorer for in-memory document ranking.
// No external dependencies — tokenization is whitespace + CJK bigram.
// v1: rebuilds index on every Recall call. TODO: incremental index, pluggable Scorer interface.
const (
	bm25K1 = 1.2  // v1: standard default. TODO: tune on real workloads.
	bm25B  = 0.75 // v1: standard default. TODO: tune on real workloads.
)

// bm25Index is an inverted index for BM25 scoring.
type bm25Index struct {
	// docFreq[term] = number of documents containing term
	docFreq map[string]int
	// docs[docID] = {term: frequency}
	docs []bm25Doc
	// avgDL is the average document length (in tokens)
	avgDL float64
}

type bm25Doc struct {
	id     string
	tf     map[string]int
	length int
}

// newBM25Index builds an index from a set of documents.
func newBM25Index(documents map[string]string) *bm25Index {
	idx := &bm25Index{
		docFreq: make(map[string]int),
	}

	totalLen := 0
	for id, text := range documents {
		tokens := tokenize(text)
		tf := make(map[string]int, len(tokens))
		for _, t := range tokens {
			tf[t]++
		}
		idx.docs = append(idx.docs, bm25Doc{id: id, tf: tf, length: len(tokens)})
		totalLen += len(tokens)

		for term := range tf {
			idx.docFreq[term]++
		}
	}

	if len(idx.docs) > 0 {
		idx.avgDL = float64(totalLen) / float64(len(idx.docs))
	}
	return idx
}

// score returns BM25 scores for all documents, sorted descending.
func (idx *bm25Index) score(query string) []bm25Result {
	queryTerms := tokenize(query)
	if len(queryTerms) == 0 {
		return nil
	}

	n := float64(len(idx.docs))
	var results []bm25Result

	for _, doc := range idx.docs {
		score := 0.0
		for _, term := range queryTerms {
			df := float64(idx.docFreq[term])
			if df == 0 {
				continue
			}
			tf := float64(doc.tf[term])
			if tf == 0 {
				continue
			}
			// IDF: log((N - df + 0.5) / (df + 0.5) + 1)
			idf := math.Log((n-df+0.5)/(df+0.5) + 1)
			// TF normalization
			dl := float64(doc.length)
			tfNorm := (tf * (bm25K1 + 1)) / (tf + bm25K1*(1-bm25B+bm25B*dl/idx.avgDL))
			score += idf * tfNorm
		}
		if score > 0 {
			results = append(results, bm25Result{id: doc.id, score: score})
		}
	}

	// Sort descending by score
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})
	return results
}

type bm25Result struct {
	id    string
	score float64
}

// tokenize splits text into tokens: lowercased words + CJK bigrams.
// v1: bigram for CJK, whitespace for Latin. TODO: proper segmentation.
func tokenize(text string) []string {
	text = strings.ToLower(text)
	var tokens []string
	var word strings.Builder
	var prevCJK rune
	hasPrevCJK := false

	for _, r := range text {
		if isCJK(r) {
			// Flush any pending ASCII word
			if word.Len() > 0 {
				tokens = append(tokens, word.String())
				word.Reset()
			}
			// CJK bigram: emit pair of (prev, current)
			if hasPrevCJK {
				tokens = append(tokens, string(prevCJK)+string(r))
			}
			// Also emit unigram for single-char queries
			tokens = append(tokens, string(r))
			prevCJK = r
			hasPrevCJK = true
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			hasPrevCJK = false
			word.WriteRune(r)
		} else {
			// Separator
			hasPrevCJK = false
			if word.Len() > 0 {
				tokens = append(tokens, word.String())
				word.Reset()
			}
		}
	}
	if word.Len() > 0 {
		tokens = append(tokens, word.String())
	}
	return tokens
}

func isCJK(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) ||
		(r >= 0x3400 && r <= 0x4DBF) ||
		(r >= 0xF900 && r <= 0xFAFF)
}
