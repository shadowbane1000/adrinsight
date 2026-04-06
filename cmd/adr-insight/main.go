package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tylerc-atx/adr-insight/internal/embedder"
	"github.com/tylerc-atx/adr-insight/internal/parser"
	"github.com/tylerc-atx/adr-insight/internal/reindex"
	"github.com/tylerc-atx/adr-insight/internal/store"
)

const queryPrefix = "search_query: "

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: adr-insight <command> [flags]\n\nCommands:\n  reindex  Parse, embed, and store ADRs\n  search   Search indexed ADRs by similarity\n")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "reindex":
		cmdReindex(os.Args[2:])
	case "search":
		cmdSearch(os.Args[2:])
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
		fmt.Printf("%d. [ADR-%03d] %s — %s (score: %.2f)\n   %s\n\n",
			i+1, r.ADRNumber, r.ADRTitle, r.Section, r.Score, preview)
	}
}
