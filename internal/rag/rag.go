package rag

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/tylerc-atx/adr-insight/internal/embedder"
	"github.com/tylerc-atx/adr-insight/internal/llm"
	"github.com/tylerc-atx/adr-insight/internal/store"
)

const queryPrefix = "search_query: "

// Pipeline orchestrates the RAG flow: embed → retrieve → expand → synthesize.
type Pipeline struct {
	Embedder embedder.Embedder
	Store    store.Store
	LLM      llm.LLM
	ADRDir   string
	TopK     int
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

	// 2. Search for relevant chunks.
	results, err := p.Store.Search(ctx, vecs[0], topK)
	if err != nil {
		return llm.QueryResponse{}, fmt.Errorf("searching store: %w", err)
	}

	if len(results) == 0 {
		return p.LLM.Synthesize(ctx, question, nil)
	}

	// 3. Deduplicate by ADR number and collect paths.
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

	// 4. Read full ADR files from disk.
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

	// 5. Synthesize answer.
	return p.LLM.Synthesize(ctx, question, adrContexts)
}
