package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/shadowbane1000/adrinsight/internal/llm"
	"github.com/shadowbane1000/adrinsight/internal/rag"
	"github.com/shadowbane1000/adrinsight/internal/store"
)

// --- mocks ---

type mockEmbedder struct{ shouldErr bool }

func (m *mockEmbedder) Embed(_ context.Context, texts []string) ([][]float32, error) {
	if m.shouldErr {
		return nil, fmt.Errorf("embedder unavailable")
	}
	result := make([][]float32, len(texts))
	for i := range texts {
		result[i] = make([]float32, 768)
	}
	return result, nil
}

type mockStore struct {
	adrs    []store.ADRSummary
	results []store.SearchResult
}

func (m *mockStore) Reset(_ context.Context) error                                  { return nil }
func (m *mockStore) StoreChunks(_ context.Context, _ []store.ChunkRecord) error     { return nil }
func (m *mockStore) Close() error                                                   { return nil }
func (m *mockStore) IsEmpty(_ context.Context) (bool, error)                 { return true, nil }
func (m *mockStore) StoreKeywords(_ context.Context, _ []string) error       { return nil }
func (m *mockStore) LoadKeywords(_ context.Context) (map[string]bool, error) { return nil, nil }
func (m *mockStore) ListADRs(_ context.Context) ([]store.ADRSummary, error) {
	return m.adrs, nil
}
func (m *mockStore) Search(_ context.Context, _ []float32, _ int) ([]store.SearchResult, error) {
	return m.results, nil
}
func (m *mockStore) SearchFTS(_ context.Context, _ string, _ int) ([]store.SearchResult, error) {
	return nil, nil
}
func (m *mockStore) HybridSearch(_ context.Context, _ []float32, _ string, _ int, _, _ float64) ([]store.SearchResult, error) {
	return m.results, nil
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

type mockLLM struct{ shouldErr bool }

func (m *mockLLM) Synthesize(_ context.Context, _ string, _ []llm.ADRContext) (llm.QueryResponse, error) {
	if m.shouldErr {
		return llm.QueryResponse{}, fmt.Errorf("synthesis service unavailable")
	}
	return llm.QueryResponse{
		Answer:    "Go was chosen for concurrency.",
		Citations: []llm.Citation{{ADRNumber: 1, Title: "Why Go", Section: "Rationale"}},
	}, nil
}

// --- helpers ---

func setupServer(t *testing.T, ms *mockStore, me *mockEmbedder, ml *mockLLM) (*Server, string) {
	t.Helper()
	dir := t.TempDir()

	// Create a sample ADR file.
	adrPath := filepath.Join(dir, "ADR-001-test.md")
	adrContent := "# ADR-001: Why Go\n\n**Status:** Accepted\n**Date:** 2026-04-06\n\n## Context\n\nTest content."
	if err := os.WriteFile(adrPath, []byte(adrContent), 0644); err != nil {
		t.Fatal(err)
	}

	if ms.adrs == nil {
		ms.adrs = []store.ADRSummary{
			{Number: 1, Title: "Why Go", Status: "Accepted", Path: adrPath},
		}
	}
	if ms.results == nil {
		ms.results = []store.SearchResult{
			{ADRNumber: 1, ADRTitle: "Why Go", ADRPath: adrPath, Section: "Context", Content: "Test"},
		}
	}

	p := &rag.Pipeline{Embedder: me, Store: ms, LLM: ml, ADRDir: dir, TopK: 5}
	s := &Server{Pipeline: p, Store: ms, Port: 0}
	return s, dir
}

// --- POST /query tests ---

func TestHandleQueryValid(t *testing.T) {
	s, _ := setupServer(t, &mockStore{}, &mockEmbedder{}, &mockLLM{})
	mux := s.NewServeMux()

	body := bytes.NewBufferString(`{"query":"Why Go?"}`)
	req := httptest.NewRequest("POST", "/query", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp queryResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Answer == "" {
		t.Error("expected non-empty answer")
	}
	if len(resp.Citations) == 0 {
		t.Error("expected at least one citation")
	}
}

func TestHandleQueryEmpty(t *testing.T) {
	s, _ := setupServer(t, &mockStore{}, &mockEmbedder{}, &mockLLM{})
	mux := s.NewServeMux()

	body := bytes.NewBufferString(`{"query":""}`)
	req := httptest.NewRequest("POST", "/query", body)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleQueryLLMError(t *testing.T) {
	s, _ := setupServer(t, &mockStore{}, &mockEmbedder{}, &mockLLM{shouldErr: true})
	mux := s.NewServeMux()

	body := bytes.NewBufferString(`{"query":"test"}`)
	req := httptest.NewRequest("POST", "/query", body)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for LLM error, got %d", w.Code)
	}
}

func TestHandleQueryEmbedderError(t *testing.T) {
	s, _ := setupServer(t, &mockStore{}, &mockEmbedder{shouldErr: true}, &mockLLM{})
	mux := s.NewServeMux()

	body := bytes.NewBufferString(`{"query":"test"}`)
	req := httptest.NewRequest("POST", "/query", body)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for embedder error, got %d", w.Code)
	}
}

// --- GET /adrs tests ---

func TestHandleListADRs(t *testing.T) {
	s, _ := setupServer(t, &mockStore{}, &mockEmbedder{}, &mockLLM{})
	mux := s.NewServeMux()

	req := httptest.NewRequest("GET", "/adrs", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp adrListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.ADRs) != 1 {
		t.Fatalf("expected 1 ADR, got %d", len(resp.ADRs))
	}
	if resp.ADRs[0].Number != 1 {
		t.Errorf("expected ADR 1, got %d", resp.ADRs[0].Number)
	}
}

func TestHandleListADRsEmpty(t *testing.T) {
	ms := &mockStore{adrs: []store.ADRSummary{}}
	s, _ := setupServer(t, ms, &mockEmbedder{}, &mockLLM{})
	mux := s.NewServeMux()

	req := httptest.NewRequest("GET", "/adrs", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// --- GET /adrs/{number} tests ---

func TestHandleGetADR(t *testing.T) {
	s, _ := setupServer(t, &mockStore{}, &mockEmbedder{}, &mockLLM{})
	mux := s.NewServeMux()

	req := httptest.NewRequest("GET", "/adrs/1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp adrDetailResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Number != 1 {
		t.Errorf("expected ADR 1, got %d", resp.Number)
	}
	if resp.Date != "2026-04-06" {
		t.Errorf("expected date 2026-04-06, got %q", resp.Date)
	}
	if resp.Content == "" {
		t.Error("expected non-empty content")
	}
}

func TestHandleGetADRNotFound(t *testing.T) {
	s, _ := setupServer(t, &mockStore{}, &mockEmbedder{}, &mockLLM{})
	mux := s.NewServeMux()

	req := httptest.NewRequest("GET", "/adrs/999", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandleGetADRInvalidNumber(t *testing.T) {
	s, _ := setupServer(t, &mockStore{}, &mockEmbedder{}, &mockLLM{})
	mux := s.NewServeMux()

	req := httptest.NewRequest("GET", "/adrs/abc", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
