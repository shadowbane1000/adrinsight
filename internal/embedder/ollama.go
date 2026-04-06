package embedder

import (
	"context"
	"fmt"
	"os"

	"github.com/ollama/ollama/api"
)

// OllamaEmbedder generates embeddings using the Ollama API.
type OllamaEmbedder struct {
	client *api.Client
	model  string
}

// NewOllamaEmbedder creates an embedder that talks to Ollama at the given URL.
// If baseURL is empty, it uses the OLLAMA_HOST env var or defaults to localhost:11434.
func NewOllamaEmbedder(baseURL, model string) (*OllamaEmbedder, error) {
	if baseURL != "" {
		if err := os.Setenv("OLLAMA_HOST", baseURL); err != nil {
			return nil, fmt.Errorf("setting OLLAMA_HOST: %w", err)
		}
	}
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, fmt.Errorf("creating ollama client: %w", err)
	}
	return &OllamaEmbedder{client: client, model: model}, nil
}

// Embed generates embeddings for the given texts.
func (e *OllamaEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	req := &api.EmbedRequest{
		Model: e.model,
		Input: texts,
	}
	resp, err := e.client.Embed(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ollama embed: %w", err)
	}

	result := make([][]float32, len(resp.Embeddings))
	for i, emb := range resp.Embeddings {
		f32 := make([]float32, len(emb))
		for j, v := range emb {
			f32[j] = float32(v)
		}
		result[i] = f32
	}
	return result, nil
}

// Verify OllamaEmbedder implements Embedder.
var _ Embedder = (*OllamaEmbedder)(nil)
