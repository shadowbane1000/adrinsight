package rag

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/tylerc-atx/adr-insight/internal/llm"
	"github.com/tylerc-atx/adr-insight/internal/store"
)

type mockEmbedder struct{}

func (m *mockEmbedder) Embed(_ context.Context, texts []string) ([][]float32, error) {
	result := make([][]float32, len(texts))
	for i := range texts {
		result[i] = make([]float32, 768)
	}
	return result, nil
}

type mockStore struct {
	results []store.SearchResult
}

func (m *mockStore) Reset(_ context.Context) error                                    { return nil }
func (m *mockStore) StoreChunks(_ context.Context, _ []store.ChunkRecord) error       { return nil }
func (m *mockStore) ListADRs(_ context.Context) ([]store.ADRSummary, error)           { return nil, nil }
func (m *mockStore) Close() error                                                     { return nil }
func (m *mockStore) IsEmpty(_ context.Context) (bool, error) { return true, nil }
func (m *mockStore) SearchFTS(_ context.Context, _ string, _ int) ([]store.SearchResult, error) {
	return nil, nil
}
func (m *mockStore) HybridSearch(_ context.Context, _ []float32, _ string, _ int, _, _ float64) ([]store.SearchResult, error) {
	return m.results, nil
}
func (m *mockStore) Search(_ context.Context, _ []float32, _ int) ([]store.SearchResult, error) {
	return m.results, nil
}

type mockLLM struct {
	receivedContexts []llm.ADRContext
}

func (m *mockLLM) Synthesize(_ context.Context, _ string, adrs []llm.ADRContext) (llm.QueryResponse, error) {
	m.receivedContexts = adrs
	return llm.QueryResponse{
		Answer: "Test answer",
		Citations: []llm.Citation{
			{ADRNumber: 1, Title: "Test", Section: "Context"},
		},
	}, nil
}

func TestPipelineQueryDeduplication(t *testing.T) {
	// Create temp ADR file.
	dir := t.TempDir()
	adrPath := filepath.Join(dir, "ADR-001.md")
	if err := os.WriteFile(adrPath, []byte("# ADR-001: Test\n\nFull content here."), 0644); err != nil {
		t.Fatal(err)
	}

	ml := &mockLLM{}
	p := Pipeline{
		Embedder: &mockEmbedder{},
		Store: &mockStore{
			results: []store.SearchResult{
				{ADRNumber: 1, ADRTitle: "Test", ADRPath: adrPath, Section: "Context", Content: "chunk1"},
				{ADRNumber: 1, ADRTitle: "Test", ADRPath: adrPath, Section: "Decision", Content: "chunk2"},
			},
		},
		LLM:    ml,
		ADRDir: dir,
		TopK:   5,
	}

	resp, err := p.Query(context.Background(), "test question")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if resp.Answer == "" {
		t.Error("expected non-empty answer")
	}

	// Should deduplicate: 2 chunks from same ADR → 1 ADRContext.
	if len(ml.receivedContexts) != 1 {
		t.Fatalf("expected 1 ADR context (deduplicated), got %d", len(ml.receivedContexts))
	}
	if ml.receivedContexts[0].Number != 1 {
		t.Errorf("expected ADR-001, got ADR-%03d", ml.receivedContexts[0].Number)
	}
}

func TestPipelineQueryMissingFile(t *testing.T) {
	ml := &mockLLM{}
	p := Pipeline{
		Embedder: &mockEmbedder{},
		Store: &mockStore{
			results: []store.SearchResult{
				{ADRNumber: 1, ADRTitle: "Test", ADRPath: "/nonexistent/ADR-001.md", Section: "Context", Content: "chunk"},
			},
		},
		LLM:  ml,
		TopK: 5,
	}

	// Should not error — just skip the missing file and continue with empty context.
	_, err := p.Query(context.Background(), "test")
	if err != nil {
		t.Fatalf("Query should not fail on missing file: %v", err)
	}
}

func TestPipelineQueryNoResults(t *testing.T) {
	ml := &mockLLM{}
	p := Pipeline{
		Embedder: &mockEmbedder{},
		Store:    &mockStore{results: nil},
		LLM:     ml,
		TopK:    5,
	}

	_, err := p.Query(context.Background(), "test")
	if err != nil {
		t.Fatalf("Query with no results: %v", err)
	}
	// LLM should be called with nil/empty ADR contexts.
	if len(ml.receivedContexts) != 0 {
		t.Errorf("expected 0 ADR contexts for no results, got %d", len(ml.receivedContexts))
	}
}
