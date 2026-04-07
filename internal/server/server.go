package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/tylerc-atx/adr-insight/internal/parser"
	"github.com/tylerc-atx/adr-insight/internal/rag"
	"github.com/tylerc-atx/adr-insight/internal/store"
)

// Server holds the dependencies for the HTTP API.
type Server struct {
	Pipeline *rag.Pipeline
	Store    store.Store
	Parser   parser.Parser
	Port     int
}

// NewServeMux creates and configures the HTTP routes.
func (s *Server) NewServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /query", s.handleQuery)
	mux.HandleFunc("GET /adrs", s.handleListADRs)
	mux.HandleFunc("GET /adrs/{number}", s.handleGetADR)
	return mux
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	addr := fmt.Sprintf(":%d", s.Port)
	log.Printf("Starting server on %s", addr)
	return http.ListenAndServe(addr, s.NewServeMux())
}
