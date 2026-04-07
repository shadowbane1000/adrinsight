package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/tylerc-atx/adr-insight/internal/embedder"
	"github.com/tylerc-atx/adr-insight/internal/llm"
	"github.com/tylerc-atx/adr-insight/internal/parser"
	"github.com/tylerc-atx/adr-insight/internal/rag"
	"github.com/tylerc-atx/adr-insight/internal/reindex"
	"github.com/tylerc-atx/adr-insight/internal/server"
	"github.com/tylerc-atx/adr-insight/internal/store"
)

const queryPrefix = "search_query: "

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: adr-insight <command> [flags]\n\nCommands:\n  reindex  Parse, embed, and store ADRs\n  search   Search indexed ADRs by similarity\n  serve    Start the HTTP API server\n")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "reindex":
		cmdReindex(os.Args[2:])
	case "search":
		cmdSearch(os.Args[2:])
	case "serve":
		cmdServe(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func cmdReindex(args []string) {
	fs := flag.NewFlagSet("reindex", flag.ExitOnError)
	adrDir := fs.String("adr-dir", "./docs/adr", "Directory containing ADR files")
	dbPath := fs.String("db", "./adr-insight.db", "Path to SQLite database")
	ollamaURL := fs.String("ollama-url", "http://localhost:11434", "Ollama API base URL")
	_ = fs.Parse(args)

	ctx := context.Background()

	p := parser.NewMarkdownParser()

	emb, err := embedder.NewOllamaEmbedder(*ollamaURL, "nomic-embed-text")
	if err != nil {
		log.Fatalf("Failed to create embedder: %v", err)
	}

	st, err := store.NewSQLiteStore(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = st.Close() }()

	r := &reindex.Reindexer{Parser: p, Embedder: emb, Store: st}
	result, err := r.Run(ctx, *adrDir)
	if err != nil {
		log.Fatalf("Reindex failed: %v", err)
	}

	fmt.Printf("Reindex complete: %d ADRs, %d chunks\n", result.ADRCount, result.ChunkCount)
}

func cmdSearch(args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	dbPath := fs.String("db", "./adr-insight.db", "Path to SQLite database")
	ollamaURL := fs.String("ollama-url", "http://localhost:11434", "Ollama API base URL")
	topK := fs.Int("top-k", 5, "Number of results to return")
	_ = fs.Parse(args)

	query := strings.Join(fs.Args(), " ")
	if query == "" {
		fmt.Fprintln(os.Stderr, "Usage: adr-insight search [flags] <query>")
		os.Exit(1)
	}

	ctx := context.Background()

	emb, err := embedder.NewOllamaEmbedder(*ollamaURL, "nomic-embed-text")
	if err != nil {
		log.Fatalf("Failed to create embedder: %v", err)
	}

	st, err := store.NewSQLiteStore(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = st.Close() }()

	// Embed the query with search_query prefix.
	vecs, err := emb.Embed(ctx, []string{queryPrefix + query})
	if err != nil {
		log.Fatalf("Failed to embed query: %v", err)
	}
	if len(vecs) == 0 {
		log.Fatal("No embedding returned for query")
	}

	results, err := st.Search(ctx, vecs[0], *topK)
	if err != nil {
		log.Fatalf("Search failed: %v", err)
	}

	fmt.Printf("Query: %q\n\n", query)
	if len(results) == 0 {
		fmt.Println("No results found.")
		return
	}

	for i, r := range results {
		preview := r.Content
		if len(preview) > 150 {
			preview = preview[:150] + "..."
		}
		fmt.Printf("%d. [ADR-%03d] %s — %s (distance: %.2f)\n   %s\n\n",
			i+1, r.ADRNumber, r.ADRTitle, r.Section, r.Score, preview)
	}
}

func cmdServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.Int("port", 8081, "HTTP server port")
	dbPath := fs.String("db", "./adr-insight.db", "Path to SQLite database")
	ollamaURL := fs.String("ollama-url", "http://localhost:11434", "Ollama API base URL")
	adrDir := fs.String("adr-dir", "./docs/adr", "ADR directory for full content reads")
	model := fs.String("model", "claude-sonnet-4-5", "Anthropic model to use")
	devMode := fs.Bool("dev", false, "Serve static files from disk for live editing")
	_ = fs.Parse(args)

	emb, err := embedder.NewOllamaEmbedder(*ollamaURL, "nomic-embed-text")
	if err != nil {
		log.Fatalf("Failed to create embedder: %v", err)
	}

	st, err := store.NewSQLiteStore(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = st.Close() }()

	// Auto-reindex if the database is empty.
	autoReindex(context.Background(), st, emb, *adrDir)

	// Build pipeline only if API key is available.
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	var pipeline *rag.Pipeline
	if apiKey == "" {
		log.Println("WARNING: ANTHROPIC_API_KEY not set — query endpoint disabled, ADR browsing available")
	} else {
		l := llm.NewAnthropicLLM(apiKey, *model)
		pipeline = &rag.Pipeline{
			Embedder: emb,
			Store:    st,
			LLM:     l,
			ADRDir:   *adrDir,
			TopK:    5,
		}
	}

	srv := &server.Server{
		Pipeline: pipeline,
		Store:    st,
		Parser:   parser.NewMarkdownParser(),
		Port:     *port,
		DevMode:  *devMode,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// autoReindex checks if the database is empty and runs reindex if needed.
// Retries with exponential backoff if Ollama is not yet ready.
func autoReindex(ctx context.Context, st *store.SQLiteStore, emb *embedder.OllamaEmbedder, adrDir string) {
	empty, err := st.IsEmpty(ctx)
	if err != nil {
		log.Printf("Warning: could not check database state: %v", err)
		return
	}
	if !empty {
		log.Println("Index already populated, skipping reindex")
		return
	}

	log.Println("Auto-indexing ADRs...")
	p := parser.NewMarkdownParser()
	r := &reindex.Reindexer{Parser: p, Embedder: emb, Store: st}

	backoff := time.Second
	maxWait := 30 * time.Second
	waited := time.Duration(0)

	for {
		result, err := r.Run(ctx, adrDir)
		if err == nil {
			log.Printf("Auto-index complete: %d ADRs, %d chunks", result.ADRCount, result.ChunkCount)
			return
		}

		waited += backoff
		if waited > maxWait {
			log.Printf("Warning: auto-reindex failed after %v: %v", maxWait, err)
			log.Println("Server starting without index — ADR browsing available, queries may fail")
			return
		}

		log.Printf("Auto-reindex failed (retrying in %v): %v", backoff, err)
		time.Sleep(backoff)
		backoff *= 2
	}
}
