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

// Store persists ADR chunks with embeddings and supports similarity search.
type Store interface {
	Reset(ctx context.Context) error
	StoreChunks(ctx context.Context, chunks []ChunkRecord) error
	Search(ctx context.Context, query []float32, topK int) ([]SearchResult, error)
	ListADRs(ctx context.Context) ([]ADRSummary, error)
	IsEmpty(ctx context.Context) (bool, error)
	Close() error
}
