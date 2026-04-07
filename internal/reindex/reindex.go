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

// KeywordExtractor extracts search keywords from an ADR.
type KeywordExtractor interface {
	ExtractKeywords(ctx context.Context, title, body string) ([]string, error)
}

// RelationshipClassifier classifies the type of a relationship from natural language.
type RelationshipClassifier interface {
	ClassifyRelationship(ctx context.Context, sourceTitle, bulletText string) (string, error)
}

// Reindexer orchestrates the parse → embed → store pipeline.
type Reindexer struct {
	Parser    parser.Parser
	Embedder  embedder.Embedder
	Store     store.Store
	ChunkFn   ChunkFunc        // optional override; defaults to Parser.ChunkADR
	RelClassifier RelationshipClassifier // optional; classifies relationship types via LLM
	Keywords  KeywordExtractor // optional; extracts search keywords from ADRs
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

	// Build ADR lookup map for title prefixing and store records.
	adrByNum := make(map[int]parser.ADR, len(adrs))
	for _, adr := range adrs {
		adrByNum[adr.Number] = adr
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

	// Extract and store keywords if a keyword extractor is available.
	if r.Keywords != nil {
		log.Println("Extracting keywords from ADRs...")
		var allKeywords []string
		for _, adr := range adrs {
			kw, err := r.Keywords.ExtractKeywords(ctx, adr.Title, adr.Body)
			if err != nil {
				log.Printf("Warning: keyword extraction failed for ADR-%03d: %v", adr.Number, err)
				continue
			}
			allKeywords = append(allKeywords, kw...)
		}
		if len(allKeywords) > 0 {
			if err := r.Store.StoreKeywords(ctx, allKeywords); err != nil {
				log.Printf("Warning: failed to store keywords: %v", err)
			} else {
				log.Printf("Stored %d keywords", len(allKeywords))
			}
		}
	}

	// Extract and store relationships if a classifier is available.
	if r.RelClassifier != nil {
		log.Println("Classifying ADR relationships...")
		var allRels []store.ADRRelationship
		for _, adr := range adrs {
			for _, raw := range adr.RelatedADRs {
				relType, err := r.RelClassifier.ClassifyRelationship(ctx, adr.Title, raw.Description)
				if err != nil {
					log.Printf("Warning: relationship classification failed for ADR-%03d → ADR-%03d: %v", adr.Number, raw.TargetADR, err)
					relType = store.RelRelatedTo
				}
				allRels = append(allRels, store.ADRRelationship{
					SourceADR:   adr.Number,
					TargetADR:   raw.TargetADR,
					RelType:     relType,
					Description: raw.Description,
				})
			}
		}
		if len(allRels) > 0 {
			if err := r.Store.StoreRelationships(ctx, allRels); err != nil {
				log.Printf("Warning: failed to store relationships: %v", err)
			} else {
				log.Printf("Stored %d relationships", len(allRels))
			}
		}
	}

	return &Result{ADRCount: len(adrs), ChunkCount: len(chunks)}, nil
}
