package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/shadowbane1000/adrinsight/internal/config"
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
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}
	config.SetupLogger(cfg)

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: adr-insight <command> [flags]\n\nCommands:\n  reindex  Parse, embed, and store ADRs\n  search   Search indexed ADRs by similarity\n  serve    Start the HTTP API server\n  eval     Evaluate answer quality against test cases\n")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "reindex":
		cmdReindex(os.Args[2:], cfg)
	case "search":
		cmdSearch(os.Args[2:], cfg)
	case "serve":
		cmdServe(os.Args[2:], cfg)
	case "eval":
		cmdEval(os.Args[2:], cfg)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func cmdReindex(args []string, cfg *config.Config) {
	fs := flag.NewFlagSet("reindex", flag.ExitOnError)
	adrDir := fs.String("adr-dir", cfg.ADRDir, "Directory containing ADR files")
	dbPath := fs.String("db", cfg.DBPath, "Path to SQLite database")
	ollamaURL := fs.String("ollama-url", cfg.OllamaURL, "Ollama API base URL")
	chunkStrategy := fs.String("chunk-strategy", "sections", "Chunking strategy: sections, wholedoc, preamble")
	_ = fs.Parse(args)

	ctx := context.Background()

	p := parser.NewMarkdownParser()

	emb, err := embedder.NewOllamaEmbedder(*ollamaURL, cfg.EmbedModel)
	if err != nil {
		slog.Error("failed to create embedder", "error", err)
		os.Exit(1)
	}

	st, err := store.NewSQLiteStore(*dbPath)
	if err != nil {
		slog.Error("failed to open database", "path", *dbPath, "error", err)
		os.Exit(1)
	}
	defer func() { _ = st.Close() }()

	r := &reindex.Reindexer{Parser: p, Embedder: emb, Store: st}

	if cfg.AnthropicKey != "" {
		l := llm.NewAnthropicLLM(cfg.AnthropicKey, "")
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
		slog.Error("unknown chunk strategy", "strategy", *chunkStrategy)
		os.Exit(1)
	}

	result, err := r.Run(ctx, *adrDir)
	if err != nil {
		slog.Error("reindex failed", "error", err)
		os.Exit(1)
	}

	slog.Info("reindex complete", "adrs", result.ADRCount, "chunks", result.ChunkCount)
}

func cmdSearch(args []string, cfg *config.Config) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	dbPath := fs.String("db", cfg.DBPath, "Path to SQLite database")
	ollamaURL := fs.String("ollama-url", cfg.OllamaURL, "Ollama API base URL")
	topK := fs.Int("top-k", 5, "Number of results to return")
	_ = fs.Parse(args)

	query := strings.Join(fs.Args(), " ")
	if query == "" {
		fmt.Fprintln(os.Stderr, "Usage: adr-insight search [flags] <query>")
		os.Exit(1)
	}

	ctx := context.Background()

	emb, err := embedder.NewOllamaEmbedder(*ollamaURL, cfg.EmbedModel)
	if err != nil {
		slog.Error("failed to create embedder", "error", err)
		os.Exit(1)
	}

	st, err := store.NewSQLiteStore(*dbPath)
	if err != nil {
		slog.Error("failed to open database", "path", *dbPath, "error", err)
		os.Exit(1)
	}
	defer func() { _ = st.Close() }()

	vecs, err := emb.Embed(ctx, []string{queryPrefix + query})
	if err != nil {
		slog.Error("failed to embed query", "error", err)
		os.Exit(1)
	}
	if len(vecs) == 0 {
		slog.Error("no embedding returned for query")
		os.Exit(1)
	}

	results, err := st.HybridSearch(ctx, vecs[0], query, *topK, 0.6, 0.4)
	if err != nil {
		slog.Error("search failed", "error", err)
		os.Exit(1)
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

func cmdServe(args []string, cfg *config.Config) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.Int("port", cfg.Port, "HTTP server port")
	dbPath := fs.String("db", cfg.DBPath, "Path to SQLite database")
	ollamaURL := fs.String("ollama-url", cfg.OllamaURL, "Ollama API base URL")
	adrDir := fs.String("adr-dir", cfg.ADRDir, "ADR directory for full content reads")
	model := fs.String("model", "claude-sonnet-4-5", "Anthropic model to use")
	devMode := fs.Bool("dev", false, "Serve static files from disk for live editing")
	_ = fs.Parse(args)

	emb, err := embedder.NewOllamaEmbedder(*ollamaURL, cfg.EmbedModel)
	if err != nil {
		slog.Error("failed to create embedder", "error", err)
		os.Exit(1)
	}

	st, err := store.NewSQLiteStore(*dbPath)
	if err != nil {
		slog.Error("failed to open database", "path", *dbPath, "error", err)
		os.Exit(1)
	}
	defer func() { _ = st.Close() }()

	autoReindex(context.Background(), st, emb, *adrDir)

	// Warm up Ollama by loading the embedding model before serving requests.
	go func() {
		slog.Info("warming up embedding model")
		if _, err := emb.Embed(context.Background(), []string{"warmup"}); err != nil {
			slog.Warn("embedding model warmup failed", "error", err)
		} else {
			slog.Info("embedding model ready")
		}
	}()

	var pipeline *rag.Pipeline
	if cfg.AnthropicKey == "" {
		slog.Warn("ANTHROPIC_API_KEY not set — query endpoint disabled, ADR browsing available")
	} else {
		l := llm.NewAnthropicLLM(cfg.AnthropicKey, *model)
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
		Pipeline:             pipeline,
		Store:                st,
		Parser:               parser.NewMarkdownParser(),
		Port:                 *port,
		DevMode:              *devMode,
		ADRDir:               *adrDir,
		OllamaURL:            *ollamaURL,
		SlowRequestThreshold: cfg.SlowRequestThreshold,
	}

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	httpServer := srv.NewHTTPServer()

	go func() {
		slog.Info("starting server", "port", *port, "dev_mode", *devMode)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	stop()
	slog.Info("shutting down", "timeout", cfg.ShutdownTimeout)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	start := time.Now()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
	slog.Info("shutdown complete", "duration_ms", time.Since(start).Milliseconds())
}

func cmdEval(args []string, cfg *config.Config) {
	fs := flag.NewFlagSet("eval", flag.ExitOnError)
	casesPath := fs.String("cases", "./testdata/eval/cases.json", "Path to test case corpus")
	baselinePath := fs.String("baseline", "./testdata/eval/baseline.json", "Path to baseline file")
	saveBaseline := fs.Bool("save-baseline", false, "Save this run's results as the new baseline")
	outputPath := fs.String("output", "", "Write full results JSON to this file")
	skipJudge := fs.Bool("skip-judge", false, "Skip LLM judge scoring (retrieval metrics only)")
	delta := fs.Float64("delta", 0.2, "Maximum allowed per-question score drop (0.0-1.0)")
	dbPath := fs.String("db", cfg.DBPath, "Path to SQLite database")
	ollamaURL := fs.String("ollama-url", cfg.OllamaURL, "Ollama API base URL")
	adrDir := fs.String("adr-dir", cfg.ADRDir, "ADR directory for full content reads")
	model := fs.String("model", "claude-sonnet-4-5", "Anthropic model for judge")
	_ = fs.Parse(args)

	if cfg.AnthropicKey == "" {
		slog.Info("evaluation skipped: ANTHROPIC_API_KEY not set")
		os.Exit(0)
	}

	cases, err := eval.LoadTestCases(*casesPath)
	if err != nil {
		slog.Error("failed to load test cases", "path", *casesPath, "error", err)
		os.Exit(1)
	}

	emb, err := embedder.NewOllamaEmbedder(*ollamaURL, cfg.EmbedModel)
	if err != nil {
		slog.Info("evaluation skipped: embedding service unavailable", "error", err)
		os.Exit(0)
	}

	ctx := context.Background()
	_, err = emb.Embed(ctx, []string{"connectivity test"})
	if err != nil {
		slog.Info("evaluation skipped: embedding service unavailable", "error", err)
		os.Exit(0)
	}

	st, err := store.NewSQLiteStore(*dbPath)
	if err != nil {
		slog.Error("failed to open database", "path", *dbPath, "error", err)
		os.Exit(1)
	}
	defer func() { _ = st.Close() }()

	l := llm.NewAnthropicLLM(cfg.AnthropicKey, *model)
	pipeline := &rag.Pipeline{
		Embedder: emb,
		Store:    st,
		LLM:     l,
		ADRDir:   *adrDir,
		TopK:     5,
		Reranker: &rag.DefaultReranker{},
	}

	var judge eval.Judge
	if !*skipJudge {
		judge = eval.NewAnthropicJudge(cfg.AnthropicKey, *model)
	}
	if *skipJudge {
		slog.Info("LLM judge scoring skipped", "flag", "--skip-judge")
	}

	report, err := eval.RunEval(ctx, cases, pipeline, judge, *adrDir)
	if err != nil {
		slog.Error("evaluation failed", "error", err)
		os.Exit(1)
	}

	baseline, err := eval.LoadBaseline(*baselinePath)
	if err != nil {
		slog.Error("failed to load baseline", "path", *baselinePath, "error", err)
		os.Exit(1)
	}

	if baseline != nil {
		eval.DetectRegressions(report, baseline, *delta)
	}

	eval.PrintReport(os.Stdout, report, cases, baseline)

	if *outputPath != "" {
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			slog.Error("failed to marshal results", "error", err)
			os.Exit(1)
		}
		if err := os.WriteFile(*outputPath, data, 0644); err != nil {
			slog.Error("failed to write results", "path", *outputPath, "error", err)
			os.Exit(1)
		}
		fmt.Printf("\nFull results written to %s\n", *outputPath)
	}

	if *saveBaseline {
		if err := eval.SaveBaseline(*baselinePath, report, *delta); err != nil {
			slog.Error("failed to save baseline", "path", *baselinePath, "error", err)
			os.Exit(1)
		}
		fmt.Printf("\nBaseline saved to %s\n", *baselinePath)
	}

	if len(report.Regressions) > 0 {
		os.Exit(1)
	}
}

func autoReindex(ctx context.Context, st *store.SQLiteStore, emb *embedder.OllamaEmbedder, adrDir string) {
	empty, err := st.IsEmpty(ctx)
	if err != nil {
		slog.Warn("could not check database state", "error", err)
		return
	}
	if !empty {
		slog.Debug("index already populated, skipping reindex")
		return
	}

	slog.Info("auto-indexing ADRs", "adr_dir", adrDir)
	p := parser.NewMarkdownParser()
	r := &reindex.Reindexer{Parser: p, Embedder: emb, Store: st}

	backoff := time.Second
	maxWait := 30 * time.Second
	waited := time.Duration(0)

	for {
		result, err := r.Run(ctx, adrDir)
		if err == nil {
			slog.Info("auto-index complete", "adrs", result.ADRCount, "chunks", result.ChunkCount)
			return
		}

		waited += backoff
		if waited > maxWait {
			slog.Warn("auto-reindex failed, starting without index", "waited", maxWait, "error", err)
			return
		}

		slog.Warn("auto-reindex failed, retrying", "backoff", backoff, "error", err)
		time.Sleep(backoff)
		backoff *= 2
	}
}
