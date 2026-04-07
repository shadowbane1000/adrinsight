package rag

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

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
		// Load authoritative supersession data if the reranker supports it.
		if dr, ok := p.Reranker.(*DefaultReranker); ok && dr.Superseded == nil {
			if allRels, err := p.Store.GetAllRelationships(ctx); err == nil {
				superseded := make(map[int]bool)
				for _, r := range allRels {
					if r.RelType == store.RelSupersededBy {
						superseded[r.SourceADR] = true
					}
				}
				dr.Superseded = superseded
			}
		}
		results = p.Reranker.Rerank(question, results, DefaultRerankConfig())
	}

	// 4. Deduplicate by ADR number and collect paths.
	seen := make(map[int]bool)
	var adrs []adrInfo
	for _, r := range results {
		if seen[r.ADRNumber] {
			continue
		}
		seen[r.ADRNumber] = true
		adrs = append(adrs, adrInfo{number: r.ADRNumber, title: r.ADRTitle, path: r.ADRPath})
	}

	// 5. Record retrieved ADR numbers (deterministic, before expansion).
	retrievedADRs := make([]int, len(adrs))
	for i, adr := range adrs {
		retrievedADRs[i] = adr.number
	}

	// 6. Expand with related ADRs for synthesis context only.
	// Expanded ADRs are sent to the LLM but not counted as "retrieved" for metrics.
	adrs, relSummary := p.expandWithRelationships(ctx, adrs, seen, topK+3)

	// 7. Read full ADR files from disk.
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

	// 8. Prepend relationship context to the first ADR context if available.
	if relSummary != "" && len(adrContexts) > 0 {
		adrContexts[0].Content = relSummary + "\n\n---\n\n" + adrContexts[0].Content
	}

	// 9. Synthesize answer.
	resp, err := p.LLM.Synthesize(ctx, question, adrContexts)
	if err != nil {
		return resp, err
	}
	resp.RetrievedADRs = retrievedADRs
	return resp, nil
}

type adrInfo struct {
	number int
	title  string
	path   string
}

// expandWithRelationships adds related ADRs to the result set and builds
// a relationship summary for the LLM. Walks supersession chains fully,
// adds 1-hop for other relationship types. Tracks visited nodes to prevent loops.
func (p *Pipeline) expandWithRelationships(ctx context.Context, adrs []adrInfo, seen map[int]bool, maxTotal int) ([]adrInfo, string) {
	// Load all relationships for the retrieved ADRs.
	var allRels []store.ADRRelationship
	for _, adr := range adrs {
		rels, err := p.Store.GetRelationships(ctx, adr.number)
		if err != nil {
			continue
		}
		allRels = append(allRels, rels...)
	}

	if len(allRels) == 0 {
		return adrs, ""
	}

	// Build ADR lookup for resolving titles/paths of expanded ADRs.
	adrList, err := p.Store.ListADRs(ctx)
	if err != nil {
		return adrs, ""
	}
	adrLookup := make(map[int]adrInfo, len(adrList))
	for _, a := range adrList {
		adrLookup[a.Number] = adrInfo{number: a.Number, title: a.Title, path: a.Path}
	}

	// Expand selectively: only add ADRs from supersession chains and
	// strong directional relationships (drives/depends_on), not all related_to.
	for _, rel := range allRels {
		if len(adrs) >= maxTotal {
			break
		}
		// Only expand for high-signal relationship types.
		switch rel.RelType {
		case store.RelSupersedes, store.RelSupersededBy, store.RelDrives, store.RelDependsOn:
			// Add the other side of the relationship.
		default:
			continue // skip related_to — too noisy for expansion
		}
		targets := []int{rel.TargetADR, rel.SourceADR}
		for _, t := range targets {
			if seen[t] {
				continue
			}
			if info, ok := adrLookup[t]; ok {
				seen[t] = true
				adrs = append(adrs, info)
			}
		}
	}

	// Walk supersession chains for any newly added ADRs.
	visited := make(map[int]bool)
	for k := range seen {
		visited[k] = true
	}
	for i := 0; i < len(adrs) && len(adrs) < maxTotal; i++ {
		p.walkSupersessionChain(ctx, adrs[i].number, visited, adrLookup, &adrs, maxTotal)
	}

	// Build relationship summary.
	var summary strings.Builder
	summary.WriteString("## Relationship Context\n\n")
	// Collect all relationships among the final ADR set.
	finalSet := make(map[int]bool, len(adrs))
	for _, a := range adrs {
		finalSet[a.number] = true
	}

	var relLines []string
	relSeen := make(map[string]bool)
	for _, a := range adrs {
		rels, _ := p.Store.GetRelationships(ctx, a.number)
		for _, r := range rels {
			if !finalSet[r.SourceADR] || !finalSet[r.TargetADR] {
				continue
			}
			key := fmt.Sprintf("%d-%s-%d", r.SourceADR, r.RelType, r.TargetADR)
			if relSeen[key] {
				continue
			}
			relSeen[key] = true
			relLines = append(relLines, fmt.Sprintf("- ADR-%03d %s ADR-%03d", r.SourceADR, r.RelType, r.TargetADR))
		}
	}

	if len(relLines) == 0 {
		return adrs, ""
	}

	for _, line := range relLines {
		summary.WriteString(line)
		summary.WriteString("\n")
	}
	return adrs, summary.String()
}

func (p *Pipeline) walkSupersessionChain(ctx context.Context, adrNum int, visited map[int]bool, lookup map[int]adrInfo, adrs *[]adrInfo, maxTotal int) {
	rels, err := p.Store.GetRelationships(ctx, adrNum)
	if err != nil {
		return
	}
	for _, r := range rels {
		if r.RelType != store.RelSupersedes && r.RelType != store.RelSupersededBy {
			continue
		}
		targets := []int{r.TargetADR, r.SourceADR}
		for _, t := range targets {
			if visited[t] || len(*adrs) >= maxTotal {
				continue
			}
			if info, ok := lookup[t]; ok {
				visited[t] = true
				*adrs = append(*adrs, info)
				// Recurse to follow the chain.
				p.walkSupersessionChain(ctx, t, visited, lookup, adrs, maxTotal)
			}
		}
	}
}
