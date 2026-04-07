package rag

import (
	"sort"
	"strings"

	"github.com/shadowbane1000/adrinsight/internal/store"
)

// RerankConfig controls the reranking heuristics.
type RerankConfig struct {
	TitleBoost    float64 // boost for ADRs whose title matches query words (default 0.2)
	StatusPenalty float64 // penalty for superseded/deprecated ADRs (default 0.1)
	SectionBoost  float64 // boost for section-relevant matches (default 0.1)
}

// DefaultRerankConfig returns sensible defaults per the spec.
func DefaultRerankConfig() RerankConfig {
	return RerankConfig{
		TitleBoost:    0.2,
		StatusPenalty: 0.1,
		SectionBoost:  0.1,
	}
}

// Reranker adjusts search result ordering using domain heuristics.
type Reranker interface {
	Rerank(query string, results []store.SearchResult, config RerankConfig) []store.SearchResult
}

// DefaultReranker applies three heuristics: title match boost, status
// deprioritization, and section relevance boost.
type DefaultReranker struct{}

func (r *DefaultReranker) Rerank(query string, results []store.SearchResult, config RerankConfig) []store.SearchResult {
	queryLower := strings.ToLower(query)
	queryWords := strings.Fields(queryLower)

	// Work on a copy to avoid mutating the input slice.
	reranked := make([]store.SearchResult, len(results))
	copy(reranked, results)

	for i := range reranked {
		// 1. Title match boost: if any query word appears in the ADR title.
		titleLower := strings.ToLower(reranked[i].ADRTitle)
		for _, w := range queryWords {
			if strings.Contains(titleLower, w) {
				reranked[i].Score += config.TitleBoost
				break
			}
		}

		// 2. Status deprioritization: penalize superseded/deprecated ADRs.
		contentLower := strings.ToLower(reranked[i].Content)
		if strings.Contains(contentLower, "superseded") || strings.Contains(contentLower, "deprecated") {
			reranked[i].Score -= config.StatusPenalty
		}

		// 3. Section relevance boost: queries about rationale/alternatives
		//    get a boost for matching section types.
		if containsAny(queryLower, "why", "rationale", "alternative") {
			sectionLower := strings.ToLower(reranked[i].Section)
			if containsAny(sectionLower, "rationale", "alternative", "consequences") {
				reranked[i].Score += config.SectionBoost
			}
		}
	}

	sort.Slice(reranked, func(i, j int) bool {
		return reranked[i].Score > reranked[j].Score
	})

	return reranked
}

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
