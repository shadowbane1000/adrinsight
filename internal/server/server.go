package server

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/shadowbane1000/adrinsight/internal/parser"
	"github.com/shadowbane1000/adrinsight/internal/rag"
	"github.com/shadowbane1000/adrinsight/internal/store"
	"github.com/shadowbane1000/adrinsight/web"
)

// Server holds the dependencies for the HTTP API.
type Server struct {
	Pipeline *rag.Pipeline
	Store    store.Store
	Parser   parser.Parser
	Port     int
	DevMode  bool
}

// NewServeMux creates and configures the HTTP routes.
func (s *Server) NewServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /query", s.handleQuery)
	mux.HandleFunc("GET /adrs", s.handleListADRs)
	mux.HandleFunc("GET /adrs/{number}", s.handleGetADR)

	// Static file serving: embedded FS in production, disk in dev mode.
	if s.DevMode {
		log.Println("Dev mode: serving static files from web/static/")
		mux.Handle("/", http.FileServer(http.Dir("web/static")))
	} else {
		staticFS, err := fs.Sub(web.StaticFS, "static")
		if err != nil {
			log.Fatalf("Failed to create sub-filesystem: %v", err)
		}
		mux.Handle("/", http.FileServer(http.FS(staticFS)))
	}

	return mux
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	addr := fmt.Sprintf(":%d", s.Port)
	log.Printf("Starting server on %s", addr)
	return http.ListenAndServe(addr, s.NewServeMux())
}
