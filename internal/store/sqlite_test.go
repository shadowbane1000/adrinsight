package store

import (
	"context"
	"path/filepath"
	"testing"
)

func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestReset(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	if err := s.Reset(ctx); err != nil {
		t.Fatalf("Reset: %v", err)
	}
	// Verify tables exist by inserting a row.
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO chunks (adr_number, adr_title, adr_status, adr_path, section, content)
		 VALUES (1, 'Test', 'Accepted', '/test.md', 'Context', 'content')`)
	if err != nil {
		t.Fatalf("insert after reset: %v", err)
	}
}

func TestStoreChunks(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	if err := s.Reset(ctx); err != nil {
		t.Fatalf("Reset: %v", err)
	}

	// Create a fake embedding (768 dimensions).
	emb := make([]float32, 768)
	for i := range emb {
		emb[i] = float32(i) / 768.0
	}

	chunks := []ChunkRecord{
		{
			ADRNumber: 1, ADRTitle: "Why Go", ADRStatus: "Accepted",
			ADRPath: "/docs/adr/ADR-001.md", Section: "Context",
			Content: "We need a language", Embedding: emb,
		},
		{
			ADRNumber: 1, ADRTitle: "Why Go", ADRStatus: "Accepted",
			ADRPath: "/docs/adr/ADR-001.md", Section: "Decision",
			Content: "Use Go", Embedding: emb,
		},
	}

	if err := s.StoreChunks(ctx, chunks); err != nil {
		t.Fatalf("StoreChunks: %v", err)
	}

	// Verify rows inserted.
	var count int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM chunks").Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 chunks, got %d", count)
	}
}

func TestSearch(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	if err := s.Reset(ctx); err != nil {
		t.Fatalf("Reset: %v", err)
	}

	// Create two different embeddings.
	emb1 := make([]float32, 768)
	emb2 := make([]float32, 768)
	for i := range emb1 {
		emb1[i] = float32(i) / 768.0
		emb2[i] = float32(768-i) / 768.0
	}

	chunks := []ChunkRecord{
		{
			ADRNumber: 1, ADRTitle: "Why Go", ADRStatus: "Accepted",
			ADRPath: "/docs/adr/ADR-001.md", Section: "Context",
			Content: "We chose Go for concurrency", Embedding: emb1,
		},
		{
			ADRNumber: 2, ADRTitle: "SQLite Storage", ADRStatus: "Accepted",
			ADRPath: "/docs/adr/ADR-002.md", Section: "Context",
			Content: "We chose SQLite for simplicity", Embedding: emb2,
		},
	}

	if err := s.StoreChunks(ctx, chunks); err != nil {
		t.Fatalf("StoreChunks: %v", err)
	}

	// Search with emb1 — should return ADR-001 first.
	results, err := s.Search(ctx, emb1, 2)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].ADRNumber != 1 {
		t.Errorf("expected ADR-001 as top result, got ADR-%03d", results[0].ADRNumber)
	}
}

func TestListADRs(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	if err := s.Reset(ctx); err != nil {
		t.Fatalf("Reset: %v", err)
	}

	emb := make([]float32, 768)
	chunks := []ChunkRecord{
		{ADRNumber: 1, ADRTitle: "Why Go", ADRStatus: "Accepted", ADRPath: "/adr/ADR-001.md", Section: "Context", Content: "content", Embedding: emb},
		{ADRNumber: 1, ADRTitle: "Why Go", ADRStatus: "Accepted", ADRPath: "/adr/ADR-001.md", Section: "Decision", Content: "content", Embedding: emb},
		{ADRNumber: 2, ADRTitle: "SQLite", ADRStatus: "Accepted", ADRPath: "/adr/ADR-002.md", Section: "Context", Content: "content", Embedding: emb},
	}
	if err := s.StoreChunks(ctx, chunks); err != nil {
		t.Fatalf("StoreChunks: %v", err)
	}

	adrs, err := s.ListADRs(ctx)
	if err != nil {
		t.Fatalf("ListADRs: %v", err)
	}
	if len(adrs) != 2 {
		t.Fatalf("expected 2 unique ADRs, got %d", len(adrs))
	}
	if adrs[0].Number != 1 || adrs[1].Number != 2 {
		t.Errorf("expected ADR 1 and 2, got %d and %d", adrs[0].Number, adrs[1].Number)
	}
	if adrs[0].Title != "Why Go" {
		t.Errorf("expected title 'Why Go', got %q", adrs[0].Title)
	}
}

func TestClose(t *testing.T) {
	s := newTestStore(t)
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}
