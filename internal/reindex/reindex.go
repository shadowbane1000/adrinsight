package reindex

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/shadowbane1000/adrinsight/internal/embedder"
	"github.com/shadowbane1000/adrinsight/internal/parser"
	"github.com/shadowbane1000/adrinsight/internal/store"
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
	Parser        parser.Parser
	Embedder      embedder.Embedder
	Store         store.Store
	ChunkFn       ChunkFunc              // optional override; defaults to Parser.ChunkADR
	RelClassifier RelationshipClassifier // optional; classifies relationship types via LLM
	Keywords      KeywordExtractor       // optional; extracts search keywords from ADRs
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
	slog.Info("parsed ADRs", "count", len(adrs), "dir", adrDir)

	chunkFn := r.ChunkFn
	if chunkFn == nil {
		chunkFn = r.Parser.ChunkADR
	}
	var chunks []parser.Chunk
	for _, adr := range adrs {
		chunks = append(chunks, chunkFn(adr)...)
	}
	slog.Info("generated chunks", "count", len(chunks))

	if len(chunks) == 0 {
		if err := r.Store.Reset(ctx); err != nil {
			return nil, fmt.Errorf("resetting store: %w", err)
		}
		return &Result{ADRCount: len(adrs), ChunkCount: 0}, nil
	}

	adrByNum := make(map[int]parser.ADR, len(adrs))
	for _, adr := range adrs {
		adrByNum[adr.Number] = adr
	}

	texts := make([]string, len(chunks))
	for i, c := range chunks {
		texts[i] = docPrefix + c.Content
	}

	slog.Info("embedding chunks", "count", len(chunks))
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

	if err := r.Store.Reset(ctx); err != nil {
		return nil, fmt.Errorf("resetting store: %w", err)
	}
	slog.Info("storing chunks", "count", len(records))
	if err := r.Store.StoreChunks(ctx, records); err != nil {
		return nil, fmt.Errorf("storing chunks: %w", err)
	}

	if r.Keywords != nil {
		slog.Info("extracting keywords from ADRs")
		var allKeywords []string
		for _, adr := range adrs {
			kw, err := r.Keywords.ExtractKeywords(ctx, adr.Title, adr.Body)
			if err != nil {
				slog.Warn("keyword extraction failed", "adr", adr.Number, "error", err)
				continue
			}
			allKeywords = append(allKeywords, kw...)
		}
		if len(allKeywords) > 0 {
			if err := r.Store.StoreKeywords(ctx, allKeywords); err != nil {
				slog.Warn("failed to store keywords", "error", err)
			} else {
				slog.Info("stored keywords", "count", len(allKeywords))
			}
		}
	}

	if r.RelClassifier != nil {
		slog.Info("classifying ADR relationships")
		var allRels []store.ADRRelationship
		for _, adr := range adrs {
			for _, raw := range adr.RelatedADRs {
				relType, err := r.RelClassifier.ClassifyRelationship(ctx, adr.Title, raw.Description)
				if err != nil {
					slog.Warn("relationship classification failed",
						"source_adr", adr.Number, "target_adr", raw.TargetADR, "error", err)
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
				slog.Warn("failed to store relationships", "error", err)
			} else {
				slog.Info("stored relationships", "count", len(allRels))
			}
		}
	}

	return &Result{ADRCount: len(adrs), ChunkCount: len(chunks)}, nil
}
