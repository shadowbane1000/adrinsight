package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type queryRequest struct {
	Query string `json:"query"`
}

type queryResponse struct {
	Answer    string     `json:"answer"`
	Citations []citation `json:"citations"`
}

type citation struct {
	ADRNumber int    `json:"adr_number"`
	Title     string `json:"title"`
	Section   string `json:"section"`
}

type adrListResponse struct {
	ADRs []adrSummaryJSON `json:"adrs"`
}

type adrSummaryJSON struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Status string `json:"status"`
	Date   string `json:"date,omitempty"`
	Path   string `json:"path"`
}

type adrDetailResponse struct {
	Number        int               `json:"number"`
	Title         string            `json:"title"`
	Status        string            `json:"status"`
	Date          string            `json:"date,omitempty"`
	Content       string            `json:"content"`
	Relationships []relationshipJSON `json:"relationships,omitempty"`
}

type relationshipJSON struct {
	TargetADR   int    `json:"target_adr"`
	TargetTitle string `json:"target_title"`
	RelType     string `json:"rel_type"`
	Description string `json:"description"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var req queryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	if strings.TrimSpace(req.Query) == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "query is required"})
		return
	}

	if s.MaxQueryLength > 0 && len(req.Query) > s.MaxQueryLength {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: fmt.Sprintf("query too long: %d characters (maximum %d)", len(req.Query), s.MaxQueryLength),
		})
		return
	}

	if s.Pipeline == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "ANTHROPIC_API_KEY not configured"})
		return
	}

	resp, err := s.Pipeline.Query(r.Context(), req.Query)
	if err != nil {
		slog.Error("query failed", "error", err, "request_id", RequestID(r.Context()))
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to process query"})
		return
	}

	out := queryResponse{Answer: resp.Answer}
	for _, c := range resp.Citations {
		out.Citations = append(out.Citations, citation{
			ADRNumber: c.ADRNumber,
			Title:     c.Title,
			Section:   c.Section,
		})
	}
	if out.Citations == nil {
		out.Citations = []citation{}
	}
	slog.Info("query processed", "query", req.Query, "request_id", RequestID(r.Context()))
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleListADRs(w http.ResponseWriter, r *http.Request) {
	adrs, err := s.Store.ListADRs(r.Context())
	if err != nil {
		slog.Error("listing ADRs failed", "error", err, "request_id", RequestID(r.Context()))
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to list ADRs"})
		return
	}

	out := adrListResponse{ADRs: make([]adrSummaryJSON, len(adrs))}
	for i, a := range adrs {
		date := ""
		if content, err := os.ReadFile(a.Path); err == nil {
			date = extractDate(string(content))
		}
		out.ADRs[i] = adrSummaryJSON{
			Number: a.Number,
			Title:  a.Title,
			Status: a.Status,
			Date:   date,
			Path:   a.Path,
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleGetADR(w http.ResponseWriter, r *http.Request) {
	numStr := r.PathValue("number")
	num, err := strconv.Atoi(numStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid ADR number"})
		return
	}

	adrs, err := s.Store.ListADRs(r.Context())
	if err != nil {
		slog.Error("listing ADRs failed", "error", err, "request_id", RequestID(r.Context()))
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to list ADRs"})
		return
	}

	var found *adrSummaryJSON
	for _, a := range adrs {
		if a.Number == num {
			found = &adrSummaryJSON{Number: a.Number, Title: a.Title, Status: a.Status, Path: a.Path}
			break
		}
	}
	if found == nil {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "ADR " + numStr + " not found"})
		return
	}

	// Validate path is within ADR directory to prevent traversal.
	absPath, _ := filepath.Abs(found.Path)
	absADRDir, _ := filepath.Abs(s.ADRDir)
	if !strings.HasPrefix(absPath, absADRDir+string(filepath.Separator)) {
		slog.Warn("ADR path outside ADR directory", "path", found.Path, "adr_dir", s.ADRDir, "request_id", RequestID(r.Context()))
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "ADR " + numStr + " not found"})
		return
	}

	content, err := os.ReadFile(found.Path)
	if err != nil {
		slog.Error("reading ADR file failed", "path", found.Path, "error", err, "request_id", RequestID(r.Context()))
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "could not read ADR file"})
		return
	}

	// Extract date from file content.
	date := extractDate(string(content))

	// Build title lookup for relationship targets.
	titleMap := make(map[int]string, len(adrs))
	for _, a := range adrs {
		titleMap[a.Number] = a.Title
	}

	// Load relationships for this ADR.
	var relJSON []relationshipJSON
	rels, err := s.Store.GetRelationships(r.Context(), num)
	if err == nil {
		for _, rel := range rels {
			// Present relationship from this ADR's perspective.
			target := rel.TargetADR
			if target == num {
				target = rel.SourceADR
			}
			relJSON = append(relJSON, relationshipJSON{
				TargetADR:   target,
				TargetTitle: titleMap[target],
				RelType:     rel.RelType,
				Description: rel.Description,
			})
		}
	}

	writeJSON(w, http.StatusOK, adrDetailResponse{
		Number:        found.Number,
		Title:         found.Title,
		Status:        found.Status,
		Date:          date,
		Content:       string(content),
		Relationships: relJSON,
	})
}

func extractDate(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		line = strings.ReplaceAll(line, "**", "")
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Date:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Date:"))
		}
	}
	return ""
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	type componentStatus struct {
		Status string `json:"status"`
		Error  string `json:"error,omitempty"`
	}
	type healthResponse struct {
		Status     string                     `json:"status"`
		Components map[string]componentStatus `json:"components"`
	}

	components := make(map[string]componentStatus)

	// Check database
	dbStatus := componentStatus{Status: "healthy"}
	if _, err := s.Store.IsEmpty(r.Context()); err != nil {
		dbStatus = componentStatus{Status: "unhealthy", Error: err.Error()}
	}
	components["database"] = dbStatus

	// Check Ollama
	ollamaStatus := componentStatus{Status: "healthy"}
	if s.OllamaURL != "" {
		client := &http.Client{Timeout: 3 * time.Second}
		resp, err := client.Head(s.OllamaURL)
		if err != nil {
			ollamaStatus = componentStatus{Status: "unhealthy", Error: err.Error()}
		} else {
			_ = resp.Body.Close()
		}
	}
	components["ollama"] = ollamaStatus

	// Check Anthropic key
	keyStatus := componentStatus{Status: "healthy"}
	if s.Pipeline == nil {
		keyStatus = componentStatus{Status: "unhealthy", Error: "ANTHROPIC_API_KEY not set"}
	}
	components["anthropic_key"] = keyStatus

	// Determine overall status
	overall := "healthy"
	for name, c := range components {
		if c.Status == "unhealthy" {
			if name == "database" {
				overall = "unhealthy"
				break
			}
			if overall != "unhealthy" {
				overall = "degraded"
			}
		}
	}

	code := http.StatusOK
	if overall == "unhealthy" {
		code = http.StatusServiceUnavailable
	}

	writeJSON(w, code, healthResponse{Status: overall, Components: components})
}

func (s *Server) handleAbout(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(s.ADRDir, "about.html")
	if _, err := os.Stat(path); err != nil {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, path)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
