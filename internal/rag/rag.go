package rag

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/shadowbane1000/adrinsight/internal/embedder"
	"github.com/shadowbane1000/adrinsight/internal/llm"
	"github.com/shadowbane1000/adrinsight/internal/store"
)

const queryPrefix = "search_query: "

// Pipeline orchestrates the RAG flow: embed → retrieve → rerank → expand → synthesize.
type Pipeline struct {
	Embedder embedder.Embedder
	Store    store.Store
	LLM      llm.LLM
	ADRDir   string
	TopK     int
	Reranker Reranker
}

// Query takes a natural-language question and returns a synthesized answer.
func (p *Pipeline) Query(ctx context.Context, question string) (llm.QueryResponse, error) {
	topK := p.TopK
	if topK <= 0 {
		topK = 5
	}

	// 1. Embed the question.
	vecs, err := p.Embedder.Embed(ctx, []string{queryPrefix + question})
	if err != nil {
		return llm.QueryResponse{}, fmt.Errorf("embedding query: %w", err)
	}
	if len(vecs) == 0 {
		return llm.QueryResponse{}, fmt.Errorf("no embedding returned for query")
	}

	// 2. Search for relevant chunks (hybrid: vector + keyword).
	results, err := p.Store.HybridSearch(ctx, vecs[0], question, topK, 0.6, 0.4)
	if err != nil {
		return llm.QueryResponse{}, fmt.Errorf("searching store: %w", err)
	}

	if len(results) == 0 {
		return p.LLM.Synthesize(ctx, question, nil)
	}

	// 3. Rerank results using domain heuristics.
	if p.Reranker != nil {
		results = p.Reranker.Rerank(question, results, DefaultRerankConfig())
	}

	// 4. Deduplicate by ADR number and collect paths.
	type adrInfo struct {
		number int
		title  string
		path   string
	}
	seen := make(map[int]bool)
	var adrs []adrInfo
	for _, r := range results {
		if seen[r.ADRNumber] {
			continue
		}
		seen[r.ADRNumber] = true
		adrs = append(adrs, adrInfo{number: r.ADRNumber, title: r.ADRTitle, path: r.ADRPath})
	}

	// 5. Record retrieved ADR numbers (deterministic, before synthesis).
	retrievedADRs := make([]int, len(adrs))
	for i, adr := range adrs {
		retrievedADRs[i] = adr.number
	}

	// 6. Read full ADR files from disk.
	var adrContexts []llm.ADRContext
	for _, adr := range adrs {
		content, err := os.ReadFile(adr.path)
		if err != nil {
			log.Printf("Warning: could not read ADR file %s: %v", adr.path, err)
			continue
		}
		adrContexts = append(adrContexts, llm.ADRContext{
			Number:  adr.number,
			Title:   adr.title,
			Content: string(content),
		})
	}

	// 7. Synthesize answer.
	resp, err := p.LLM.Synthesize(ctx, question, adrContexts)
	if err != nil {
		return resp, err
	}
	resp.RetrievedADRs = retrievedADRs
	return resp, nil
}
