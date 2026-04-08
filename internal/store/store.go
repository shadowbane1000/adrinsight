// Package store provides vector and metadata storage via SQLite.
package store

import "context"

// ChunkRecord holds all data for a single chunk to be stored.
type ChunkRecord struct {
	ADRNumber int
	ADRTitle  string
	ADRStatus string
	ADRPath   string
	Section   string
	Content   string
	Embedding []float32
}

// SearchResult is a single result from a similarity search.
type SearchResult struct {
	ADRNumber int
	ADRTitle  string
	ADRPath   string
	Section   string
	Content   string
	Score     float64
}

// ADRSummary is a lightweight representation of an indexed ADR.
type ADRSummary struct {
	Number int
	Title  string
	Status string
	Path   string
}

// Relationship type constants.
const (
	RelSupersedes   = "supersedes"
	RelSupersededBy = "superseded_by"
	RelDependsOn    = "depends_on"
	RelDrives       = "drives"
	RelRelatedTo    = "related_to"
)

// ADRRelationship is a directed, typed relationship between two ADRs.
type ADRRelationship struct {
	SourceADR   int
	TargetADR   int
	RelType     string
	Description string
}

// Store persists ADR chunks with embeddings and supports similarity search.
type Store interface {
	Reset(ctx context.Context) error
	StoreChunks(ctx context.Context, chunks []ChunkRecord) error
	Search(ctx context.Context, query []float32, topK int) ([]SearchResult, error)
	SearchFTS(ctx context.Context, query string, topK int) ([]SearchResult, error)
	HybridSearch(ctx context.Context, queryVec []float32, queryText string, topK int, vecWeight, kwWeight float64) ([]SearchResult, error)
	StoreKeywords(ctx context.Context, words []string) error
	LoadKeywords(ctx context.Context) (map[string]bool, error)
	StoreRelationships(ctx context.Context, rels []ADRRelationship) error
	GetRelationships(ctx context.Context, adrNumber int) ([]ADRRelationship, error)
	GetAllRelationships(ctx context.Context) ([]ADRRelationship, error)
	ListADRs(ctx context.Context) ([]ADRSummary, error)
	IsEmpty(ctx context.Context) (bool, error)
	Close() error
}
