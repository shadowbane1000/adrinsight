package embedder

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewOllamaEmbedder(t *testing.T) {
	e, err := NewOllamaEmbedder("http://localhost:11434", "nomic-embed-text")
	if err != nil {
		t.Fatalf("NewOllamaEmbedder: %v", err)
	}
	if e == nil {
		t.Fatal("expected non-nil embedder")
	}
}

func TestOllamaEmbedderInterfaceCompliance(t *testing.T) {
	var _ Embedder = (*OllamaEmbedder)(nil)
}

func TestOllamaEmbedUnreachable(t *testing.T) {
	// Use a server that immediately closes to simulate unreachable.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	e, err := NewOllamaEmbedder(srv.URL, "nomic-embed-text")
	if err != nil {
		t.Fatalf("NewOllamaEmbedder: %v", err)
	}

	_, err = e.Embed(context.Background(), []string{"test"})
	if err == nil {
		t.Error("expected error when Ollama returns error, got nil")
	}
}

func TestOllamaEmbedIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	e, err := NewOllamaEmbedder("http://localhost:11434", "nomic-embed-text")
	if err != nil {
		t.Fatalf("NewOllamaEmbedder: %v", err)
	}

	ctx := context.Background()
	vecs, err := e.Embed(ctx, []string{"Why did we choose Go?"})
	if err != nil {
		t.Skipf("Ollama not available: %v", err)
	}

	if len(vecs) != 1 {
		t.Fatalf("expected 1 embedding, got %d", len(vecs))
	}
	if len(vecs[0]) != 768 {
		t.Errorf("expected 768 dimensions, got %d", len(vecs[0]))
	}
}
