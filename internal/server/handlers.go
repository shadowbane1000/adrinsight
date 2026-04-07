package server

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	var req queryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	if strings.TrimSpace(req.Query) == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "query is required"})
		return
	}

	if s.Pipeline == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "ANTHROPIC_API_KEY not configured"})
		return
	}

	resp, err := s.Pipeline.Query(r.Context(), req.Query)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
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
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleListADRs(w http.ResponseWriter, r *http.Request) {
	adrs, err := s.Store.ListADRs(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}

	out := adrListResponse{ADRs: make([]adrSummaryJSON, len(adrs))}
	for i, a := range adrs {
		out.ADRs[i] = adrSummaryJSON{
			Number: a.Number,
			Title:  a.Title,
			Status: a.Status,
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
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
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

	content, err := os.ReadFile(found.Path)
	if err != nil {
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

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
