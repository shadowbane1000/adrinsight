package llm

import "context"

// ADRContext holds the full content of an ADR for synthesis.
type ADRContext struct {
	Number  int
	Title   string
	Content string
}

// Citation references a specific ADR section used in a synthesized answer.
type Citation struct {
	ADRNumber int    `json:"adr_number"`
	Title     string `json:"title"`
	Section   string `json:"section"`
}

// QueryResponse holds the synthesized answer and its citations.
type QueryResponse struct {
	Answer        string     `json:"answer"`
	Citations     []Citation `json:"citations"`
	RetrievedADRs []int      `json:"retrieved_adrs,omitempty"` // ADR numbers from retrieval (before synthesis)
}

// LLM synthesizes answers from ADR context.
type LLM interface {
	Synthesize(ctx context.Context, query string, adrContents []ADRContext) (QueryResponse, error)
}
