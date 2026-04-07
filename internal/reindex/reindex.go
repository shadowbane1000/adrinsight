package reindex

import (
	"context"
	"fmt"
	"log"

	"github.com/tylerc-atx/adr-insight/internal/embedder"
	"github.com/tylerc-atx/adr-insight/internal/parser"
	"github.com/tylerc-atx/adr-insight/internal/store"
)

const docPrefix = "search_document: "

// ChunkFunc is a function that splits an ADR into chunks.
type ChunkFunc func(adr parser.ADR) []parser.Chunk

// Reindexer orchestrates the parse → embed → store pipeline.
type Reindexer struct {
	Parser   parser.Parser
	Embedder embedder.Embedder
	Store    store.Store
	ChunkFn  ChunkFunc // optional override; defaults to Parser.ChunkADR
}

// Result holds summary information about a reindex run.
type Result struct {
	ADRCount   int
	ChunkCount int
}

// Run executes the full reindex pipeline.
func (r *Reindexer) Run(ctx context.Context, adrDir string) (*Result, error) {
	adrs, err := r.Parser.ParseDir(adrDir)
	if err != nil {
		return nil, fmt.Errorf("parsing ADRs: %w", err)
	}
	log.Printf("Parsed %d ADRs from %s", len(adrs), adrDir)

	// Chunk all ADRs.
	chunkFn := r.ChunkFn
	if chunkFn == nil {
		chunkFn = r.Parser.ChunkADR
	}
	var chunks []parser.Chunk
	for _, adr := range adrs {
		chunks = append(chunks, chunkFn(adr)...)
	}
	log.Printf("Generated %d chunks", len(chunks))

	if len(chunks) == 0 {
		// Reset the store even with no chunks.
		if err := r.Store.Reset(ctx); err != nil {
			return nil, fmt.Errorf("resetting store: %w", err)
		}
		return &Result{ADRCount: len(adrs), ChunkCount: 0}, nil
	}

	// Prepare texts for embedding with search_document prefix.
	texts := make([]string, len(chunks))
	for i, c := range chunks {
		texts[i] = docPrefix + c.Content
	}

	// Embed all chunks.
	log.Printf("Embedding %d chunks via Ollama...", len(chunks))
	embeddings, err := r.Embedder.Embed(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("embedding chunks: %w", err)
	}
	if len(embeddings) != len(chunks) {
		return nil, fmt.Errorf("embedding count mismatch: got %d, want %d", len(embeddings), len(chunks))
	}

	// Build store records by joining chunk metadata with ADR info.
	adrByNum := make(map[int]parser.ADR, len(adrs))
	for _, adr := range adrs {
		adrByNum[adr.Number] = adr
	}

	records := make([]store.ChunkRecord, len(chunks))
	for i, c := range chunks {
		adr := adrByNum[c.ADRNumber]
		records[i] = store.ChunkRecord{
			ADRNumber: c.ADRNumber,
			ADRTitle:  adr.Title,
			ADRStatus: adr.Status,
			ADRPath:   adr.FilePath,
			Section:   c.SectionKey,
			Content:   c.Content,
			Embedding: embeddings[i],
		}
	}

	// Reset and store.
	if err := r.Store.Reset(ctx); err != nil {
		return nil, fmt.Errorf("resetting store: %w", err)
	}
	log.Printf("Storing %d chunks...", len(records))
	if err := r.Store.StoreChunks(ctx, records); err != nil {
		return nil, fmt.Errorf("storing chunks: %w", err)
	}

	return &Result{ADRCount: len(adrs), ChunkCount: len(chunks)}, nil
}
