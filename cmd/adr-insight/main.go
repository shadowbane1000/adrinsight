package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/shadowbane1000/adrinsight/internal/embedder"
	"github.com/shadowbane1000/adrinsight/internal/eval"
	"github.com/shadowbane1000/adrinsight/internal/llm"
	"github.com/shadowbane1000/adrinsight/internal/parser"
	"github.com/shadowbane1000/adrinsight/internal/rag"
	"github.com/shadowbane1000/adrinsight/internal/reindex"
	"github.com/shadowbane1000/adrinsight/internal/server"
	"github.com/shadowbane1000/adrinsight/internal/store"
)

const queryPrefix = "search_query: "

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: adr-insight <command> [flags]\n\nCommands:\n  reindex  Parse, embed, and store ADRs\n  search   Search indexed ADRs by similarity\n  serve    Start the HTTP API server\n  eval     Evaluate answer quality against test cases\n")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "reindex":
		cmdReindex(os.Args[2:])
	case "search":
		cmdSearch(os.Args[2:])
	case "serve":
		cmdServe(os.Args[2:])
	case "eval":
		cmdEval(os.Args[2:])
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
	chunkStrategy := fs.String("chunk-strategy", "sections", "Chunking strategy: sections, wholedoc, preamble")
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

	// Use Anthropic for keyword extraction and relationship classification if API key is available.
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		l := llm.NewAnthropicLLM(apiKey, "")
		r.Keywords = l
		r.RelClassifier = l
	}

	switch *chunkStrategy {
	case "sections":
		// default — uses Parser.ChunkADR
	case "wholedoc":
		r.ChunkFn = p.ChunkADRWithWholeDoc
	case "preamble":
		r.ChunkFn = p.ChunkADRWithPreamble
	default:
		log.Fatalf("Unknown chunk strategy: %s (use sections, wholedoc, or preamble)", *chunkStrategy)
	}
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

	results, err := st.HybridSearch(ctx, vecs[0], query, *topK, 0.6, 0.4)
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
			TopK:     5,
			Reranker: &rag.DefaultReranker{},
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

func cmdEval(args []string) {
	fs := flag.NewFlagSet("eval", flag.ExitOnError)
	casesPath := fs.String("cases", "./testdata/eval/cases.json", "Path to test case corpus")
	baselinePath := fs.String("baseline", "./testdata/eval/baseline.json", "Path to baseline file")
	saveBaseline := fs.Bool("save-baseline", false, "Save this run's results as the new baseline")
	outputPath := fs.String("output", "", "Write full results JSON to this file")
	skipJudge := fs.Bool("skip-judge", false, "Skip LLM judge scoring (retrieval metrics only)")
	delta := fs.Float64("delta", 0.2, "Maximum allowed per-question score drop (0.0-1.0)")
	dbPath := fs.String("db", "./adr-insight.db", "Path to SQLite database")
	ollamaURL := fs.String("ollama-url", "http://localhost:11434", "Ollama API base URL")
	adrDir := fs.String("adr-dir", "./docs/adr", "ADR directory for full content reads")
	model := fs.String("model", "claude-sonnet-4-5", "Anthropic model for judge")
	_ = fs.Parse(args)

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Println("Evaluation skipped: ANTHROPIC_API_KEY not set")
		os.Exit(0)
	}

	// Load test cases.
	cases, err := eval.LoadTestCases(*casesPath)
	if err != nil {
		log.Fatalf("Failed to load test cases: %v", err)
	}

	// Create embedder — test connectivity.
	emb, err := embedder.NewOllamaEmbedder(*ollamaURL, "nomic-embed-text")
	if err != nil {
		log.Printf("Evaluation skipped: embedding service unavailable: %v", err)
		os.Exit(0)
	}

	// Test Ollama connectivity with a trivial embed.
	ctx := context.Background()
	_, err = emb.Embed(ctx, []string{"connectivity test"})
	if err != nil {
		log.Printf("Evaluation skipped: embedding service unavailable: %v", err)
		os.Exit(0)
	}

	st, err := store.NewSQLiteStore(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = st.Close() }()

	l := llm.NewAnthropicLLM(apiKey, *model)
	pipeline := &rag.Pipeline{
		Embedder: emb,
		Store:    st,
		LLM:     l,
		ADRDir:   *adrDir,
		TopK:     5,
		Reranker: &rag.DefaultReranker{},
	}

	// Create judge (nil if --skip-judge).
	var judge eval.Judge
	if !*skipJudge && apiKey != "" {
		judge = eval.NewAnthropicJudge(apiKey, *model)
	}
	if *skipJudge {
		log.Println("LLM judge scoring skipped (--skip-judge)")
	}

	// Run evaluation.
	report, err := eval.RunEval(ctx, cases, pipeline, judge, *adrDir)
	if err != nil {
		log.Fatalf("Evaluation failed: %v", err)
	}

	// Load baseline for comparison.
	baseline, err := eval.LoadBaseline(*baselinePath)
	if err != nil {
		log.Fatalf("Failed to load baseline: %v", err)
	}

	// Detect regressions.
	if baseline != nil {
		eval.DetectRegressions(report, baseline, *delta)
	}

	// Print report.
	eval.PrintReport(os.Stdout, report, cases, baseline)

	// Write full results JSON if requested.
	if *outputPath != "" {
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal results: %v", err)
		}
		if err := os.WriteFile(*outputPath, data, 0644); err != nil {
			log.Fatalf("Failed to write results: %v", err)
		}
		fmt.Printf("\nFull results written to %s\n", *outputPath)
	}

	// Save baseline if requested.
	if *saveBaseline {
		if err := eval.SaveBaseline(*baselinePath, report, *delta); err != nil {
			log.Fatalf("Failed to save baseline: %v", err)
		}
		fmt.Printf("\nBaseline saved to %s\n", *baselinePath)
	}

	// Exit code based on regressions.
	if len(report.Regressions) > 0 {
		os.Exit(1)
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
