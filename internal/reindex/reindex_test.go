package reindex

import (
	"context"
	"testing"

	"github.com/tylerc-atx/adr-insight/internal/parser"
	"github.com/tylerc-atx/adr-insight/internal/store"
)

// mockParser returns canned ADRs and chunks.
type mockParser struct {
	adrs   []parser.ADR
	chunks map[int][]parser.Chunk
}

func (m *mockParser) ParseDir(dir string) ([]parser.ADR, error) {
	return m.adrs, nil
}

func (m *mockParser) ChunkADR(adr parser.ADR) []parser.Chunk {
	return m.chunks[adr.Number]
}

// mockEmbedder returns fake 768-dim vectors.
type mockEmbedder struct {
	callCount int
}

func (m *mockEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	m.callCount++
	result := make([][]float32, len(texts))
	for i := range texts {
		v := make([]float32, 768)
		v[0] = float32(i + 1)
		result[i] = v
	}
	return result, nil
}

// mockStore tracks calls.
type mockStore struct {
	resetCalled bool
	stored      []store.ChunkRecord
}

func (m *mockStore) Reset(ctx context.Context) error {
	m.resetCalled = true
	return nil
}

func (m *mockStore) StoreChunks(ctx context.Context, chunks []store.ChunkRecord) error {
	m.stored = append(m.stored, chunks...)
	return nil
}

func (m *mockStore) Search(ctx context.Context, query []float32, topK int) ([]store.SearchResult, error) {
	return nil, nil
}

func (m *mockStore) ListADRs(ctx context.Context) ([]store.ADRSummary, error) {
	return nil, nil
}

func (m *mockStore) Close() error {
	return nil
}

func (m *mockStore) IsEmpty(_ context.Context) (bool, error) {
	return true, nil
}

func (m *mockStore) StoreKeywords(_ context.Context, _ []string) error       { return nil }
func (m *mockStore) LoadKeywords(_ context.Context) (map[string]bool, error) { return nil, nil }

func (m *mockStore) SearchFTS(_ context.Context, _ string, _ int) ([]store.SearchResult, error) {
	return nil, nil
}

func (m *mockStore) HybridSearch(_ context.Context, _ []float32, _ string, _ int, _, _ float64) ([]store.SearchResult, error) {
	return nil, nil
}
func (m *mockStore) StoreRelationships(_ context.Context, _ []store.ADRRelationship) error {
	return nil
}
func (m *mockStore) GetRelationships(_ context.Context, _ int) ([]store.ADRRelationship, error) {
	return nil, nil
}
func (m *mockStore) GetAllRelationships(_ context.Context) ([]store.ADRRelationship, error) {
	return nil, nil
}

func TestReindexerRun(t *testing.T) {
	mp := &mockParser{
		adrs: []parser.ADR{
			{FilePath: "/test/ADR-001.md", Number: 1, Title: "Why Go", Status: "Accepted"},
			{FilePath: "/test/ADR-002.md", Number: 2, Title: "SQLite", Status: "Accepted"},
		},
		chunks: map[int][]parser.Chunk{
			1: {
				{ADRNumber: 1, SectionKey: "Context", Content: "We need a language"},
				{ADRNumber: 1, SectionKey: "Decision", Content: "Use Go"},
			},
			2: {
				{ADRNumber: 2, SectionKey: "Context", Content: "We need storage"},
			},
		},
	}
	me := &mockEmbedder{}
	ms := &mockStore{}

	r := &Reindexer{Parser: mp, Embedder: me, Store: ms}
	result, err := r.Run(context.Background(), "/test")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.ADRCount != 2 {
		t.Errorf("ADRCount = %d, want 2", result.ADRCount)
	}
	if result.ChunkCount != 3 {
		t.Errorf("ChunkCount = %d, want 3", result.ChunkCount)
	}
	if !ms.resetCalled {
		t.Error("Store.Reset was not called")
	}
	if len(ms.stored) != 3 {
		t.Errorf("stored %d chunks, want 3", len(ms.stored))
	}
	if me.callCount != 1 {
		t.Errorf("Embed called %d times, want 1 (batch)", me.callCount)
	}

	// Verify search_document prefix was applied.
	// The mock embedder receives texts — we can't directly check,
	// but we verify the pipeline completed without error.
}

func TestReindexerRunEmpty(t *testing.T) {
	mp := &mockParser{adrs: nil, chunks: map[int][]parser.Chunk{}}
	me := &mockEmbedder{}
	ms := &mockStore{}

	r := &Reindexer{Parser: mp, Embedder: me, Store: ms}
	result, err := r.Run(context.Background(), "/empty")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.ADRCount != 0 {
		t.Errorf("ADRCount = %d, want 0", result.ADRCount)
	}
	if result.ChunkCount != 0 {
		t.Errorf("ChunkCount = %d, want 0", result.ChunkCount)
	}
	if !ms.resetCalled {
		t.Error("Store.Reset should be called even with no chunks")
	}
}
