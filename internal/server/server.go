package server

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"github.com/shadowbane1000/adrinsight/internal/parser"
	"github.com/shadowbane1000/adrinsight/internal/rag"
	"github.com/shadowbane1000/adrinsight/internal/store"
	"github.com/shadowbane1000/adrinsight/web"
)

// Server holds the dependencies for the HTTP API.
type Server struct {
	Pipeline             *rag.Pipeline
	Store                store.Store
	Parser               parser.Parser
	Port                 int
	DevMode              bool
	ADRDir               string
	OllamaURL            string
	SlowRequestThreshold time.Duration
	RateLimitRequests    int
	RateLimitWindow      time.Duration
	MaxQueryLength       int
}

// NewServeMux creates and configures the HTTP routes.
func (s *Server) NewServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	rl := newRateLimiter(s.RateLimitRequests, s.RateLimitWindow)
	mux.HandleFunc("POST /query", rateLimitMiddleware(rl, s.handleQuery))
	mux.HandleFunc("GET /adrs", s.handleListADRs)
	mux.HandleFunc("GET /adrs/{number}", s.handleGetADR)
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /about.html", s.handleAbout)

	if s.DevMode {
		slog.Info("serving static files from disk", "path", "web/static/")
		mux.Handle("/", http.FileServer(http.Dir("web/static")))
	} else {
		staticFS, err := fs.Sub(web.StaticFS, "static")
		if err != nil {
			slog.Error("failed to create sub-filesystem", "error", err)
			panic("failed to create sub-filesystem: " + err.Error())
		}
		mux.Handle("/", http.FileServer(http.FS(staticFS)))
	}

	return mux
}

// NewHTTPServer creates an http.Server with middleware applied.
func (s *Server) NewHTTPServer() *http.Server {
	mux := s.NewServeMux()

	// Apply middleware: requestID first, then logging
	handler := requestIDMiddleware(loggingMiddleware(s.SlowRequestThreshold)(mux))

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Port),
		Handler: handler,
	}
}

// ListenAndServe starts the HTTP server (kept for backward compatibility).
func (s *Server) ListenAndServe() error {
	srv := s.NewHTTPServer()
	slog.Info("starting server", "port", s.Port)
	return srv.ListenAndServe()
}
